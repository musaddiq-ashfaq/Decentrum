package main

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// User represents the user structure in the application
// User represents the user structure in the application

type User struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Password       string `json:"password"`
	ProfilePicture string `json:"profilePicture,omitempty"`
}

// UserForIPFS represents user data to be stored in IPFS
type UserForIPFS struct {
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	ProfilePicture string `json:"profilePicture,omitempty"`
}

// Post represents a social media post
type Post struct {
	ID            int               `json:"id"`
	User          User              `json:"user"`
	Content       string            `json:"content"`
	Timestamp     time.Time         `json:"timestamp"`
	Reactions     map[string]string `json:"reactions,omitempty"`
	ReactionCount int               `json:"reactionCount"`
}

// Reaction represents a user's reaction to a post
type Reaction struct {
	PostID int    `json:"postId"`
	UserID string `json:"userId"`
	Type   string `json:"type"`
}

// APIError represents a structured API error response
type APIError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// Global variables
var (
	ipfsShell *shell.Shell
	contract  *client.Contract
)

// initFabric initializes the connection to the Fabric network
func initFabric() error {
	clientConnection := newGrpcConnection()
	id := newIdentity()
	sign := newSign()

	// Configure gateway with increased timeouts
	gateway, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		client.WithEvaluateTimeout(6*time.Minute),     // 5 minutes for evaluation
		client.WithEndorseTimeout(6*time.Minute),      // 5 minutes for endorsement
		client.WithSubmitTimeout(6*time.Minute),       // 5 minutes for submission
		client.WithCommitStatusTimeout(3*time.Minute), // 2 minutes for commit status
	)
	if err != nil {
		return fmt.Errorf("failed to connect to gateway: %w", err)
	}

	network := gateway.GetNetwork("mychannel")
	contract = network.GetContract("social_media")
	log.Println("Successfully connected to Fabric network")
	return nil
}

// newGrpcConnection creates a new gRPC connection with optimized settings
func newGrpcConnection() *grpc.ClientConn {
	certificate, err := loadCertificate()
	if err != nil {
		log.Fatalf("Failed to load certificate: %v", err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, "peer0.org1.example.com")

	// Configure keepalive options
	kaOpts := keepalive.ClientParameters{
		Time:                10 * time.Second, // Send pings every 10 seconds
		Timeout:             30 * time.Second, // Wait 30 seconds for ping ack
		PermitWithoutStream: true,             // Send pings even without active streams
	}

	// Create connection with optimized settings
	connection, err := grpc.Dial(
		"localhost:7051",
		grpc.WithTransportCredentials(transportCredentials),
		grpc.WithKeepaliveParams(kaOpts),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(20*1024*1024), // 20MB max receive message size
			grpc.MaxCallSendMsgSize(20*1024*1024), // 20MB max send message size
		),
	)
	if err != nil {
		log.Fatalf("Failed to create gRPC connection: %v", err)
	}

	return connection
}

// loadCertificate loads the certificate from the filesystem
func loadCertificate() (*x509.Certificate, error) {
	pemPath := filepath.Join(
		"..", "..", "fabric-samples", "test-network", "organizations",
		"peerOrganizations", "org1.example.com", "peers",
		"peer0.org1.example.com", "tls", "ca.crt",
	)

	certificatePEM, err := os.ReadFile(pemPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	return identity.CertificateFromPEM(certificatePEM)
}

// newIdentity creates a new X509 identity
func newIdentity() *identity.X509Identity {
	certificatePath := filepath.Join(
		"..", "..", "fabric-samples", "test-network", "organizations",
		"peerOrganizations", "org1.example.com", "users",
		"User1@org1.example.com", "msp", "signcerts", "cert.pem",
	)

	certificatePEM, err := os.ReadFile(certificatePath)
	if err != nil {
		log.Fatalf("Failed to read certificate file: %v", err)
	}

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		log.Fatalf("Failed to create certificate from PEM: %v", err)
	}

	id, err := identity.NewX509Identity("Org1MSP", certificate)
	if err != nil {
		log.Fatalf("Failed to create X509 identity: %v", err)
	}

	return id
}

