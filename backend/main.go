package main

import (
	"crypto/aes"
	"crypto/cipher"
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
	"strings"
	"time"
    "strconv"
	
	
	

	"github.com/gorilla/mux"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"sync"
	"crypto/sha256"
	"sort"
	"bytes"
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
// In-memory wallet storage
var walletStore = struct {
	sync.Mutex
	wallets []Wallet
}{}

// Post represents a social media post
type Post struct {
	ID            int               `json:"id"`
	User          User              `json:"user"`
	Wallet        Wallet            `json:"wallet"`
	Content       string            `json:"content,omitempty"` // Optional text content
	Timestamp     time.Time         `json:"timestamp"`
	Reactions     map[string]string `json:"reactions,omitempty"`
	ReactionCount int               `json:"reactionCount"`
	ShareCount    int               `json:"shareCount"`
	ImageHash     string            `json:"imageHash,omitempty"` // Add this field for image IPFS hash
	VideoHash     string            `json:"videoHash,omitempty"` // Add this field for video IPFS hash
	IPFSHASH      string            `json:"ipfsHASH,omitempty"`
}

type Group struct {
	ID      string   `json:"id"`
	GroupName    string   `json:"groupname"`
	Members []string `json:"members"`
}

var (
	ipfsShell *shell.Shell
	contract  *client.Contract
)
// var upgrader = websocket.Upgrader{
// 	CheckOrigin: func(r *http.Request) bool {
// 		return true
// 	},
// }

// var connections = make(map[string]*websocket.Conn)


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
		Timeout:             40 * time.Second, // Wait 30 seconds for ping ack
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

func generateWallet() (*Wallet, error) {
    priv, pub, err := generateKeys()
    if err != nil {
        return nil, err
    }

    // Encode the private key in DER format
    privKeyBytes, err := x509.MarshalECPrivateKey(priv)
    if err != nil {
        return nil, fmt.Errorf("failed to encode private key: %v", err)
    }
    privKeyHex := hex.EncodeToString(privKeyBytes)

    // Encode the public key in DER format
    pubKeyBytes, err := x509.MarshalPKIXPublicKey(pub)
    if err != nil {
        return nil, fmt.Errorf("failed to encode public key: %v", err)
    }
    pubKeyHex := hex.EncodeToString(pubKeyBytes)

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

	// Generate wallet for the user (Public/Private Key pair)
	wallet, err := generateWallet()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating wallet: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare the user data for the blockchain
	userData := map[string]interface{}{
		"name":       user.Name,
		"phone":      user.Phone,
		"publicKey":  wallet.PublicKey,
		"privateKey": wallet.PrivateKey,
	}

	// Save keys to a file named {name}.key
	keyFilename := fmt.Sprintf("%s.key", user.Name)
	keyFile, err := os.Create(keyFilename)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating key file: %v", err), http.StatusInternalServerError)
		return
	}
	defer keyFile.Close()

	// Write public and private keys to the file
	keyData := fmt.Sprintf("PublicKey: %s\nPrivateKey: %s", wallet.PublicKey, wallet.PrivateKey)
	if _, err := keyFile.WriteString(keyData); err != nil {
		http.Error(w, fmt.Sprintf("Error writing keys to file: %v", err), http.StatusInternalServerError)
		return
	}

	// Store the user data in the blockchain 
	_, err = contract.SubmitTransaction("RegisterUser", user.Name, user.Phone, wallet.PublicKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error registering user on blockchain: %v", err), http.StatusInternalServerError)
		return
	}

	// Verify that the data was stored on the blockchain
	response, err := contract.EvaluateTransaction("GetUser", wallet.PublicKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching user data from blockchain: %v", err), http.StatusInternalServerError)
		return
	}

	// Expected user data as a JSON string
	expectedData := fmt.Sprintf("{\"name\":\"%s\",\"phone\":\"%s\",\"publicKey\":\"%s\"}", 
		user.Name, user.Phone, wallet.PublicKey)

	// Compare the expected and actual data
	if string(response) == expectedData {
		// Respond with the user data
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(userData)
	} else {
		http.Error(w, "User data verification failed on blockchain", http.StatusInternalServerError)
	}
}

func GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Evaluate transaction to get all users from the blockchain
	response, err := contract.EvaluateTransaction("GetAllUsers")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching users from blockchain: %v", err), http.StatusInternalServerError)
		return
	}

	// Parse the JSON response into a slice of User structs
	var users []User
	if err := json.Unmarshal(response, &users); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing users data: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with the list of users
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}
// LoginHandler handles user login and stores keys in the wallet
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

	// Store keys in wallet
	storeInWallet(request.PublicKey, request.PrivateKey)

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

