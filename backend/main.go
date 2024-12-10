package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
type User struct {
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	PublicKey string `json:"publicKey"`
}

// Wallet represents a crypto wallet
type Wallet struct {
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
}

// Post represents a social media post
type Post struct {
	ID             int               `json:"id"`
	User           User              `json:"user"`
	Wallet         Wallet            `json:"wallet"`
	Content        string            `json:"content,omitempty"` // Optional text content
	Timestamp      time.Time         `json:"timestamp"`
	Reactions      map[string]string `json:"reactions,omitempty"`
	ReactionCount  int               `json:"reactionCount"`
	ShareCount     int               `json:"shareCount"`
	ImageHash      string            `json:"imageHash,omitempty"` // Add this field for image IPFS hash
	VideoHash      string            `json:"videoHash,omitempty"` // Add this field for video IPFS hash
	IPFSHASH       string            `json:"ipfsHASH,omitempty"`
	ReactionCounts map[string]int    `json:"reactionCounts,omitempty"`
}

// type Post struct {
// 	ID            int               `json:"id"`
// 	User          User              `json:"user"`
// 	Wallet        Wallet            `json:"wallet"`
// 	Content       string            `json:"content,omitempty"`
// 	Timestamp     time.Time         `json:"timestamp"`
// 	Reactions     map[string]string `json:"reactions,omitempty"` // UserPublicKey -> ReactionType
// 	ReactionCount int               `json:"reactionCount"`       // Total reaction count
// 	ShareCount    int               `json:"shareCount"`
// 	ImageHash     string            `json:"imageHash,omitempty"`
// 	VideoHash     string            `json:"videoHash,omitempty"`
// 	IPFSHASH      string            `json:"ipfsHASH,omitempty"`
// }

type ReactionRequest struct {
	UserPublicKey string `json:"userPublicKey"`
	ReactionType  string `json:"reactionType"`
}

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
		client.WithEvaluateTimeout(20*time.Minute),    // Increased timeout for evaluation
		client.WithEndorseTimeout(20*time.Minute),     // Increased timeout for endorsement
		client.WithSubmitTimeout(20*time.Minute),      // Increased timeout for submission
		client.WithCommitStatusTimeout(8*time.Minute), // Increased timeout for commit status
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
		// Time:                10 * time.Second, // Send pings every 10 seconds
		Time:                20 * time.Second,
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

// generateWallet creates a new wallet with a public-private key pair
func generateWallet() (*Wallet, error) {
	priv, pub, err := generateKeys()
	if err != nil {
		return nil, err
	}

	privKeyHex := hex.EncodeToString(priv.D.Bytes())
	pubKeyHex := hex.EncodeToString(pub.X.Bytes()) + hex.EncodeToString(pub.Y.Bytes())

	return &Wallet{
		PublicKey:  pubKeyHex,
		PrivateKey: privKeyHex,
	}, nil
}

func generateKeys() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader) // Use rand.Reader
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate keys: %w", err)
	}
	pub := &priv.PublicKey
	return priv, pub, nil
}

