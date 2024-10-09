package main

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type User struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Password       string `json:"password"`
	ProfilePicture string `json:"profilePicture,omitempty"`
}

type Post struct {
	ID            int               `json:"id"`
	User          User              `json:"user"`
	Content       string            `json:"content"`
	Reactions     map[string]string `json:"reactions,omitempty"`
	ReactionCount int               `json:"reactionCount"`
}

type Reaction struct {
	PostID int    `json:"postId"`
	UserID string `json:"userId"`
	Type   string `json:"type"`
}

var ipfsShell *shell.Shell
var contract *client.Contract

func initFabric() error {
	clientConnection := newGrpcConnection()
	id := newIdentity()
	sign := newSign()

	gateway, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		client.WithEvaluateTimeout(5),
		client.WithEndorseTimeout(15),
		client.WithSubmitTimeout(5),
		client.WithCommitStatusTimeout(1),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to gateway: %w", err)
	}

	network := gateway.GetNetwork("mychannel")
	contract = network.GetContract("social_media")
	return nil
}

func newGrpcConnection() *grpc.ClientConn {
	certificate, err := loadCertificate()
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, "peer0.org1.example.com")

	connection, err := grpc.Dial("localhost:7051", grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}

	return connection
}

func loadCertificate() (*x509.Certificate, error) {
	pemPath := filepath.Join(
		"..",
		"..",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"peers",
		"peer0.org1.example.com",
		"tls",
		"ca.crt",
	)

	certificatePEM, err := os.ReadFile(pemPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	return identity.CertificateFromPEM(certificatePEM)
}

func newIdentity() *identity.X509Identity {
	certificatePath := filepath.Join(
		"..",
		"..",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
		"signcerts",
		"cert.pem",
	)
	certificatePEM, err := os.ReadFile(certificatePath)
	if err != nil {
		panic(fmt.Errorf("failed to read certificate file: %w", err))
	}

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		panic(fmt.Errorf("failed to create certificate from PEM: %w", err))
	}

	id, err := identity.NewX509Identity("Org1MSP", certificate)
	if err != nil {
		panic(fmt.Errorf("failed to create X509 identity: %w", err))
	}

	return id
}

func newSign() identity.Sign {
	keyPath := filepath.Join(
		"..",
		"..",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
		"keystore",
		"46f45f1fa65eede3d6fbbf315989c37a746610b8c7cf309b85f87d7a2629d9f6_sk",
	)
	privateKeyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read private key file: %w", err))
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(fmt.Errorf("failed to create private key: %w", err))
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(fmt.Errorf("failed to create signing function: %w", err))
	}

	return sign
}


func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Store user data in IPFS
	userJSON, _ := json.Marshal(user)
	hash, err := ipfsShell.Add(strings.NewReader(string(userJSON)))
	if err != nil {
		http.Error(w, "Failed to store in IPFS", http.StatusInternalServerError)
		return
	}

	// Store IPFS hash in blockchain
	_, err = contract.Submit(
		"CreateUser",
		client.WithArguments(user.Email, hash),
	)
	if err != nil {
		http.Error(w, "Failed to store in blockchain", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "User registered successfully: %s\n", user.Email)
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var post Post
		if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		// Store post content in IPFS
		postJSON, _ := json.Marshal(post)
		hash, err := ipfsShell.Add(strings.NewReader(string(postJSON)))
		if err != nil {
			http.Error(w, "Failed to store in IPFS", http.StatusInternalServerError)
			return
		}

		// Store IPFS hash in blockchain
		_, err = contract.Submit(
			"CreatePost",
			client.WithArguments(fmt.Sprintf("%d", post.ID), post.User.Email, hash),
		)
		if err != nil {
			http.Error(w, "Failed to store in blockchain", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(post)

	case http.MethodGet:
		// Implementation for getting posts
		// This would involve querying the blockchain and IPFS
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]*Post{})
	}
}

func main() {
	// Initialize IPFS shell
	ipfsShell = shell.NewShell("localhost:5001")
	log.Println("Connected to IPFS")

	// Initialize Fabric gateway
	if err := initFabric(); err != nil {
		log.Fatalf("Failed to initialize Fabric gateway: %v", err)
	}

	router := mux.NewRouter()

	router.HandleFunc("/signup", SignUpHandler).Methods(http.MethodPost)
	router.HandleFunc("/feed", PostHandler).Methods(http.MethodPost, http.MethodGet)

	handler := cors.Default().Handler(router)

	log.Fatal(http.ListenAndServe(":8081", handler))
}