// storeInWallet adds a key pair to the in-memory wallet store
func storeInWallet(publicKey, privateKey string) {
	walletStore.Lock()
	defer walletStore.Unlock()

	// Check for existing entry
	for _, wallet := range walletStore.wallets {
		if wallet.PublicKey == publicKey {
			fmt.Println("Key pair already exists in wallet.")
			return
		}
	}

	// Add new wallet entry
	walletStore.wallets = append(walletStore.wallets, Wallet{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	})

	fmt.Println("Key pair added to wallet.")
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
		postID :=strconv.Itoa(post.ID)
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
		result, err := submitPostWithRetry(post.Wallet.PublicKey, post.IPFSHASH,postID)
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
			client.WithArguments(publicKey, ipfsHash,postID),
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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the post ID from the URL 
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	postID := parts[2] // Corrected index to parts[3]

	// Retrieve post hash using postID
	

	// Parse the request body 
	var request struct {
		UserPublicKey string `json:"userPublicKey"`
		ReactionType  string `json:"reactionType"`

	}
    

	// Decode the request body and handle potential errors 
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	

	// Validate input fields 
	if request.UserPublicKey == "" {
		http.Error(w, "User public key is required", http.StatusBadRequest)
		return
	}

	// Validate reaction type 
	validReactions := map[string]bool{
		"like":      true,
		"love":      true,
		"laugh":     true,
		"angry":     true,
		"sad":       true,
		"celebrate": true,
	}

	if !validReactions[request.ReactionType] {
		http.Error(w, "Invalid reaction type", http.StatusBadRequest)
		return
	}

    
	postHash, err := getPostHashByID(postID)
	if err != nil {
		log.Printf("Failed to retrieve post hash: %v. PostID: %s", err, postID)
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Add the reaction using the smart contract 
	log.Printf("Submitting transaction: AddReaction with PostHash: %s, UserPublicKey: %s, ReactionType: %s", postHash, request.UserPublicKey, request.ReactionType)
	result, err := contract.SubmitTransaction("AddReaction", postHash, request.UserPublicKey, request.ReactionType)
	if err != nil {
		log.Printf("Failed to add reaction: %v. PostHash: %s, UserPublicKey: %s, ReactionType: %s", err, postHash, request.UserPublicKey, request.ReactionType)
		http.Error(w, "Failed to add reaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Verify result is not empty 
	if len(result) == 0 {
		log.Printf("Empty result from AddReaction transaction. PostHash: %s", postHash)
		http.Error(w, "No response from reaction submission", http.StatusInternalServerError)
		return
	}

	log.Printf("AddReaction transaction successful. Result: %s", string(result))

	// Get the updated post from IPFS 
	post, err := getPostFromIPFS(postHash)
	if err != nil {
		log.Printf("Failed to get updated post from IPFS: %v. PostHash: %s", err, postHash)
		http.Error(w, "Failed to retrieve updated post", http.StatusInternalServerError)
		return
	}

	// Return the updated post 
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(post); err != nil {
		log.Printf("Failed to encode post: %v. Post: %+v", err, post)
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



// SearchUserByName searches the blockchain for a user by their name and retrieves their public key.
func SearchUserByName(name string, contract *client.Contract) (string, error) {
    // Query the blockchain for user details.
    queryResult, err := contract.EvaluateTransaction("QueryUserByName", name)
    if err != nil {
        return "", fmt.Errorf("failed to query user: %v", err)
    }

    // Parse the query result.
    var user User
    if err := json.Unmarshal(queryResult, &user); err != nil {
        return "", fmt.Errorf("failed to unmarshal query result: %v", err)
    }

    if user.PublicKey == "" {
        return "", fmt.Errorf("user not found or public key missing")
    }

    return user.PublicKey, nil
}
// Message represents an individual message in a chat
type Message struct {
    IPFSHash   string    `json:"ipfsHash"`   // IPFS hash of the encrypted message
    Signature  string    `json:"signature"`  // Signature for authenticity
    Sender     string    `json:"sender"`     // Sender's public key
    Receiver   string    `json:"receiver"`   // Receiver's public key
    Timestamp  time.Time `json:"timestamp"`  // Time of message
}

// Chat represents a chat between two users
type Chat struct {
    Participants [2]string `json:"participants"` // Public keys of the two participants
    Messages     []Message `json:"messages"`     // List of messages exchanged
}

// EncryptMessage encrypts plaintext using ECIES with AES-GCM
func EncryptMessage(plainText string, publicKey string) (string, error) {
	// Decode the public key from hex
	log.Printf("Received public key: %s", plainText) 
	decodedKey, err := hex.DecodeString(publicKey)
	log.Printf("decoded key %s", decodedKey)
	if err != nil {
		return "", fmt.Errorf("invalid public key encoding: %v", err)
	}

	// Parse the public key
	pubKey, err := x509.ParsePKIXPublicKey(decodedKey)
	if err != nil {
		return "", fmt.Errorf("invalid public key format: %v", err)
	}

	ecdsaPubKey, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("invalid public key type")
	}

	// Generate an ephemeral private key
	ephemeralPrivKey, err := ecdsa.GenerateKey(ecdsaPubKey.Curve, rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate ephemeral key: %v", err)
	}

	// Compute shared secret
	sharedX, _ := ecdsaPubKey.Curve.ScalarMult(ecdsaPubKey.X, ecdsaPubKey.Y, ephemeralPrivKey.D.Bytes())

	// Derive AES key from shared secret (use the first 32 bytes of sharedX)
	aesKey := sharedX.Bytes()
	if len(aesKey) > 32 {
		aesKey = aesKey[:32]
	}

	// Create AES-GCM cipher
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %v", err)
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create AES-GCM: %v", err)
	}

	// Generate a random nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %v", err)
	}

	// Encrypt the plaintext
	cipherText := aesGCM.Seal(nil, nonce, []byte(plainText), nil)

	// Marshal ephemeral public key
	ephemeralPubKey := elliptic.Marshal(ephemeralPrivKey.Curve, ephemeralPrivKey.PublicKey.X, ephemeralPrivKey.PublicKey.Y)

	// Combine ephemeral public key, nonce, and ciphertext
	encryptedMessage := append(ephemeralPubKey, append(nonce, cipherText...)...)

	log.Printf("Encrypted message length: %d", len(encryptedMessage))  // Add log to check length

	return hex.EncodeToString(encryptedMessage), nil
}

// DecryptMessage decrypts the encrypted message using ECIES with AES-GCM
func DecryptMessage(encryptedText string, privateKey string) (string, error) {
	// Decode the private key
	decodedKey, err := hex.DecodeString(privateKey)
	if err != nil {
		return "", fmt.Errorf("invalid private key encoding: %v", err)
	}

	// Parse the private key
	privKey, err := x509.ParseECPrivateKey(decodedKey)
	if err != nil {
		return "", fmt.Errorf("invalid private key format: %v", err)
	}

	// Decode the encrypted message
	encrypted, err := hex.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted message: %v", err)
	}

	// Extract curve details
	curve := privKey.Curve
	keySize := (curve.Params().BitSize + 7) / 8
	ephemeralKeySize := 2*keySize + 1

	// Check if the message is long enough for the ephemeral public key
	if len(encrypted) < ephemeralKeySize {
		return "", fmt.Errorf("encrypted message is too short for ephemeral key")
	}

	// Extract ephemeral public key
	ephemeralPubKey := encrypted[:ephemeralKeySize]
	ephemeralX, ephemeralY := elliptic.Unmarshal(curve, ephemeralPubKey)
	if ephemeralX == nil || ephemeralY == nil {
		return "", fmt.Errorf("invalid ephemeral public key")
	}

	// Compute shared secret
	sharedX, _ := curve.ScalarMult(ephemeralX, ephemeralY, privKey.D.Bytes())

	// Derive AES key
	aesKey := sharedX.Bytes()
	if len(aesKey) > 32 {
		aesKey = aesKey[:32]
	}

	// Create AES-GCM cipher
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES block: %v", err)
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create AES-GCM: %v", err)
	}

	// Validate nonce and ciphertext sizes
	nonceSize := aesGCM.NonceSize()
	if len(encrypted) < ephemeralKeySize+nonceSize {
		return "", fmt.Errorf("encrypted message is too short for nonce")
	}
    
	nonce := encrypted[ephemeralKeySize : ephemeralKeySize+nonceSize]
	cipherText := encrypted[ephemeralKeySize+nonceSize:]
	if len(cipherText) == 0 {
		return "", fmt.Errorf("no ciphertext found")
	}
    log.Println("Cipher Text %s",cipherText)
    
    
	// Decrypt the ciphertext
	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %v", err)
	}
    log.Println("Plain Text %+v",string(plainText))
	return string(plainText), nil
}