// uploadToIPFS uploads a file to IPFS and returns the IPFS link
func uploadToIPFS(filePath string) (string, error) {
	// Initialize IPFS connection if not already initialized
	if ipfsShell == nil {
		ipfsShell = shell.NewShell("localhost:5001")
	}

	// Verify IPFS shell connection
	if ipfsShell == nil {
		return "", fmt.Errorf("failed to connect to IPFS shell")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash, err := ipfsShell.Add(file)
	if err != nil {
		return "", fmt.Errorf("failed to add file to IPFS: %w", err)
	}

	return "http://localhost:8081/ipfs/" + hash, nil
}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	// Decode the incoming request body into the user struct
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Debug: Print incoming user data for validation
	fmt.Printf("Received user data: %+v\n", user)

	// Generate wallet for the user (Public/Private Key pair)
	wallet, err := generateWallet()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating wallet: %v", err), http.StatusInternalServerError)
		return
	}

	// Debug: Print generated wallet information
	fmt.Printf("Generated wallet: PublicKey = %s, PrivateKey = %s\n", wallet.PublicKey, wallet.PrivateKey)

	// Prepare the user data for the blockchain
	userData := map[string]interface{}{
		"name":       user.Name,
		"phone":      user.Phone,
		"publicKey":  wallet.PublicKey,
		"privateKey": wallet.PrivateKey,
	}

	// Debug: Print the user data to be stored on the blockchain
	fmt.Printf("Storing user data on blockchain: %+v\n", userData)

	// Store the user data in the blockchain using the chaincode "RegisterUser"
	_, err = contract.SubmitTransaction("RegisterUser", user.Name, user.Phone, wallet.PublicKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error registering user on blockchain: %v", err), http.StatusInternalServerError)
		return
	}

	// Verify that the data was stored on the blockchain by querying the data
	// Assuming "GetUser" is a chaincode function to retrieve the user based on their publicKey
	fmt.Println("Sending Public Key:", wallet.PublicKey)
	response, err := contract.EvaluateTransaction("GetUser", wallet.PublicKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching user data from blockchain: %v", err), http.StatusInternalServerError)
		return
	}

	// Debug: Print the retrieved blockchain data for validation
	fmt.Println("Retrieved Blockchain Data:", string(response))

	// Expected user data as a JSON string
	expectedData := fmt.Sprintf("{\"name\":\"%s\",\"phone\":\"%s\",\"publicKey\":\"%s\"}",
		user.Name, user.Phone, wallet.PublicKey)

	// Compare the expected and actual data
	if string(response) == expectedData {
		// Respond with the user data and wallet information
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(userData)
	} else {
		// If there's a mismatch, print the differences
		fmt.Println("Mismatch found between expected and retrieved data:")
		fmt.Println("Expected Data:", expectedData)
		fmt.Println("Actual Data:", string(response))

		// Respond with an error message
		http.Error(w, "User data verification failed on blockchain", http.StatusInternalServerError)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PublicKey  string `json:"publicKey"`
		PrivateKey string `json:"privateKey"`
	}

	// Decode the incoming request body
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Debug: Print received keys
	fmt.Printf("Received Public Key: %s\n", request.PublicKey)
	fmt.Printf("Received Private Key: %s\n", request.PrivateKey)

	// Split the public key into X and Y components
	if len(request.PublicKey) < 128 {
		http.Error(w, "Invalid public key length", http.StatusBadRequest)
		return
	}
	pubXHex := request.PublicKey[:64] // First 64 characters for X
	pubYHex := request.PublicKey[64:] // Last 64 characters for Y

	// Convert public key components to big.Int
	pubX := new(big.Int)
	pubY := new(big.Int)
	pubX.SetString(pubXHex, 16)
	pubY.SetString(pubYHex, 16)

	// Reconstruct the public key
	pubKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     pubX,
		Y:     pubY,
	}

	// Decode the private key from hex
	privKeyInt := new(big.Int)
	privKeyInt.SetString(request.PrivateKey, 16)

	// Reconstruct the private key
	privKey := &ecdsa.PrivateKey{
		PublicKey: *pubKey,
		D:         privKeyInt,
	}

	// Verify if the public key corresponds to the private key
	if privKey.PublicKey.X.Cmp(pubKey.X) != 0 || privKey.PublicKey.Y.Cmp(pubKey.Y) != 0 {
		http.Error(w, "Public key does not match private key", http.StatusUnauthorized)
		return
	}

	// Blockchain Verification: Check if the public key exists in the blockchain
	response, err := contract.EvaluateTransaction("GetUser", request.PublicKey)
	if err != nil {
		http.Error(w, "Error querying blockchain: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if string(response) == "" {
		http.Error(w, "Public key not found in blockchain", http.StatusUnauthorized)
		return
	}

	// If all checks pass, send a success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Optionally respond with user data
	var userData map[string]interface{}
	if err := json.Unmarshal(response, &userData); err != nil {
		http.Error(w, "Error unmarshalling blockchain data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(userData)
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// Parse multipart form data
		err := r.ParseMultipartForm(10 << 20) // 10 MB limit
		if err != nil {
			http.Error(w, "Error parsing multipart form", http.StatusBadRequest)
			log.Printf("Error parsing form: %v", err)
			return
		}

		var post Post
		// Extract user and wallet details
		post.User.Name = r.FormValue("user.name")
		post.Wallet.PublicKey = r.FormValue("wallet.publicKey")
		post.User.PublicKey = post.Wallet.PublicKey

		if post.Wallet.PublicKey == "" {
			http.Error(w, "User public key is required.", http.StatusBadRequest)
			log.Println("Missing user public key.")
			return
		}

		// Extract post content
		post.Content = r.FormValue("content")

		// Validate that at least one of content, photo, or video is provided
		hasPhoto := r.MultipartForm.File["photo"] != nil
		hasVideo := r.MultipartForm.File["video"] != nil

		if post.Content == "" && !hasPhoto && !hasVideo {
			http.Error(w, "At least one of content, photo, or video is required.", http.StatusBadRequest)
			log.Println("No content, photo, or video provided.")
			return
		}

		// Check if the user exists on the blockchain
		isValid, err := verifyUserExists(post.Wallet.PublicKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to verify user: %v", err), http.StatusInternalServerError)
			log.Printf("Error verifying user existence: %v", err)
			return
		}

		if !isValid {
			http.Error(w, "User does not exist.", http.StatusUnauthorized)
			log.Printf("Unauthorized attempt to create post for non-existing user: %s", post.Wallet.PublicKey)
			return
		}

		// Generate a unique post ID and timestamp
		post.ID = int(time.Now().Unix())
		postID := strconv.Itoa(post.ID)
		post.Timestamp = time.Now()

		// Ensure IPFS is initialized
		if ipfsShell == nil {
			ipfsShell = shell.NewShell("localhost:5001")
		}

		// Upload image or video to IPFS if provided
		if hasPhoto {
			file, err := r.MultipartForm.File["photo"][0].Open()
			if err != nil {
				http.Error(w, "Failed to open photo file.", http.StatusInternalServerError)
				log.Printf("Error opening photo file: %v", err)
				return
			}
			defer file.Close()

			ipfsHash, err := ipfsShell.Add(file)
			if err != nil {
				http.Error(w, "Failed to store photo in IPFS. Please try again later.", http.StatusInternalServerError)
				log.Printf("IPFS photo storage error: %v", err)
				return
			}

			post.ImageHash = ipfsHash
		}

		if hasVideo {
			file, err := r.MultipartForm.File["video"][0].Open()
			if err != nil {
				http.Error(w, "Failed to open video file.", http.StatusInternalServerError)
				log.Printf("Error opening video file: %v", err)
				return
			}
			defer file.Close()

			ipfsHash, err := ipfsShell.Add(file)
			if err != nil {
				http.Error(w, "Failed to store video in IPFS. Please try again later.", http.StatusInternalServerError)
				log.Printf("IPFS video storage error: %v", err)
				return
			}

			post.VideoHash = ipfsHash
		}

		// Serialize the Post struct to JSON and upload it to IPFS
		postJSON, err := json.Marshal(post)
		if err != nil {
			http.Error(w, "Failed to marshal post data", http.StatusInternalServerError)
			log.Printf("Failed to marshal post: %v", err)
			return
		}

		ipfsHash, err := ipfsShell.Add(strings.NewReader(string(postJSON)))
		if err != nil {
			http.Error(w, "Failed to store post in IPFS. Please try again later.", http.StatusInternalServerError)
			log.Printf("IPFS storage error: %v", err)
			return
		}

		post.IPFSHASH = ipfsHash
		// Submit the post to the blockchain
		result, err := submitPostWithRetry(post.Wallet.PublicKey, post.IPFSHASH, postID)
		if err != nil {
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
			"publicKey": post.Wallet.PublicKey,
			"ipfsHASH":  post.IPFSHASH,
		})

	case http.MethodGet:
		publicKey := r.URL.Query().Get("publicKey")
		if publicKey == "" {
			http.Error(w, "Public key is required to fetch posts.", http.StatusBadRequest)
			log.Println("Missing public key in query params.")
			return
		}

		result, err := contract.EvaluateTransaction("GetPostsByUser", publicKey)
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

func FeedHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch all post hashes from the blockchain (now using PublicKey)
	result, err := contract.EvaluateTransaction("GetAllPosts")
	if err != nil {
		log.Printf("Error calling GetAllPostsByPublicKey: %v", err)
		http.Error(w, fmt.Sprintf("Failed to fetch posts: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Raw result from GetAllPostsByPublicKey: %s", string(result))

	if len(result) == 0 || string(result) == "null" {
		log.Println("GetAllPostsByPublicKey returned null or empty result")
		json.NewEncoder(w).Encode([]Post{})
		return
	}

	var postHashes []string
	if err := json.Unmarshal(result, &postHashes); err != nil {
		log.Printf("Error unmarshalling post hashes: %v", err)
		http.Error(w, "Failed to parse post data.", http.StatusInternalServerError)
		return
	}

	log.Printf("Unmarshalled %d post hashes", len(postHashes))

	var posts []Post
	for _, hash := range postHashes {
		post, err := getPostFromIPFS(hash)
		if err != nil {
			log.Printf("Failed to fetch post from IPFS: %v", err)
			continue
		}
		posts = append(posts, *post)
	}

	log.Printf("Retrieved %d posts from IPFS", len(posts))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func submitPostWithRetry(publicKey string, ipfsHash string, postID string) ([]byte, error) {
	maxRetries := 4
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Create a new transaction proposal
		transaction, err := contract.NewProposal(
			"CreatePost",
			client.WithArguments(publicKey, ipfsHash, postID),
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

// verifyUserExists checks if a user exists in the blockchain by publicKey
func verifyUserExists(publicKey string) (bool, error) {
	result, err := contract.EvaluateTransaction("UserExists", publicKey)
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

func ReactionHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("ReactionHandler invoked")

	// Ensure the method is POST
	if r.Method != http.MethodPost {
		log.Printf("Invalid HTTP method: %s. Expected POST.", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the post ID from the URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		log.Println("Invalid URL format. Expected: /post/<postID>/react")
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	postID := parts[2]
	log.Printf("Post ID extracted: %s", postID)

	// Parse the request body
	var request ReactionRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Prevent unknown fields in JSON
	if err := decoder.Decode(&request); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Request parsed: UserPublicKey=%s, ReactionType=%s", request.UserPublicKey, request.ReactionType)

	// Validate user public key
	if request.UserPublicKey == "" {
		log.Println("UserPublicKey is missing in the request")
		http.Error(w, "User public key is required", http.StatusBadRequest)
		return
	}

	// Validate reaction type
	validReactions := map[string]bool{
		"like":  true,
		"love":  true,
		"laugh": true,
		"angry": true,
		"sad":   true,
	}
	if !validReactions[request.ReactionType] {
		log.Printf("Invalid ReactionType: %s", request.ReactionType)
		http.Error(w, "Invalid reaction type", http.StatusBadRequest)
		return
	}

	// Retrieve post hash using postID
	postHash, err := getPostHashByID(postID)
	if err != nil {
		log.Printf("Failed to retrieve post hash for PostID=%s: %v", postID, err)
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}
	log.Printf("Post hash retrieved: %s", postHash)

	// Retrieve the post from IPFS
	post, err := getPostFromIPFS(postHash)
	if err != nil {
		log.Printf("Failed to fetch post from IPFS. PostHash=%s, Error=%v", postHash, err)
		http.Error(w, "Failed to retrieve post data", http.StatusInternalServerError)
		return
	}

	// Update the reaction for the user
	if post.Reactions == nil {
		post.Reactions = make(map[string]string)
	}
	oldReaction := post.Reactions[request.UserPublicKey]
	post.Reactions[request.UserPublicKey] = request.ReactionType

	// Recalculate reaction counts
	if post.ReactionCounts == nil {
		post.ReactionCounts = make(map[string]int)
	}
	if oldReaction != "" {
		post.ReactionCounts[oldReaction]--
	}
	post.ReactionCounts[request.ReactionType]++
	post.ReactionCount = len(post.Reactions)

	log.Printf("Updated reactions for post %d: %+v", post.ID, post.ReactionCounts)

	// Serialize the updated post to JSON and store it back in IPFS
	postJSON, err := json.Marshal(post)
	if err != nil {
		log.Printf("Failed to marshal updated post. Error=%v", err)
		http.Error(w, "Failed to update post data", http.StatusInternalServerError)
		return
	}

	updatedIPFSHash, err := ipfsShell.Add(strings.NewReader(string(postJSON)))
	if err != nil {
		log.Printf("Failed to store updated post in IPFS. Error=%v", err)
		http.Error(w, "Failed to store updated post in IPFS", http.StatusInternalServerError)
		return
	}

	log.Printf("Updated post stored in IPFS with hash: %s", updatedIPFSHash)

	// Respond with the updated post
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(post); err != nil {
		log.Printf("Failed to encode updated post response. Error=%v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Println("Reaction successfully added and updated post returned.")
}

func getPostHashByID(postID string) (string, error) {
	// Step 1: Query the blockchain for all posts
	result, err := contract.EvaluateTransaction("GetAllUserPosts")
	if err != nil {
		return "", fmt.Errorf("failed to evaluate transaction: %v", err)
	}

	// Step 2: Parse the result into a map of posts grouped by public key
	var allPosts map[string][]string
	if err := json.Unmarshal(result, &allPosts); err != nil {
		return "", fmt.Errorf("failed to unmarshal result: %v", err)
	}
	log.Printf("All Posts: %+v", allPosts)

	// Convert postID to integer
	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		return "", fmt.Errorf("invalid postID format: %v", err)
	}

	// Step 3: Iterate through all users' posts
	for _, postHashes := range allPosts {
		for _, hash := range postHashes {
			// Call getPostFromIPFS function to fetch the post from IPFS
			post, err := getPostFromIPFS(hash)
			if err != nil {
				log.Printf("Failed to retrieve post for hash %s: %v", hash, err)
				continue // Skip this hash if retrieval fails
			}

			log.Printf("Checking post with POST ID: %d", post.ID)

			// Step 4: Match the postID
			if post.ID == postIDInt {
				if hash != "" {
					return hash, nil
				}
				return "", fmt.Errorf("post found but IPFS hash is empty for post ID: %s", postID)
			}
		}
	}

	// If no matching post is found, return an error
	return "", fmt.Errorf("no post found with post ID: %s", postID)
}

func main() {

	// Initialize Fabric connection
	if err := initFabric(); err != nil {
		log.Fatalf("Error initializing Fabric: %v", err)
	}
	ipfsShell = shell.NewShell("localhost:5001")

	// Register handlers
	r := mux.NewRouter()
	r.HandleFunc("/signup", SignUpHandler).Methods("POST")
	r.HandleFunc("/login", LoginHandler).Methods("POST")
	r.HandleFunc("/post", PostHandler).Methods("POST")
	r.HandleFunc("/feed", FeedHandler).Methods("GET")
	r.HandleFunc("/post/{id}/react", ReactionHandler).Methods("POST")

	// Apply CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Replace with specific domains for production
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	})
	handler := c.Handler(r)

	// Start HTTP server
	port := "8081" // Set desired port
	log.Printf("Server running on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