// newSign creates a new signing function
func newSign() identity.Sign {
	keyPath := filepath.Join(
		"..", "..", "fabric-samples", "test-network", "organizations",
		"peerOrganizations", "org1.example.com", "users",
		"User1@org1.example.com", "msp", "keystore",
	)

	files, err := os.ReadDir(keyPath)
	if err != nil || len(files) == 0 {
		log.Fatalf("Failed to read keystore directory or no keys found: %v", err)
	}

	privateKeyPath := filepath.Join(keyPath, files[0].Name())
	privateKeyPEM, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatalf("Failed to read private key file: %v", err)
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		log.Fatalf("Failed to create private key: %v", err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		log.Fatalf("Failed to create signing function: %v", err)
	}

	return sign
}

// submitWithRetry attempts to submit a transaction with retry logic
func submitWithRetry(function string, args ...string) ([]byte, error) {
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		result, err := contract.Submit(
			function,
			client.WithArguments(args...),
		)
		if err == nil {
			return result, nil
		}

		lastErr = err
		log.Printf("Attempt %d failed: %v", attempt, err)

		if attempt < maxRetries {
			backoffDuration := time.Duration(attempt) * time.Second
			log.Printf("Waiting %v before retry...", backoffDuration)
			time.Sleep(backoffDuration)
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %v", maxRetries, lastErr)
}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if user.Email == "" || user.Password == "" || user.Name == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create UserForIPFS struct excluding sensitive data
	userForIPFS := UserForIPFS{
		Email:          user.Email,
		Phone:          user.Phone,
		ProfilePicture: user.ProfilePicture,
	}

	// Convert user data to JSON for IPFS storage
	userJSON, err := json.Marshal(userForIPFS)
	if err != nil {
		log.Printf("Failed to marshal user data for IPFS: %v", err)
		http.Error(w, "Failed to process user data", http.StatusInternalServerError)
		return
	}

	// Store user data in IPFS
	ipfsHash, err := ipfsShell.Add(strings.NewReader(string(userJSON)))
	if err != nil {
		log.Printf("Failed to store user data in IPFS: %v", err)
		http.Error(w, "Failed to store user data", http.StatusInternalServerError)
		return
	}

	log.Printf("User data stored in IPFS with hash: %s", ipfsHash)

	// Hash password for blockchain storage
	hashedPassword := sha256.New()
	hashedPassword.Write([]byte(user.Password))
	hashedPasswordHex := hex.EncodeToString(hashedPassword.Sum(nil))
	log.Printf("Pass: %v", hashedPasswordHex)
	// Submit to blockchain with retry logic - now including IPFS hash
	// Arguments: email, hashedPassword, name, ipfsHash
	result, err := submitWithRetry(
		"CreateUser",
		user.Email,
		hashedPasswordHex,
		user.Name,
		ipfsHash,
	)
	if err != nil {
		// Try to remove the IPFS content if blockchain storage fails
		if unPinErr := ipfsShell.Unpin(ipfsHash); unPinErr != nil {
			log.Printf("Warning: Failed to unpin IPFS content after blockchain error: %v", unPinErr)
		}
		log.Printf("Failed to store user in blockchain: %v", err)
		http.Error(w, fmt.Sprintf("Failed to register user: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("User registered successfully. Blockchain response: %s", result)

	// Return success response with both blockchain and IPFS information
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "User registered successfully",
		"email":    user.Email,
		"name":     user.Name,
		"ipfsHash": ipfsHash,
	})
}

// Helper function to retrieve user data from IPFS

func getUserFromIPFS(ipfsHash string) (*UserForIPFS, error) {
	// Get the data from IPFS
	reader, err := ipfsShell.Cat(ipfsHash)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve from IPFS: %v", err)
	}
	defer reader.Close()

	// Read the data
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read IPFS data: %v", err)
	}

	// Unmarshal the JSON data
	var user UserForIPFS
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %v", err)
	}

	return &user, nil
}

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if user.Email == "" || user.Password == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Hash the password to compare with stored hash
	hashedPassword := sha256.New()
	hashedPassword.Write([]byte(user.Password))
	hashedPasswordHex := hex.EncodeToString(hashedPassword.Sum(nil))
	log.Printf("Hashed Password: %v", hashedPasswordHex)

	// Submit to blockchain to verify user credentials
	isValid, err := verifyUserCredentials(user.Email, hashedPasswordHex)
	if err != nil {
		log.Printf("Login failed: %v", err)
		http.Error(w, fmt.Sprintf("Login failed: %v", err), http.StatusUnauthorized)
		return
	}

	if !isValid {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	// Return success response with user information
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
	})
}