func SignMessage(plainText string, privateKeyHex string) (string, error) {
    // Decode the private key from hex
    privKeyBytes, err := hex.DecodeString(privateKeyHex)
    if err != nil {
        return "", fmt.Errorf("invalid private key encoding: %v", err)
    }

    // Parse the private key from DER format
    privKey, err := x509.ParseECPrivateKey(privKeyBytes)
    if err != nil {
        return "", fmt.Errorf("invalid private key format: %v", err)
    }

    // Compute the SHA-256 hash of the message
    hash := sha256.Sum256([]byte(plainText))
    

    // Sign the hash using the private key
    r, s, err := ecdsa.Sign(rand.Reader, privKey, hash[:])
    if err != nil {
        return "", fmt.Errorf("failed to sign message: %v", err)
    }

   

    // Encode the signature as "r,s"
    return fmt.Sprintf("%s,%s", r.String(), s.String()), nil
}


func VerifySignature(plainText string, signature string, publicKeyHex string) (bool, error) {
    // Decode the public key from hex
	
    pubKeyBytes, err := hex.DecodeString(publicKeyHex)
    if err != nil {
        return false, fmt.Errorf("invalid public key encoding: %v", err)
    }

    // Parse the public key from DER format
    pubKey, err := x509.ParsePKIXPublicKey(pubKeyBytes)
    if err != nil {
        return false, fmt.Errorf("invalid public key format: %v", err)
    }

    // Assert the public key is of type ECDSA
    ecdsaPubKey, ok := pubKey.(*ecdsa.PublicKey)
    if !ok {
        return false, fmt.Errorf("invalid public key type")
    }

    // Compute the SHA-256 hash of the message
    hash := sha256.Sum256([]byte(plainText))
    

    // Split the signature into r and s
    parts := strings.Split(signature, ",")
    if len(parts) != 2 {
        return false, fmt.Errorf("invalid signature format")
    }

    // Parse r and s as big integers
	r := new(big.Int)
	r.SetString(parts[0], 10) // Assign r from the signature
	s := new(big.Int)
	s.SetString(parts[1], 10) // Properly assign s from the signature
	
	

    // Verify the signature using the public key and hash
    isValid := ecdsa.Verify(ecdsaPubKey, hash[:], r, s)
    

    return isValid, nil
}



// UploadToIPFS uploads content to IPFS and returns the IPFS hash
func UploadMessageToIPFS(content string) (string, error) {
    sh := shell.NewShell("localhost:5001") // Ensure IPFS daemon is running on localhost:5001
    hash, err := sh.Add(strings.NewReader(content))
    if err != nil {
        return "", err
    }
    return hash, nil
}
func GenerateChatID(participants []string) string {
    // Sort participants lexicographically (this makes the order consistent)
    sort.Strings(participants)

    // Concatenate the sorted participant public keys into a single string
    concatenated := strings.Join(participants, "")

    // Generate the chat ID using SHA-256 hash of the concatenated string
    chatID := fmt.Sprintf("%x", sha256.Sum256([]byte(concatenated)))
    return chatID
}

func SendMessage(chat *Chat, senderPrivateKey string, senderPublicKey string, receiverPublicKey string, plainText string, chatID string) error {
    // Debugging statement: Log the public key of the receiver
    fmt.Println("Receiver Public Key:", receiverPublicKey)

    // Encrypt the message
    encryptedMessage, err := EncryptMessage(plainText, receiverPublicKey)
    if err != nil {
        fmt.Println("Error while encrypting message:", err) // Print statement for debugging
        return fmt.Errorf("failed to encrypt message: %v", err)
    }
    fmt.Println("Encrypted Message:", encryptedMessage) // Print statement for debugging

    // Upload the encrypted message to IPFS
    ipfsHash, err := UploadMessageToIPFS(encryptedMessage)
    if err != nil {
        fmt.Println("Error while uploading message to IPFS:", err) // Print statement for debugging
        return fmt.Errorf("failed to upload message to IPFS: %v", err)
    }
    fmt.Println("IPFS Hash:", ipfsHash) // Print statement for debugging

    // Sign the original message
    signature, err := SignMessage(plainText, senderPrivateKey)
    if err != nil {
        fmt.Println("Error while signing message:", err) // Print statement for debugging
        return fmt.Errorf("failed to sign message: %v", err)
    }
    fmt.Println("Signature:", signature) // Print statement for debugging

    // Prepare the message object
    message := Message{
        IPFSHash:  ipfsHash,
        Signature: signature,
        Sender:    senderPublicKey,
        Receiver:  receiverPublicKey,
        Timestamp: time.Now(),
    }

    // Log the chat ID for debugging
    fmt.Println("Generated Chat ID:", chatID) // Print statement for debugging

    // Add the message to the blockchain
    err = AddMessageToBlockchain(chatID, message, senderPublicKey, receiverPublicKey)
    if err != nil {
        fmt.Println("Error while adding message to blockchain:", err) // Print statement for debugging
        return fmt.Errorf("failed to add message to blockchain: %v", err)
    }

    // Print statement to confirm successful completion
    fmt.Println("Message successfully sent and added to blockchain.") // Print statement for debugging
    return nil
}

