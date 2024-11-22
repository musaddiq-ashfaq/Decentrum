package main

import (
	"encoding/json"
	"fmt"
	"log"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// User struct defines the user structure
// User represents the user structure in the application (including public/private keys)
type User struct {
	Name           string `json:"name"`
	Phone          string `json:"phone"`
	PublicKey  	   string `json:"publicKey"`
	
}

type Post struct {
	UserPublicKey     string            `json:"userPublicKey"`
	ContentCID    string            `json:"contentCID"`
	Timestamp     int64             `json:"timestamp"`
	Reactions     map[string]string `json:"reactions"`
	ReactionCount int               `json:"reactionCount"`
	ShareCount    int                `json:"shareCount"`
}

// SmartContract defines the chaincode structure
type SmartContract struct {
	contractapi.Contract
}

// RegisterUser registers a user with their public key and stores user data
func (s *SmartContract) RegisterUser(ctx contractapi.TransactionContextInterface, name string, phone string, publicKey string) error {
	// Check if user already exists
	userExists, err := s.UserExists(ctx, publicKey)
	if err != nil {
		return fmt.Errorf("error checking if user exists: %v", err)
	}
	if userExists {
		return fmt.Errorf("user with public key %s already exists", publicKey)
	}

	

	// Create a new user object
	user := User{
		Name:           name,
		Phone:          phone,
		PublicKey:  publicKey,
	
	}

	// Convert user struct to JSON
	userJSON, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %v", err)
	}

	// Store the user data on the ledger
	err = ctx.GetStub().PutState(publicKey, userJSON)
	if err != nil {
		return fmt.Errorf("failed to store user data on ledger: %v", err)
	}

	// Successfully stored the user
	return nil
}

// GetUser retrieves a user's data based on their public key
func (s *SmartContract) GetUser(ctx contractapi.TransactionContextInterface, publicKey string) (*User, error) {
	// Check if user exists
	userExists, err := s.UserExists(ctx, publicKey)
	if err != nil {
		return nil, fmt.Errorf("error checking if user exists: %v", err)
	}
	if !userExists {
		return nil, fmt.Errorf("user with public key %s does not exist", publicKey)
	}

	// Retrieve the user data from the ledger
	userJSON, err := ctx.GetStub().GetState(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get user data: %v", err)
	}
	if userJSON == nil {
		return nil, fmt.Errorf("user data not found for public key %s", publicKey)
	}

	// Unmarshal the JSON data into a User object
	var user User
	err = json.Unmarshal(userJSON, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %v", err)
	}

	return &user, nil
}

// UserExists checks whether a user already exists on the ledger by their public key
func (s *SmartContract) UserExists(ctx contractapi.TransactionContextInterface, publicKey string) (bool, error) {
	userJSON, err := ctx.GetStub().GetState(publicKey)
	if err != nil {
		return false, fmt.Errorf("failed to read user data for public key %s: %v", publicKey, err)
	}
	return userJSON != nil, nil
}

func (s *SmartContract) CreatePost(ctx contractapi.TransactionContextInterface, publicKey string, ipfsHash string) error {
	// Check if the user exists
	userBytes, err := ctx.GetStub().GetState(publicKey)
	if err != nil {
		return fmt.Errorf("failed to read user data: %v", err)
	}
	if userBytes == nil {
		return fmt.Errorf("user does not exist: %s", publicKey)
	}

	// Create a composite key for posts
	postsKey, err := ctx.GetStub().CreateCompositeKey("posts", []string{publicKey})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// Get existing posts
	postsBytes, err := ctx.GetStub().GetState(postsKey)
	var posts []string
	if err != nil {
		return fmt.Errorf("failed to read posts: %v", err)
	}
	if postsBytes != nil {
		err = json.Unmarshal(postsBytes, &posts)
		if err != nil {
			return fmt.Errorf("failed to unmarshal posts: %v", err)
		}
	}

	// Add new post
	posts = append(posts, ipfsHash)
	updatedPostsBytes, err := json.Marshal(posts)
	if err != nil {
		return fmt.Errorf("failed to marshal updated posts: %v", err)
	}

	// Put the updated posts back to the state
	err = ctx.GetStub().PutState(postsKey, updatedPostsBytes)
	if err != nil {
		return fmt.Errorf("failed to update posts: %v", err)
	}

	// Create a separate key for all posts (for easier retrieval in GetAllPosts)
	allPostsKey, err := ctx.GetStub().CreateCompositeKey("allposts", []string{ipfsHash})
	if err != nil {
		return fmt.Errorf("failed to create all posts composite key: %v", err)
	}

	// Store the post under the all posts key
	err = ctx.GetStub().PutState(allPostsKey, []byte(publicKey))
	if err != nil {
		return fmt.Errorf("failed to store post in all posts: %v", err)
	}

	// Log the creation of the post
	log.Printf("Created post for user %s with IPFS hash: %s", publicKey, ipfsHash)
	log.Printf("Stored post with composite key: %s", allPostsKey)

	return nil
}