// verifyUserCredentials verifies user credentials on the blockchain
func verifyUserCredentials(email, hashedPassword string) (bool, error) {
	result, err := contract.EvaluateTransaction("UserExists", email, hashedPassword)
	if err != nil {
		return false, fmt.Errorf("error evaluating transaction: %v", err)
	}

	var isValid bool
	if err := json.Unmarshal(result, &isValid); err != nil {
		return false, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return isValid, nil
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var post Post
		if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
			http.Error(w, "Invalid input. Please check your data.", http.StatusBadRequest)
			log.Printf("Error decoding request body: %v", err)
			return
		}

		// Validate user information
		if post.User.Email == "" || post.Content == "" {
			http.Error(w, "User email and post content are required.", http.StatusBadRequest)
			log.Println("Missing user email or post content.")
			return
		}

		// Check if the user exists on the blockchain (optional validation)
		isValid, err := verifyUserExists(post.User.Email)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to verify user: %v", err), http.StatusInternalServerError)
			log.Printf("Error verifying user existence: %v", err)
			return
		}

		if !isValid {
			http.Error(w, "User does not exist.", http.StatusUnauthorized)
			log.Printf("Unauthorized attempt to create post for non-existing user: %s", post.User.Email)
			return
		}

		// Generate a unique post ID
		post.ID = int(time.Now().Unix())
		post.Timestamp = time.Now()

		// Convert post to JSON
		postJSON, _ := json.Marshal(post)

		// Store post content in IPFS
		ipfsHash, err := ipfsShell.Add(strings.NewReader(string(postJSON)))
		if err != nil {
			http.Error(w, "Failed to store post in IPFS. Please try again later.", http.StatusInternalServerError)
			log.Printf("IPFS storage error: %v", err)
			return
		}

		// Submit the post to the blockchain with retry logic
		result, err := submitPostWithRetry(post.User.Email, ipfsHash)
		if err != nil {
			// If blockchain submission fails after all attempts, unpin the IPFS content
			if unPinErr := ipfsShell.Unpin(ipfsHash); unPinErr != nil {
				log.Printf("Warning: Failed to unpin IPFS content after blockchain error: %v", unPinErr)
			}
			log.Printf("Failed to store post in blockchain: %v", err)
			http.Error(w, fmt.Sprintf("Failed to store post in blockchain: %v", err), http.StatusInternalServerError)
			return
		}

		// Successfully created the post
		log.Printf("Post created successfully. Blockchain response: %s", string(result))

		// Return success response with post ID and IPFS hash
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":   "Post created successfully.",
			"postID":    post.ID,
			"userEmail": post.User.Email,
			"ipfsHash":  ipfsHash,
		})

	case http.MethodGet:
		// Existing logic for GET request to fetch posts by user
		email := r.URL.Query().Get("email")
		if email == "" {
			http.Error(w, "Email is required to fetch posts.", http.StatusBadRequest)
			log.Println("Missing email in query params.")
			return
		}

		result, err := contract.EvaluateTransaction("GetPostsByUser", email)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to fetch posts: %v", err), http.StatusInternalServerError)
			log.Printf("Blockchain query error for posts by user: %v", err)
			return
		}

		var postHashes []string
		if err := json.Unmarshal(result, &postHashes); err != nil {
			http.Error(w, "Failed to parse post data.", http.StatusInternalServerError)
			log.Printf("Error unmarshalling post hashes: %v", err)
			return
		}

		var posts []Post
		for _, hash := range postHashes {
			post, err := getPostFromIPFS(hash)
			if err != nil {
				log.Printf("Failed to fetch post from IPFS: %v", err)
				continue
			}
			posts = append(posts, *post)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(posts)
	}
}