func AddMessageToBlockchain(chatID string, message Message, senderPublicKey string, receiverPublicKey string) error {
    // Convert the message to JSON
    messageBytes, err := json.Marshal(message)
    if err != nil {
        return fmt.Errorf("failed to marshal message: %v", err)
    }

    // Submit the transaction to the blockchain
    result, err := contract.SubmitTransaction("AddMessage", chatID, string(messageBytes), senderPublicKey, receiverPublicKey)
    if err != nil {
        return fmt.Errorf("failed to submit transaction: %v", err)
    }

    // Process the blockchain response (result)
    // If the result is not empty, log it or handle it as necessary
    if len(result) > 0 {
        fmt.Printf("Blockchain response: %s\n", result)
    } else {
        fmt.Println("Transaction submitted successfully with no response.")
    }

    return nil
}



func FetchFromIPFS(ipfsHash string) (string, error) {
    sh := shell.NewShell("localhost:5001") // Ensure IPFS daemon is running on localhost:5001
    file, err := sh.Cat(ipfsHash)
    if err != nil {
        return "", fmt.Errorf("failed to fetch file from IPFS: %v", err)
    }

    // Read the content from IPFS
    content, err := io.ReadAll(file)
    if err != nil {
        return "", fmt.Errorf("failed to read content from IPFS: %v", err)
    }

    return string(content), nil
}
func DecryptAndFetchMessages(chatID string, senderPublicKey string, receiverPublicKey string, senderPrivateKey string, receiverPrivateKey string) ([]string, error) {
    // Fetch the chat data from the blockchain using the chaincode
    chat, err := GetChatFromBlockchain(chatID)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch chat: %v", err)
    }

    var decryptedMessages []string

    // For each message in the chat, fetch the IPFS content, decrypt it, and then verify its signature
    for _, message := range chat.Messages {
        // Fetch the encrypted message from IPFS
        encryptedMessage, err := FetchFromIPFS(message.IPFSHash)
        if err != nil {
            return nil, fmt.Errorf("failed to fetch message from IPFS: %v", err)
        }


        // Determine which private key to use for decryption
        var privateKeyToUse string
        if message.Sender == senderPublicKey {
            privateKeyToUse = receiverPrivateKey
        } else {
            privateKeyToUse = senderPrivateKey
        }

       
        // Decrypt the message using the chosen private key
        decryptedMessage, err := DecryptMessage(encryptedMessage, privateKeyToUse)
        if err != nil {
            return nil, fmt.Errorf("failed to decrypt message: %v", err)
        }
        
        log.Printf("Message sender: %s", message.Sender)
        log.Printf("Message sign: %s", message.Signature)
        // Verify the decrypted message's signature
        isValid, err := VerifySignature(decryptedMessage, message.Signature, message.Sender)
        if err != nil || !isValid {
            if err != nil {
                log.Printf("Signature verification error: %v", err)
            }
            return nil, fmt.Errorf("signature verification failed for decrypted message: %s", decryptedMessage)
        }
        log.Printf("Signature verified for decrypted message: %s", decryptedMessage)

        // Add the decrypted message to the result
        decryptedMessages = append(decryptedMessages, decryptedMessage)
    }

    return decryptedMessages, nil
}


func GetChatFromBlockchain(chatID string) (*Chat, error) {
    // Query the blockchain for the chat data
    result, err := contract.EvaluateTransaction("GetChat", chatID)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch chat: %v", err)
    }

    // Unmarshal the result into a Chat object
    var chat Chat
    err = json.Unmarshal(result, &chat)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal chat data: %v", err)
    }

    return &chat, nil
}



func ChatHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("ChatHandler invoked")

    if r.Method != http.MethodPost {
        log.Println("Invalid method. Only POST is allowed")
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    log.Println("Reading raw request body")
    body, _ := io.ReadAll(r.Body)
    log.Printf("Raw request body: %s", string(body))

    log.Println("Parsing operation and username")
    var baseReq struct {
        Operation string `json:"operation"`
        Username  string `json:"username"`
    }
    err := json.Unmarshal(body, &baseReq)
    if err != nil || (baseReq.Operation != "send" && baseReq.Operation != "get") {
        log.Println("Invalid operation or username specified")
        http.Error(w, "invalid operation or username specified. Use 'send' or 'get'.", http.StatusBadRequest)
        return
    }
    log.Printf("Operation: %s, Username: %s", baseReq.Operation, baseReq.Username)

    log.Println("Loading user keys")
    userKeys, err := loadKeys(baseReq.Username)
    if err != nil {
        log.Printf("Failed to load keys for user %s: %v", baseReq.Username, err)
        http.Error(w, fmt.Sprintf("failed to load keys for user %s: %v", baseReq.Username, err), http.StatusInternalServerError)
        return
    }
   
    log.Println("Resetting the request body for further parsing")
    r.Body = io.NopCloser(bytes.NewBuffer(body))

    if baseReq.Operation == "send" {
        log.Println("Handling 'send' operation")
        var sendReq struct {
            ReceiverUsername string `json:"receiverUsername"`
            PlainText        string `json:"plainText"`
        }

        log.Println("Decoding 'send' request")
        err := json.NewDecoder(r.Body).Decode(&sendReq)
        if err != nil {
            log.Println("Invalid request body for 'send'")
            http.Error(w, "invalid request body for 'send'", http.StatusBadRequest)
            return
        }
        log.Printf("ReceiverUsername: %s, PlainText: %s", sendReq.ReceiverUsername, sendReq.PlainText)

        log.Println("Loading receiver's keys")
        receiverKeys, err := loadKeys(sendReq.ReceiverUsername)
        if err != nil {
            log.Printf("Failed to load keys for receiver %s: %v", sendReq.ReceiverUsername, err)
            http.Error(w, fmt.Sprintf("failed to load keys for receiver %s: %v", sendReq.ReceiverUsername, err), http.StatusInternalServerError)
            return
        }
        

        participants := []string{userKeys.PublicKey, receiverKeys.PublicKey}
        chatID := GenerateChatID(participants)
        log.Printf("Generated chat ID: %s", chatID)

        log.Println("Sending the message")
        err = SendMessage(&Chat{}, userKeys.PrivateKey, userKeys.PublicKey, receiverKeys.PublicKey, sendReq.PlainText, chatID)
        if err != nil {
            log.Printf("Failed to send message: %v", err)
            http.Error(w, fmt.Sprintf("failed to send message: %v", err), http.StatusInternalServerError)
            return
        }
        log.Println("Message sent successfully")

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"status": "message sent successfully"})
        return
    }

    if baseReq.Operation == "get" {
        log.Println("Handling 'get' operation")
        var getReq struct {
            SenderUsername string `json:"senderUsername"`
        }

        log.Println("Decoding 'get' request")
        err := json.NewDecoder(r.Body).Decode(&getReq)
        if err != nil {
            log.Println("Invalid request body for 'get'")
            http.Error(w, "invalid request body for 'get'", http.StatusBadRequest)
            return
        }
        log.Printf("SenderUsername: %s", getReq.SenderUsername)

        log.Println("Loading sender's keys")
        senderKeys, err := loadKeys(getReq.SenderUsername)
        if err != nil {
            log.Printf("Failed to load keys for sender %s: %v", getReq.SenderUsername, err)
            http.Error(w, fmt.Sprintf("failed to load keys for sender %s: %v", getReq.SenderUsername, err), http.StatusInternalServerError)
            return
        }
        log.Println("Sender's keys loaded successfully")
        
        participants := []string{userKeys.PublicKey, senderKeys.PublicKey}
        chatID := GenerateChatID(participants)
        log.Printf("Generated chat ID: %s", chatID)

        log.Println("Fetching and decrypting messages")
        decryptedMessages, err := DecryptAndFetchMessages(chatID, senderKeys.PublicKey, userKeys.PublicKey, senderKeys.PrivateKey, userKeys.PrivateKey)
        if err != nil {
            log.Printf("Failed to fetch chat messages: %v", err)
            http.Error(w, fmt.Sprintf("failed to fetch chat messages: %v", err), http.StatusInternalServerError)
            return
        }
        log.Println("Messages fetched and decrypted successfully %v",decryptedMessages)

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(decryptedMessages)
        return
    }

    log.Println("Invalid operation specified")
    http.Error(w, "invalid operation", http.StatusInternalServerError)
}


// Helper function to load and parse keys from a file
func loadKeys(username string) (struct {
    PublicKey  string
    PrivateKey string
}, error) {
    var keys struct {
        PublicKey  string
        PrivateKey string
    }

    // Read the key file
    keyFile := fmt.Sprintf("%s.key", username)
    content, err := os.ReadFile(keyFile)
    if err != nil {
        return keys, fmt.Errorf("failed to read key file: %v", err)
    }

    // Parse the content for keys
    lines := strings.Split(string(content), "\n")
    for _, line := range lines {
        if strings.HasPrefix(line, "PublicKey:") {
            keys.PublicKey = strings.TrimSpace(strings.TrimPrefix(line, "PublicKey:"))
        } else if strings.HasPrefix(line, "PrivateKey:") {
            keys.PrivateKey = strings.TrimSpace(strings.TrimPrefix(line, "PrivateKey:"))
        }
    }

    if keys.PublicKey == "" || keys.PrivateKey == "" {
        return keys, fmt.Errorf("incomplete keys in file")
    }

    return keys, nil
}

// HTTP Handler to create a new group
func CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
    var groupRequest struct {
        GroupName    string   `json:"groupname"`
        Members []string `json:"members"`
    }

    // Decode the request body
    if err := json.NewDecoder(r.Body).Decode(&groupRequest); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    // Validate input
    if groupRequest.GroupName == "" || len(groupRequest.Members) == 0 {
        http.Error(w, "Group name and members are required", http.StatusBadRequest)
        return
    }

    // Generate a unique group ID
    groupID := fmt.Sprintf("group-%d", time.Now().UnixNano())

    // Serialize member names
    membersJSON, err := json.Marshal(groupRequest.Members)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to serialize member names: %v", err), http.StatusInternalServerError)
        return
    }

    // Create group on blockchain using member names
    _, err = contract.SubmitTransaction("CreateGroup", groupID, groupRequest.GroupName, string(membersJSON))
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to create group on blockchain: %v", err), http.StatusInternalServerError)
        return
    }

    // Prepare the response
    newGroup := map[string]interface{}{
        "id":      groupID,
        "name":    groupRequest.GroupName,
        "members": groupRequest.Members,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(newGroup)
}