// GetPost retrieves a post by ID
func (s *SmartContract) GetPost(ctx contractapi.TransactionContextInterface, postID string) (*Post, error) {
	postJSON, err := ctx.GetStub().GetState(postID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if postJSON == nil {
		return nil, fmt.Errorf("post does not exist: %s", postID)
	}

	var post Post
	err = json.Unmarshal(postJSON, &post)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (s *SmartContract) GetAllPosts(ctx contractapi.TransactionContextInterface) (string, error) {
	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("allposts", []string{})
	if err != nil {
		log.Printf("Failed to get iterator for all posts: %v", err)
		return "[]", fmt.Errorf("failed to get all posts: %v", err)
	}
	defer resultsIterator.Close()

	var posts []string
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			log.Printf("Failed to get next item from iterator: %v", err)
			return "[]", fmt.Errorf("failed to iterate posts: %v", err)
		}

		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(queryResponse.Key)
		if err != nil {
			log.Printf("Failed to split composite key: %v", err)
			continue
		}

		if len(compositeKeyParts) > 0 {
			posts = append(posts, compositeKeyParts[0])
			log.Printf("Added post with IPFS hash: %s", compositeKeyParts[0])
		}
	}

	// Ensure we always return a valid JSON array, even if it's empty
	jsonPosts, err := json.Marshal(posts)
	if err != nil {
		log.Printf("Failed to marshal posts: %v", err)
		return "[]", fmt.Errorf("failed to marshal posts: %v", err)
	}

	log.Printf("Retrieved %d posts", len(posts))
	log.Printf("Returning JSON: %s", string(jsonPosts))

	return string(jsonPosts), nil
}

// AddReaction allows a user to react to a post
func (s *SmartContract) AddReaction(ctx contractapi.TransactionContextInterface, postID string, publicKey string, reactionType string) error {
	post, err := s.GetPost(ctx, postID)
	if err != nil {
		return err
	}

	// Add or update reaction
	post.Reactions[publicKey] = reactionType
	post.ReactionCount = len(post.Reactions)

	// Marshal the updated post
	postJSON, err := json.Marshal(post)
	if err != nil {
		return err
	}

	// Update the post in state
	return ctx.GetStub().PutState(postID, postJSON)
}

// GetPostsByUser retrieves all posts created by a specific user
func (s *SmartContract) GetPostsByUser(ctx contractapi.TransactionContextInterface, publicKey string) ([]string, error) {
	// Check if the user exists
	userBytes, err := ctx.GetStub().GetState(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read user data: %v", err)
	}
	if userBytes == nil {
		return nil, fmt.Errorf("user does not exist: %s", publicKey)
	}

	// Create the composite key for the user's posts
	postsKey, err := ctx.GetStub().CreateCompositeKey("posts", []string{publicKey})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	// Get the posts from the state
	postsBytes, err := ctx.GetStub().GetState(postsKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve posts: %v", err)
	}
	if postsBytes == nil {
		return nil, fmt.Errorf("no posts found for user: %s", publicKey)
	}

	// Unmarshal the posts into a list of IPFS hashes
	var posts []string
	err = json.Unmarshal(postsBytes, &posts)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal posts: %v", err)
	}

	return posts, nil
}


// main function starts the chaincode
func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v", err)
		return
	}

	// Start the chaincode
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v", err)
	}
}