func submitPostWithRetry(email string, ipfsHash string) ([]byte, error) {
	maxRetries := 4
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Create a new transaction proposal
		transaction, err := contract.NewProposal(
			"CreatePost",
			client.WithArguments(email, ipfsHash),
		)
		if err != nil {
			log.Printf("Failed to create transaction proposal: %v", err)
			continue
		}

		// Endorse transaction
		endorsed, err := transaction.Endorse()
		if err != nil {
			log.Printf("Transaction endorsement failed: %v", err)
			if attempt < maxRetries {
				backoffDuration := time.Duration(attempt) * time.Second
				log.Printf("Waiting %v before retrying...", backoffDuration)
				time.Sleep(backoffDuration)
			}
			lastErr = err
			continue
		}

		// Submit endorsed transaction
		result, err := endorsed.Submit()
		if err == nil {
			// Assuming the commit has some byte data to return (adjust based on actual commit object)
			return []byte(result.TransactionID()), nil
		}

		log.Printf("Transaction submission failed: %v", err)
		if attempt < maxRetries {
			backoffDuration := time.Duration(attempt) * time.Second
			log.Printf("Waiting %v before retrying...", backoffDuration)
			time.Sleep(backoffDuration)
		}
		lastErr = err
	}

	return nil, fmt.Errorf("failed to store post after %d attempts: %v", maxRetries, lastErr)
}

// verifyUserExists checks if a user exists in the blockchain by email
func verifyUserExists(email string) (bool, error) {
	result, err := contract.EvaluateTransaction("UserExists", email)
	if err != nil {
		return false, fmt.Errorf("error evaluating transaction: %v", err)
	}

	var isValid bool
	if err := json.Unmarshal(result, &isValid); err != nil {
		return false, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return isValid, nil
}

// Helper function to retrieve a post from IPFS by its hash
func getPostFromIPFS(ipfsHash string) (*Post, error) {
	// Get the data from IPFS
	reader, err := ipfsShell.Cat(ipfsHash)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve from IPFS: %v", err)
	}
	defer reader.Close()

	// Read the data
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read IPFS data: %v", err)
	}

	// Unmarshal the JSON data into a Post struct
	var post Post
	if err := json.Unmarshal(data, &post); err != nil {
		return nil, fmt.Errorf("failed to unmarshal post data: %v", err)
	}

	return &post, nil
}

// main is the entry point of the application
func main() {
	// Initialize IPFS connection
	ipfsShell = shell.NewShell("localhost:5001")
	log.Println("Connected to IPFS")

	// Initialize Fabric connection
	if err := initFabric(); err != nil {
		log.Fatalf("Failed to initialize Fabric gateway: %v", err)
	}

	// Set up router with CORS
	router := mux.NewRouter()
	router.HandleFunc("/signup", SignUpHandler).Methods(http.MethodPost)
	router.HandleFunc("/login", LoginHandler).Methods(http.MethodPost)
	router.HandleFunc("/feed", PostHandler).Methods(http.MethodPost, http.MethodGet)

	// Configure CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	// Start server
	handler := corsHandler.Handler(router)
	server := &http.Server{
		Addr:         ":8081",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("Server starting on :8081")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