// HTTP Handler to retrieve user's groups by user name
func GetAllGroupsHandler(w http.ResponseWriter, r *http.Request) {
    // Extract user name from request
    var requestBody struct {
        UserName string `json:"user_name"`
    }

    // Decode the request body
    if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    // Validate user name
    if requestBody.UserName == "" {
        http.Error(w, "User name is required", http.StatusBadRequest)
        return
    }

    // Retrieve all groups from the blockchain
    allGroupsData, err := contract.EvaluateTransaction("GetAllGroups")
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to retrieve groups: %v", err), http.StatusInternalServerError)
        return
    }

    // Log the raw data returned from the blockchain
    log.Printf("Raw groups data: %s", string(allGroupsData))

    // Deserialize groups
    var allGroups []*Group
    if err := json.Unmarshal(allGroupsData, &allGroups); err != nil {
        http.Error(w, "Failed to process groups", http.StatusInternalServerError)
        return
    }

    // Filter groups where the provided user is a member
    var userGroups []*Group
    for _, group := range allGroups {
        for _, member := range group.Members {
            if member == requestBody.UserName {
                userGroups = append(userGroups, group)
                break
            }
        }
    }

    // Log the filtered groups
    log.Printf("Filtered groups for user %s: %+v", requestBody.UserName, userGroups)

    // Prepare response
    response := struct {
        Groups []*Group `json:"groups"`
    }{
        Groups: userGroups,
    }

    // Set response headers and encode
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

//-----------------------------------------------------------//

func generateGroupChatID(groupID string) string {
    // Concatenate groupID and senderPublicKey
	
    combined := groupID 
    // Hash the concatenated string to generate the chatID
    hash := sha256.New()
    hash.Write([]byte(combined))
    chatID := hash.Sum(nil)
    
    // Return the chatID as a hexadecimal string
    return hex.EncodeToString(chatID)
}

func SendGroupMessage(chat *Chat, senderPrivateKey string, senderPublicKey string, receiverPublicKey string, plainText string, chatID string) error {
    // Debugging statement: Log the public key of the receiver
    fmt.Println("Receiver Public Key:", receiverPublicKey)

    // Encrypt the message
    encryptedMessage, err := EncryptGroupMessage(plainText, receiverPublicKey)
    if err != nil {
        fmt.Println("Error while encrypting message:", err) // Print statement for debugging
        return fmt.Errorf("failed to encrypt message: %v", err)
    }
    fmt.Println("Encrypted Message:", encryptedMessage) // Print statement for debugging

    // Upload the encrypted message to IPFS
    ipfsHash, err := UploadMessageToIPFS(encryptedMessage)
    if err != nil {
        fmt.Println("Error while uploading message to IPFS:", err) // Print statement for debugging
        return fmt.Errorf("failed to upload message to IPFS: %v", err)
    }
    fmt.Println("IPFS Hash:", ipfsHash) // Print statement for debugging

    // Sign the original message
    signature, err := SignGroupMessage(plainText, senderPrivateKey)
    if err != nil {
        fmt.Println("Error while signing message:", err) // Print statement for debugging
        return fmt.Errorf("failed to sign message: %v", err)
    }
    fmt.Println("Signature:", signature) // Print statement for debugging

    // Prepare the message object
    message := Message{
        IPFSHash:  ipfsHash,
        Signature: signature,
        Sender:    senderPublicKey,
        Receiver:  receiverPublicKey,
        Timestamp: time.Now(),
    }

    // Log the chat ID for debugging
    fmt.Println("Generated Chat ID:", chatID) // Print statement for debugging

    // Add the message to the blockchain
    err = AddMessageToBlockchain(chatID, message, senderPublicKey, receiverPublicKey)
    if err != nil {
        fmt.Println("Error while adding message to blockchain:", err) // Print statement for debugging
        return fmt.Errorf("failed to add message to blockchain: %v", err)
    }

    // Print statement to confirm successful completion
    fmt.Println("Message successfully sent and added to blockchain.") // Print statement for debugging
    return nil
}


// EncryptMessage encrypts plaintext using ECIES with AES-GCM for a specific recipient
func EncryptGroupMessage(plainText string, publicKey string) (string, error) {
	// Validate input
	if plainText == "" || publicKey == "" {
		return "", fmt.Errorf("empty plaintext or public key")
	}

	// Decode the public key from hex
	decodedKey, err := hex.DecodeString(publicKey)
	if err != nil {
		return "", fmt.Errorf("invalid public key encoding: %v", err)
	}

	// Parse the public key
	pubKey, err := x509.ParsePKIXPublicKey(decodedKey)
	if err != nil {
		return "", fmt.Errorf("invalid public key format: %v", err)
	}

	// Ensure the key is an ECDSA public key
	ecdsaPubKey, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("invalid public key type")
	}

	// Generate an ephemeral private key on the same curve as the recipient's public key
	ephemeralPrivKey, err := ecdsa.GenerateKey(ecdsaPubKey.Curve, rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate ephemeral key: %v", err)
	}

	// Compute shared secret using ECDH key agreement
	sharedX, sharedY := ecdsaPubKey.Curve.ScalarMult(ecdsaPubKey.X, ecdsaPubKey.Y, ephemeralPrivKey.D.Bytes())

	// Derive AES key using SHA-256 hash of shared secret
	sharedSecret := append(sharedX.Bytes(), sharedY.Bytes()...)
	hash := sha256.Sum256(sharedSecret)
	aesKey := hash[:32]

	// Create AES-GCM cipher
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %v", err)
	}
	
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create AES-GCM: %v", err)
	}

	// Generate a cryptographically secure random nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %v", err)
	}

	// Encrypt the plaintext
	cipherText := aesGCM.Seal(nil, nonce, []byte(plainText), nil)

	// Marshal ephemeral public key
	ephemeralPubKey := elliptic.Marshal(ephemeralPrivKey.Curve, ephemeralPrivKey.PublicKey.X, ephemeralPrivKey.PublicKey.Y)

	// Combine components: ephemeral public key + nonce + ciphertext
	encryptedMessage := append(ephemeralPubKey, append(nonce, cipherText...)...)

	// Convert to hex for storage/transmission
	return hex.EncodeToString(encryptedMessage), nil
}

