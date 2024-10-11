package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// User represents a user profile
// type User struct {
// 	Name           string `json:"name"`
// 	Email          string `json:"email"`
// 	Phone          string `json:"phone"`
// 	Password       string `json:"password"`
// 	ProfilePicture string `json:"profilePicture,omitempty"`
// 	DataCID        string `json:"dataCID"`
// }

type User struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Password       string `json:"password"`
	ProfilePicture string `json:"profilePicture"`
	DataCID        string `json:"dataCID"`
}

// Post represents a social media post
type Post struct {
	UserEmail     string            `json:"userEmail"`
	ContentCID    string            `json:"contentCID"`
	Timestamp     int64             `json:"timestamp"`
	Reactions     map[string]string `json:"reactions"`
	ReactionCount int               `json:"reactionCount"`
}

// CreateUser creates a new user profile
// func (s *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface, email string, dataCID string) error {
// 	exists, err := s.UserExists(ctx, email)
// 	if err != nil {
// 		return err
// 	}
// 	if exists {
// 		return fmt.Errorf("user already exists: %s", email)
// 	}

// 	user := User{
// 		Email:   email,
// 		DataCID: dataCID,
// 	}
// 	userJSON, err := json.Marshal(user)
// 	if err != nil {
// 		return err
// 	}

// 	return ctx.GetStub().PutState(email, userJSON)
// }

// GetAllUsers retrieves all users from the blockchain
func (s *SmartContract) GetAllUsers(ctx contractapi.TransactionContextInterface) ([]*User, error) {
	// Get all user keys
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %v", err)
	}
	defer resultsIterator.Close()

	var users []*User
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate users: %v", err)
		}

		var user User
		err = json.Unmarshal(queryResponse.Value, &user)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal user: %v", err)
		}

		// Clear sensitive data
		user.Password = ""

		users = append(users, &user)
	}

	return users, nil
}

func (s *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface, email, hashedPassword, name, ipfsHash string) error {
	exists, err := s.UserExists(ctx, email)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("user already exists: %s", email)
	}

	user := User{
		Email:          email,
		Password:       hashedPassword,
		Name:           name,
		DataCID:        ipfsHash,
		ProfilePicture: "", // Add a default empty string for ProfilePicture
	}
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	fmt.Printf("Creating user with email: %s, hashedPassword: %s\n", email, hashedPassword)
	err = ctx.GetStub().PutState(email, userJSON)
	if err != nil {
		fmt.Printf("Failed to put user state: %v\n", err)
		return err
	}
	fmt.Printf("User created successfully\n")

	return nil
}

// UserExists checks if a user exists by email
func (s *SmartContract) UserExists(ctx contractapi.TransactionContextInterface, email string) (bool, error) {
	userJSON, err := ctx.GetStub().GetState(email)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return userJSON != nil, nil
}

// GetUser retrieves a user profile by email
// func (s *SmartContract) GetUser(ctx contractapi.TransactionContextInterface, email string) (*User, error) {
// 	userJSON, err := ctx.GetStub().GetState(email)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read from world state: %v", err)
// 	}
// 	if userJSON == nil {
// 		return nil, fmt.Errorf("user does not exist: %s", email)
// 	}

// 	var user User
// 	err = json.Unmarshal(userJSON, &user)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &user, nil
// }

func (s *SmartContract) GetUser(ctx contractapi.TransactionContextInterface, email string) (*User, error) {
	userJSON, err := ctx.GetStub().GetState(email)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if userJSON == nil {
		return nil, fmt.Errorf("user does not exist: %s", email)
	}

	var user User
	err = json.Unmarshal(userJSON, &user)
	if err != nil {
		return nil, err
	}

	// Clear sensitive data before returning
	user.Password = ""

	return &user, nil
}

func (s *SmartContract) VerifyUserCredentials(ctx contractapi.TransactionContextInterface, email, hashedPassword string) (bool, error) {
	fmt.Printf("Verifying credentials for email: %s, hashedPassword: %s\n", email, hashedPassword)

	userJSON, err := ctx.GetStub().GetState(email)
	if err != nil {
		fmt.Printf("Failed to read from world state: %v\n", err)
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	if userJSON == nil {
		fmt.Printf("User not found for email: %s\n", email)
		return false, nil
	}

	var user User
	err = json.Unmarshal(userJSON, &user)
	if err != nil {
		fmt.Printf("Failed to unmarshal user data: %v\n", err)
		return false, err
	}

	fmt.Printf("Stored hashed password: %s\n", user.Password)
	fmt.Printf("Provided hashed password: %s\n", hashedPassword)

	isValid := user.Password == hashedPassword
	fmt.Printf("Credentials valid: %v\n", isValid)

	return isValid, nil
}

// CreatePost creates a new post for a user
func (s *SmartContract) CreatePost(ctx contractapi.TransactionContextInterface, email string, ipfsHash string) error {
	// Check if the user exists
	userBytes, err := ctx.GetStub().GetState(email)
	if err != nil {
		return fmt.Errorf("failed to read user data: %v", err)
	}
	if userBytes == nil {
		return fmt.Errorf("user does not exist: %s", email)
	}

	// Create a composite key for posts
	postsKey, err := ctx.GetStub().CreateCompositeKey("posts", []string{email})
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
	err = ctx.GetStub().PutState(allPostsKey, []byte(email))
	if err != nil {
		return fmt.Errorf("failed to store post in all posts: %v", err)
	}

	// Log the creation of the post
	log.Printf("Created post for user %s with IPFS hash: %s", email, ipfsHash)
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
func (s *SmartContract) AddReaction(ctx contractapi.TransactionContextInterface, postID string, userEmail string, reactionType string) error {
	post, err := s.GetPost(ctx, postID)
	if err != nil {
		return err
	}

	// Add or update reaction
	post.Reactions[userEmail] = reactionType
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
func (s *SmartContract) GetPostsByUser(ctx contractapi.TransactionContextInterface, email string) ([]string, error) {
	// Check if the user exists
	userBytes, err := ctx.GetStub().GetState(email)
	if err != nil {
		return nil, fmt.Errorf("failed to read user data: %v", err)
	}
	if userBytes == nil {
		return nil, fmt.Errorf("user does not exist: %s", email)
	}

	// Create the composite key for the user's posts
	postsKey, err := ctx.GetStub().CreateCompositeKey("posts", []string{email})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	// Get the posts from the state
	postsBytes, err := ctx.GetStub().GetState(postsKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve posts: %v", err)
	}
	if postsBytes == nil {
		return nil, fmt.Errorf("no posts found for user: %s", email)
	}

	// Unmarshal the posts into a list of IPFS hashes
	var posts []string
	err = json.Unmarshal(postsBytes, &posts)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal posts: %v", err)
	}

	return posts, nil
}

// main function starts up the chaincode
func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		fmt.Printf("Error creating social media chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting social media chaincode: %s", err.Error())
	}
}