// DecryptMessage decrypts a message encrypted with ECIES and AES-GCM
func DecryptGroupMessage(encryptedText string, privateKey string) (string, error) {
	// Validate input
	if encryptedText == "" || privateKey == "" {
		return "", fmt.Errorf("empty encrypted text or private key")
	}

	// Decode the private key
	decodedPrivKey, err := hex.DecodeString(privateKey)
	if err != nil {
		return "", fmt.Errorf("invalid private key encoding: %v", err)
	}

	// Parse the private key
	privKey, err := x509.ParseECPrivateKey(decodedPrivKey)
	if err != nil {
		return "", fmt.Errorf("invalid private key format: %v", err)
	}

	// Decode the encrypted message
	encrypted, err := hex.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted message: %v", err)
	}

	// Extract curve details
	curve := privKey.Curve
	keySize := (curve.Params().BitSize + 7) / 8
	ephemeralKeySize := 2*keySize + 1

	// Validate message length
	nonceSize := 12 // Standard AES-GCM nonce size
	if len(encrypted) < ephemeralKeySize+nonceSize {
		return "", fmt.Errorf("encrypted message is too short")
	}

	// Extract ephemeral public key
	ephemeralPubKey := encrypted[:ephemeralKeySize]
	ephemeralX, ephemeralY := elliptic.Unmarshal(curve, ephemeralPubKey)
	if ephemeralX == nil || ephemeralY == nil {
		return "", fmt.Errorf("invalid ephemeral public key")
	}

	// Extract nonce
	nonce := encrypted[ephemeralKeySize:ephemeralKeySize+nonceSize]

	// Extract ciphertext
	cipherText := encrypted[ephemeralKeySize+nonceSize:]

	// Compute shared secret using ECDH
	sharedX, sharedY := curve.ScalarMult(ephemeralX, ephemeralY, privKey.D.Bytes())

	// Derive AES key using SHA-256 hash of shared secret
	sharedSecret := append(sharedX.Bytes(), sharedY.Bytes()...)
	hash := sha256.Sum256(sharedSecret)
	aesKey := hash[:32]

	// Create AES-GCM cipher
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES block: %v", err)
	}
	
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create AES-GCM: %v", err)
	}

	// Decrypt the ciphertext
	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %v", err)
	}

	return string(plainText), nil
}

// Optional: Helper function to generate a signature for the message
func SignGroupMessage(plainText string, privateKey string) (string, error) {
	// Decode the private key
	decodedPrivKey, err := hex.DecodeString(privateKey)
	if err != nil {
		return "", fmt.Errorf("invalid private key encoding: %v", err)
	}

	// Parse the private key
	privKey, err := x509.ParseECPrivateKey(decodedPrivKey)
	if err != nil {
		return "", fmt.Errorf("invalid private key format: %v", err)
	}

	// Compute hash of the message
	hash := sha256.Sum256([]byte(plainText))

	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %v", err)
	}

	// Return signature as comma-separated r and s values
	return fmt.Sprintf("%s,%s", r.String(), s.String()), nil
}

// Optional: Helper function to verify signature
func VerifyGroupSignature(plainText string, signature string, publicKeyHex string) (bool, error) {
	// Decode the public key
	pubKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return false, fmt.Errorf("invalid public key encoding: %v", err)
	}

	// Parse the public key
	pubKey, err := x509.ParsePKIXPublicKey(pubKeyBytes)
	if err != nil {
		return false, fmt.Errorf("invalid public key format: %v", err)
	}

	// Assert the public key is of type ECDSA
	ecdsaPubKey, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("invalid public key type")
	}

	// Split the signature into r and s
	parts := bytes.Split([]byte(signature), []byte(","))
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid signature format")
	}

	// Parse r and s
	r := new(big.Int)
	r.SetString(string(parts[0]), 10)
	s := new(big.Int)
	s.SetString(string(parts[1]), 10)

	// Compute hash of the message
	hash := sha256.Sum256([]byte(plainText))

	// Verify the signature
	return ecdsa.Verify(ecdsaPubKey, hash[:], r, s), nil
}



func DecryptAndFetchGroupMessages(chatID string, senderPublicKey string, receiverPublicKey string, senderPrivateKey string, receiverPrivateKey string, waitTime time.Duration) ([]string, error) {
	// Fetch the chat data from the blockchain using the chaincode
	chat, err := GetChatFromBlockchain(chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chat: %v", err)
	}

	var decryptedMessages []string
	var mu sync.Mutex

	// For each message in the chat, fetch the IPFS content, decrypt it, and then verify its signature
	for _, message := range chat.Messages {
		// Simulate wait time until decryption
		log.Printf("Waiting for %v before decrypting message from sender: %s", waitTime, message.Sender)
		time.Sleep(waitTime)

		// Fetch the encrypted message from IPFS
		encryptedMessage, err := FetchFromIPFS(message.IPFSHash)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch message from IPFS: %v", err)
		}

		// Determine which private key to use for decryption
		var privateKeyToUse string
		if message.Sender == senderPublicKey {
			privateKeyToUse = receiverPrivateKey
		} else {
			privateKeyToUse = senderPrivateKey
		}

		log.Printf("Sender Public Key: %s", message.Sender)
		log.Printf("Receiver Public Key: %s", receiverPublicKey)
		log.Printf("Private Key being used: %s", privateKeyToUse)
		log.Printf("Encrypted Message: %s", encryptedMessage)

		// Decrypt the message using the chosen private key
		decryptedMessage, err := DecryptGroupMessage(encryptedMessage, privateKeyToUse)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt message: %v", err)
		}

		log.Printf("Message sender: %s", message.Sender)
		log.Printf("Message signature: %s", message.Signature)

		// Verify the decrypted message's signature
		isValid, err := VerifyGroupSignature(decryptedMessage, message.Signature, message.Sender)
		if err != nil || !isValid {
			if err != nil {
				log.Printf("Signature verification error: %v", err)
			}
			return nil, fmt.Errorf("signature verification failed for decrypted message: %s", decryptedMessage)
		}
		log.Printf("Signature verified for decrypted message: %s", decryptedMessage)

		// Synchronize access to the decryptedMessages slice
		mu.Lock()
		decryptedMessages = append(decryptedMessages, decryptedMessage)
		mu.Unlock()
	}

	return decryptedMessages, nil
}

func GroupChatHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("ChatHandler invoked")

    if r.Method != http.MethodPost {
        log.Println("Invalid method. Only POST is allowed")
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    log.Println("Reading raw request body")
    body, _ := io.ReadAll(r.Body)
    log.Printf("Raw request body: %s", string(body))

    log.Println("Parsing operation, username, groupID, and participants")
    var baseReq struct {
        Operation    string   `json:"operation"`
        GroupID      string   `json:"groupID"`
        Participants []string `json:"participants"`
    }
    err := json.Unmarshal(body, &baseReq)
    if err != nil || (baseReq.Operation != "send" && baseReq.Operation != "get") || baseReq.GroupID == "" {
        log.Println("Invalid operation or groupID specified")
        http.Error(w, "invalid operation or groupID specified. Use 'send' or 'get'.", http.StatusBadRequest)
        return
    }
    
    log.Printf("Operation: %s, GroupID: %s", baseReq.Operation, baseReq.GroupID)
    
    if len(baseReq.Participants) == 0 {
        log.Println("No participants in the group")
        http.Error(w, "no participants in the group", http.StatusBadRequest)
        return
    }

    for _, participant := range baseReq.Participants {
        log.Printf("Processing participant: %s", participant)

        if baseReq.Operation == "send" {
            log.Println("Handling 'send' operation")
            var sendReq struct {
                Username  string `json:"username"`
                PlainText string `json:"plainText"`
            }

            r.Body = io.NopCloser(bytes.NewBuffer(body))

            log.Println("Decoding 'send' request")
            err := json.NewDecoder(r.Body).Decode(&sendReq)
            if err != nil {
                log.Println("Invalid request body for 'send'")
                http.Error(w, "invalid request body for 'send'", http.StatusBadRequest)
                return
            }
            log.Printf("Sender: %s, PlainText: %s", sendReq.Username, sendReq.PlainText)

            if participant == sendReq.Username {
                log.Printf("Skipping sender: %s", participant)
                continue
            }

            

            senderKeys, err := loadKeys(sendReq.Username)
            if err != nil {
                log.Printf("Failed to load keys for sender %s: %v", sendReq.Username, err)
                http.Error(w, fmt.Sprintf("failed to load keys for sender %s: %v", sendReq.Username, err), http.StatusInternalServerError)
                return
            }

            participantKeys, err := loadKeys(participant)
            if err != nil {
                log.Printf("Failed to load keys for participant %s: %v", participant, err)
                http.Error(w, fmt.Sprintf("failed to load keys for participant %s: %v", participant, err), http.StatusInternalServerError)
                return
            }

            chatID := generateGroupChatID(baseReq.GroupID)
            log.Printf("Generated chat ID: %s", chatID)

            err = SendGroupMessage(&Chat{}, senderKeys.PrivateKey, senderKeys.PublicKey, participantKeys.PublicKey, sendReq.PlainText, chatID)
            if err != nil {
                log.Printf("Failed to send message to participant %s: %v", participant, err)
                continue
            }
            log.Printf("Message sent successfully to participant %s", participant)
        }

        if baseReq.Operation == "get" {
            log.Println("Handling 'get' operation")
            var getReq struct {
                SenderUsername string `json:"senderUsername"`
            }

            r.Body = io.NopCloser(bytes.NewBuffer(body))

            log.Println("Decoding 'get' request")
            err := json.NewDecoder(r.Body).Decode(&getReq)
            if err != nil {
                log.Println("Invalid request body for 'get'")
                http.Error(w, "invalid request body for 'get'", http.StatusBadRequest)
                return
            }

            log.Printf("Fetching messages for participant: %s", participant)

            if participant == getReq.SenderUsername {
                log.Printf("Skipping sender: %s for message fetching", participant)
                continue
            }
            participantKeys, err := loadKeys(participant)
            if err != nil {
                log.Printf("Failed to load keys for participant %s: %v", participant, err)
                http.Error(w, fmt.Sprintf("failed to load keys for participant %s: %v", participant, err), http.StatusInternalServerError)
                continue
            }

            senderKeys, err := loadKeys(getReq.SenderUsername)
            if err != nil {
                log.Printf("Failed to load keys for sender %s: %v", getReq.SenderUsername, err)
                http.Error(w, fmt.Sprintf("failed to load keys for sender %s: %v", getReq.SenderUsername, err), http.StatusInternalServerError)
                continue
            }
            
            log.Println("Fetching messages...")
      
            

            chatID := generateGroupChatID(baseReq.GroupID)
            decryptedMessages, err := DecryptAndFetchGroupMessages(chatID, senderKeys.PublicKey, participantKeys.PublicKey, senderKeys.PrivateKey, participantKeys.PrivateKey,10*time.Second)
            if err != nil {
                log.Printf("Failed to fetch messages for participant %s: %v", participant, err)
                continue
            }

            log.Printf("Messages fetched successfully for participant %s", participant)
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(map[string]interface{}{"participant": participant, "messages": decryptedMessages})
        }
    }
}



//-----------------------------------------------------------//
func main() {

	// Initialize Fabric connection
	if err := initFabric(); err != nil {
		log.Fatalf("Error initializing Fabric: %v", err)
	}

	// Register handlers
	r := mux.NewRouter()
	r.HandleFunc("/signup", SignUpHandler).Methods("POST")
	r.HandleFunc("/login", LoginHandler).Methods("POST")
	r.HandleFunc("/post", PostHandler).Methods("POST")
	r.HandleFunc("/feed", FeedHandler).Methods("GET")
	r.HandleFunc("/post/{id}/react", ReactionHandler).Methods("POST")
	r.HandleFunc("/users", GetAllUsersHandler).Methods("GET")
	r.HandleFunc("/chat", ChatHandler)
	r.HandleFunc("/groups", CreateGroupHandler).Methods("POST")
    r.HandleFunc("/usergroups",GetAllGroupsHandler).Methods("POST")
	r.HandleFunc("/groupchat", GroupChatHandler)
	//r.HandleFunc("/getchat", GetChatMessagesHandler)
    
	
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
