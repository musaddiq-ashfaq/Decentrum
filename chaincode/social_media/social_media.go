package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// User struct defines the user structure
// User represents the user structure in the application (including public/private keys)
type User struct {
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	PublicKey string `json:"publicKey"`
}

//	type Post struct {
//		UserPublicKey     string            `json:"userPublicKey"`
//		ContentCID    string            `json:"contentCID"`
//		Timestamp     int64             `json:"timestamp"`
//		Reactions     map[string]string `json:"reactions"`
//		ReactionCount int               `json:"reactionCount"`
//		ShareCount    int                `json:"shareCount"`
//	}
type Post struct {
	ID            string            `json:"id"`
	UserPublicKey string            `json:"userPublicKey"`
	ContentCID    string            `json:"contentCID"`
	Timestamp     int64             `json:"timestamp"`
	Reactions     map[string]string `json:"reactions"`
	ReactionCount int               `json:"reactionCount"`
	ShareCount    int               `json:"shareCount"`
}

// Message represents a chat message structure
// Message structure
type Message struct {
	IPFSHash  string `json:"ipfsHash"`
	Signature string `json:"signature"`
	Sender    string `json:"sender"`
	Receiver  string `json:"receiver"`
	Timestamp string `json:"timestamp"`
}
type Chat struct {
	Participants [2]string `json:"participants"` // Public keys of the two participants
	Messages     []Message `json:"messages"`     // List of messages exchanged
}

// Group represents a group structure in the blockchain
type Group struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

// SmartContract defines the chaincode structure
type SmartContract struct {
	contractapi.Contract
}

type FriendsList struct {
	Friends []string `json:"friends"`
}

type FriendRequest struct {
	Sender       string `json:"sender"`
	SenderName   string `json:"senderName"`
	Receiver     string `json:"receiver"`
	ReceiverName string `json:"receiverName"`
	Status       string `json:"status"` // "pending", "accepted", "rejected"
	Timestamp    int64  `json:"timestamp"`
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
		Name:      name,
		Phone:     phone,
		PublicKey: publicKey,
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

func (s *SmartContract) CreatePost(ctx contractapi.TransactionContextInterface, publicKey string, ipfsHash string, postID string) error {
	// Check if the user exists
	userBytes, err := ctx.GetStub().GetState(publicKey)
	if err != nil {
		return fmt.Errorf("failed to read user data: %v", err)
	}
	if userBytes == nil {
		return fmt.Errorf("user does not exist: %s", publicKey)
	}

	// Create a new Post struct
	post := Post{
		ID:            postID,
		UserPublicKey: publicKey,
		ContentCID:    ipfsHash,
		Timestamp:     time.Now().Unix(),
		Reactions:     make(map[string]string),
		ReactionCount: 0,
		ShareCount:    0,
	}

	// Serialize the post
	postJSON, err := json.Marshal(post)
	if err != nil {
		return fmt.Errorf("failed to marshal post: %v", err)
	}

	// Store the post using the IPFS hash as the key
	err = ctx.GetStub().PutState(ipfsHash, postJSON)
	if err != nil {
		return fmt.Errorf("failed to store post: %v", err)
	}

	// Create a composite key for posts
	postsKey, err := ctx.GetStub().CreateCompositeKey("posts", []string{publicKey})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	log.Printf("Generated posts composite key: %s", postsKey)

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

	log.Printf("Generated allposts composite key: %s", allPostsKey)
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
	log.Printf("Fetching post with ID: %s", postID)

	postJSON, err := ctx.GetStub().GetState(postID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if postJSON == nil {
		return nil, fmt.Errorf("post does not exist: %s", postID)
	}
	log.Printf("Retrieved post data for ID %s: %s", postID, string(postJSON))
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

func (s *SmartContract) AddReaction(ctx contractapi.TransactionContextInterface, postID string, userPublicKey string, reactionType string) (*Post, error) {
	// Retrieve the existing post state directly using the postID (IPFS hash)
	existingPostJSON, err := ctx.GetStub().GetState(postID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve post state for postID '%s': %v", postID, err)
	}

	// Check if the post exists
	if existingPostJSON == nil {
		return nil, fmt.Errorf("post with postID '%s' does not exist", postID)
	}

	// Deserialize the post JSON into the Post struct
	var post Post
	err = json.Unmarshal(existingPostJSON, &post)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal post data for postID '%s': %v", postID, err)
	}

	// Initialize reactions map if nil
	if post.Reactions == nil {
		post.Reactions = make(map[string]string)
	}

	// Add or update the user's reaction
	post.Reactions[userPublicKey] = reactionType

	// Update reaction count
	post.ReactionCount = len(post.Reactions)

	// Serialize the updated post
	updatedPostJSON, err := json.Marshal(post)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated post for postID '%s': %v", postID, err)
	}

	// Save the updated post state
	err = ctx.GetStub().PutState(postID, updatedPostJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to update post state for postID '%s': %v", postID, err)
	}

	// Return the updated post
	return &post, nil
}
func (s *SmartContract) GetAllUserPosts(ctx contractapi.TransactionContextInterface) (map[string][]string, error) {
	// Create a map to store all posts, keyed by the user's public key
	allPosts := make(map[string][]string)

	// Query for all posts using the composite key prefix
	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("posts", []string{})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve posts: %v", err)
	}
	defer resultsIterator.Close()

	// Iterate through the results and collect posts for each user
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate through posts: %v", err)
		}

		// Extract the public key from the composite key
		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(queryResponse.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to split composite key: %v", err)
		}
		if len(compositeKeyParts) == 0 {
			continue // Skip if the key doesn't contain a public key
		}
		publicKey := compositeKeyParts[0]

		// Unmarshal the posts into a list of IPFS hashes
		var posts []string
		err = json.Unmarshal(queryResponse.Value, &posts)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal posts for user %s: %v", publicKey, err)
		}

		// Store the posts in the map
		allPosts[publicKey] = posts
	}

	return allPosts, nil
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

// QueryUserByName retrieves a user by their name
func (s *SmartContract) QueryUserByName(ctx contractapi.TransactionContextInterface, name string) (*User, error) {
	// Get all keys from the ledger
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve states from ledger: %v", err)
	}
	defer resultsIterator.Close()

	// Iterate through all ledger entries
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("error iterating through ledger states: %v", err)
		}

		// Parse the user JSON
		var user User
		if err := json.Unmarshal(queryResponse.Value, &user); err != nil {
			continue // Ignore invalid entries
		}

		// Check if the name matches
		if user.Name == name {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("user with name %s not found", name)
}

func (s *SmartContract) AddMessage(ctx contractapi.TransactionContextInterface, chatID string, message string, senderPublicKey string, receiverPublicKey string) error {
	// Retrieve existing chat data
	chatData, err := ctx.GetStub().GetState(chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %v", err)
	}

	var chat Chat
	if len(chatData) == 0 {
		// Create a new chat if it doesn't exist

		chat = Chat{
			Participants: [2]string{senderPublicKey, receiverPublicKey}, // Add sender and receiver to participants
			Messages:     []Message{},
		}
	} else {
		log.Printf("i am here")
		// Unmarshal existing chat data
		err = json.Unmarshal(chatData, &chat)
		if err != nil {
			return fmt.Errorf("failed to unmarshal chat data: %v", err)
		}
	}

	// Unmarshal the new message
	var newMessage Message
	err = json.Unmarshal([]byte(message), &newMessage)
	if err != nil {
		return fmt.Errorf("failed to unmarshal message data: %v", err)
	}

	// Append the new message
	chat.Messages = append(chat.Messages, newMessage)

	// Update the state
	chatBytes, err := json.Marshal(chat)
	if err != nil {
		return fmt.Errorf("failed to marshal chat: %v", err)
	}

	// Store the updated chat data in the blockchain state
	err = ctx.GetStub().PutState(chatID, chatBytes)
	if err != nil {
		return fmt.Errorf("failed to store updated chat: %v", err)
	}

	// Fire an event to notify the client
	eventPayload := fmt.Sprintf("Message added to chat %s", chatID)
	err = ctx.GetStub().SetEvent("MessageAddedEvent", []byte(eventPayload))
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	return nil
}

// GetChat retrieves a chat with all its messages using the chatID
func (s *SmartContract) GetChat(ctx contractapi.TransactionContextInterface, chatID string) (*Chat, error) {
	// Retrieve the chat data from the state
	chatData, err := ctx.GetStub().GetState(chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %v", err)
	}

	if len(chatData) == 0 {
		return nil, fmt.Errorf("chat with ID %s not found", chatID)
	}

	var chat Chat
	err = json.Unmarshal(chatData, &chat)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal chat data: %v", err)
	}

	return &chat, nil
}

func (s *SmartContract) GetAllUsers(ctx contractapi.TransactionContextInterface) ([]*User, error) {
	// Range query with empty string for startKey and endKey does a full scan
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get state iterator: %v", err)
	}
	defer resultsIterator.Close()

	var users []*User
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next query result: %v", err)
		}

		var user User
		err = json.Unmarshal(queryResponse.Value, &user)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal user: %v", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

func (s *SmartContract) CreateGroup(ctx contractapi.TransactionContextInterface, id string, name string, members []string) error {
	// Check if the group already exists
	exists, err := s.GroupExists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check if group exists: %v", err)
	}
	if exists {
		return fmt.Errorf("group with ID %s already exists", id)
	}

	// Create the group object
	group := Group{
		ID:      id,
		Name:    name,
		Members: members,
	}

	// Serialize the group to JSON
	groupJSON, err := json.Marshal(group)
	if err != nil {
		return fmt.Errorf("failed to serialize group: %v", err)
	}

	// Store the group on the blockchain
	err = ctx.GetStub().PutState(id, groupJSON)
	if err != nil {
		return fmt.Errorf("failed to put group state: %v", err)
	}

	return nil
}

// ReadGroup retrieves a group from the blockchain by its ID
func (s *SmartContract) ReadGroup(ctx contractapi.TransactionContextInterface, id string) (*Group, error) {
	// Get the group JSON from the blockchain
	groupJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read group: %v", err)
	}
	if groupJSON == nil {
		return nil, fmt.Errorf("group with ID %s does not exist", id)
	}

	// Deserialize the group JSON
	var group Group
	err = json.Unmarshal(groupJSON, &group)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize group: %v", err)
	}

	return &group, nil
}

// GroupExists checks if a group exists on the blockchain
func (s *SmartContract) GroupExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	groupJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read group: %v", err)
	}

	return groupJSON != nil, nil
}

// SendFriendRequest creates a new friend request
func (s *SmartContract) SendFriendRequest(ctx contractapi.TransactionContextInterface, sender string, receiver string) (string, error) {
	// Validate that both sender and receiver exist
	senderExists, err := s.UserExists(ctx, sender)
	if err != nil || !senderExists {
		return "", fmt.Errorf("sender user does not exist: %s", sender)
	}

	receiverExists, err := s.UserExists(ctx, receiver)
	if err != nil || !receiverExists {
		return "", fmt.Errorf("receiver user does not exist: %s", receiver)
	}

	// Check if a friend request already exists
	existingRequest, err := s.GetFriendRequest(ctx, sender, receiver)
	if err == nil && existingRequest != nil {
		return "", fmt.Errorf("friend request already exists with status: %s", existingRequest.Status)
	}

	// Create a new friend request
	friendRequest := FriendRequest{
		Sender:    sender,
		Receiver:  receiver,
		Status:    "pending",
		Timestamp: time.Now().Unix(),
	}

	// Generate a unique key for the friend request
	requestKey, err := ctx.GetStub().CreateCompositeKey("friendrequest", []string{sender, receiver})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %v", err)
	}

	// Serialize the friend request
	requestJSON, err := json.Marshal(friendRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal friend request: %v", err)
	}

	// Store the friend request on the ledger
	err = ctx.GetStub().PutState(requestKey, requestJSON)
	if err != nil {
		return "", fmt.Errorf("failed to store friend request: %v", err)
	}

	// Return the request key as the result
	return requestKey, nil
}

// GetFriendRequest retrieves a specific friend request
func (s *SmartContract) GetFriendRequest(ctx contractapi.TransactionContextInterface, sender string, receiver string) (*FriendRequest, error) {
	// Create the composite key for the friend request
	requestKey, err := ctx.GetStub().CreateCompositeKey("friendrequest", []string{sender, receiver})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	// Retrieve the friend request from the ledger
	requestJSON, err := ctx.GetStub().GetState(requestKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read friend request: %v", err)
	}
	if requestJSON == nil {
		return nil, fmt.Errorf("friend request not found")
	}

	// Deserialize the friend request
	var friendRequest FriendRequest
	err = json.Unmarshal(requestJSON, &friendRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal friend request: %v", err)
	}

	return &friendRequest, nil
}

// RespondToFriendRequest allows a user to accept or reject a friend request
func (s *SmartContract) RespondToFriendRequest(ctx contractapi.TransactionContextInterface, sender string, receiver string, response string) error {
	// Validate response
	if response != "accepted" && response != "rejected" {
		return fmt.Errorf("invalid response. Must be 'accepted' or 'rejected'")
	}

	// Retrieve the existing friend request
	existingRequest, err := s.GetFriendRequest(ctx, sender, receiver)
	if err != nil {
		return fmt.Errorf("friend request not found: %v", err)
	}

	// Check if the request is still pending
	if existingRequest.Status != "pending" {
		return fmt.Errorf("friend request has already been %s", existingRequest.Status)
	}

	// Update the friend request status
	existingRequest.Status = response

	// Recreate the composite key
	requestKey, err := ctx.GetStub().CreateCompositeKey("friendrequest", []string{sender, receiver})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// Serialize the updated friend request
	updatedRequestJSON, err := json.Marshal(existingRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal updated friend request: %v", err)
	}

	// Store the updated friend request
	err = ctx.GetStub().PutState(requestKey, updatedRequestJSON)
	if err != nil {
		return fmt.Errorf("failed to update friend request: %v", err)
	}

	// If accepted, add to friends list (you would need to implement this separately)
	if response == "accepted" {
		err = s.addFriend(ctx, sender, receiver)
		if err != nil {
			return fmt.Errorf("failed to add friend: %v", err)
		}
	}

	return nil
}

func (s *SmartContract) GetFriendRequestsByUser(ctx contractapi.TransactionContextInterface, publicKey string) (string, error) {
	// Create an iterator for friend requests
	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("friendrequest", []string{})
	if err != nil {
		return "", fmt.Errorf("failed to get iterator for friend requests: %v", err)
	}
	defer resultsIterator.Close()

	var userFriendRequests []*FriendRequest
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return "", fmt.Errorf("failed to get next item from iterator: %v", err)
		}

		// Split the composite key to extract sender and receiver
		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(queryResponse.Key)
		if err != nil {
			return "", fmt.Errorf("failed to split composite key: %v", err)
		}

		// Check if the current user is sender or receiver
		if len(compositeKeyParts) == 2 && (compositeKeyParts[0] == publicKey || compositeKeyParts[1] == publicKey) {
			var friendRequest FriendRequest
			err = json.Unmarshal(queryResponse.Value, &friendRequest)
			if err != nil {
				return "", fmt.Errorf("failed to unmarshal friend request: %v", err)
			}

			// Fetch sender's name based on the sender public key
			senderName, err := s.getUserName(ctx, friendRequest.Sender)
			if err != nil {
				return "", fmt.Errorf("failed to fetch sender name: %v", err)
			}

			// Fetch receiver's name based on the receiver public key
			receiverName, err := s.getUserName(ctx, friendRequest.Receiver)
			if err != nil {
				return "", fmt.Errorf("failed to fetch receiver name: %v", err)
			}

			// Add sender and receiver names to the request
			friendRequest.SenderName = senderName
			log.Printf(friendRequest.SenderName)
			friendRequest.ReceiverName = receiverName

			userFriendRequests = append(userFriendRequests, &friendRequest)
		}
	}

	// Encode the friend requests to base64
	friendRequestsJSON, err := json.Marshal(userFriendRequests)
	if err != nil {
		return "", fmt.Errorf("failed to marshal friend requests: %v", err)
	}
	return base64.StdEncoding.EncodeToString(friendRequestsJSON), nil
}

// Helper function to fetch user name by public key
func (s *SmartContract) getUserName(ctx contractapi.TransactionContextInterface, publicKey string) (string, error) {
	// Query the user details from the ledger using the public key
	userBytes, err := ctx.GetStub().GetState(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to get user data: %v", err)
	}
	if userBytes == nil {
		return "", fmt.Errorf("user not found")
	}

	// Assuming the user data is a JSON object with a 'name' field
	var user User
	err = json.Unmarshal(userBytes, &user)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal user data: %v", err)
	}

	return user.Name, nil
}

func (s *SmartContract) addFriend(ctx contractapi.TransactionContextInterface, user1 string, user2 string) error {
	// Retrieve the friends list of user1
	user1FriendsKey := fmt.Sprintf("friends_%s", user1)
	user1FriendsJSON, err := ctx.GetStub().GetState(user1FriendsKey)
	if err != nil {
		return fmt.Errorf("failed to get friends list for user1: %v", err)
	}

	user1Friends := &FriendsList{}
	if user1FriendsJSON != nil {
		err = json.Unmarshal(user1FriendsJSON, user1Friends)
		if err != nil {
			return fmt.Errorf("failed to unmarshal friends list for user1: %v", err)
		}
	}

	// Check if user2 is already a friend of user1
	for _, friend := range user1Friends.Friends {
		if friend == user2 {
			return fmt.Errorf("user %s is already a friend of user %s", user2, user1)
		}
	}

	// Add user2 to user1's friends list
	user1Friends.Friends = append(user1Friends.Friends, user2)

	// Serialize and save the updated friends list for user1
	updatedUser1FriendsJSON, err := json.Marshal(user1Friends)
	if err != nil {
		return fmt.Errorf("failed to marshal updated friends list for user1: %v", err)
	}

	err = ctx.GetStub().PutState(user1FriendsKey, updatedUser1FriendsJSON)
	if err != nil {
		return fmt.Errorf("failed to save updated friends list for user1: %v", err)
	}

	// Repeat the same steps for user2
	user2FriendsKey := fmt.Sprintf("friends_%s", user2)
	user2FriendsJSON, err := ctx.GetStub().GetState(user2FriendsKey)
	if err != nil {
		return fmt.Errorf("failed to get friends list for user2: %v", err)
	}

	user2Friends := &FriendsList{}
	if user2FriendsJSON != nil {
		err = json.Unmarshal(user2FriendsJSON, user2Friends)
		if err != nil {
			return fmt.Errorf("failed to unmarshal friends list for user2: %v", err)
		}
	}

	for _, friend := range user2Friends.Friends {
		if friend == user1 {
			return fmt.Errorf("user %s is already a friend of user %s", user1, user2)
		}
	}

	user2Friends.Friends = append(user2Friends.Friends, user1)

	updatedUser2FriendsJSON, err := json.Marshal(user2Friends)
	if err != nil {
		return fmt.Errorf("failed to marshal updated friends list for user2: %v", err)
	}

	err = ctx.GetStub().PutState(user2FriendsKey, updatedUser2FriendsJSON)
	if err != nil {
		return fmt.Errorf("failed to save updated friends list for user2: %v", err)
	}

	return nil
}

func (s *SmartContract) GetFriendsByUser(ctx contractapi.TransactionContextInterface, publicKey string) ([]string, error) {
	// Retrieve the friends list key for the given user
	friendsKey := fmt.Sprintf("friends_%s", publicKey)
	friendsListJSON, err := ctx.GetStub().GetState(friendsKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve friends list: %v", err)
	}

	// If no friends list is found, return an empty list
	if friendsListJSON == nil {
		return []string{}, nil
	}

	// Unmarshal the friends list JSON
	var friendsList FriendsList
	err = json.Unmarshal(friendsListJSON, &friendsList)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal friends list: %v", err)
	}

	return friendsList.Friends, nil
}

func (s *SmartContract) GetFriendsWithDetailsByUser(ctx contractapi.TransactionContextInterface, publicKey string) (string, error) {
	// Retrieve the friends list for the given user
	friendsKey := fmt.Sprintf("friends_%s", publicKey)
	friendsListJSON, err := ctx.GetStub().GetState(friendsKey)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve friends list: %v", err)
	}

	// If no friends list is found, return an empty list
	if friendsListJSON == nil {
		return "[]", nil
	}

	// Unmarshal the friends list JSON
	var friendsList FriendsList
	err = json.Unmarshal(friendsListJSON, &friendsList)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal friends list: %v", err)
	}

	// Fetch details (names) of each friend
	var friendsDetails []map[string]string
	for _, friendKey := range friendsList.Friends {
		// Fetch the friend's user data from the ledger
		userBytes, err := ctx.GetStub().GetState(friendKey)
		if err != nil {
			return "", fmt.Errorf("failed to fetch friend data: %v", err)
		}
		if userBytes == nil {
			// If no data is found for the friend, skip this entry
			continue
		}

		// Parse user data
		var user User
		err = json.Unmarshal(userBytes, &user)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal friend data: %v", err)
		}

		// Add friend's details to the list
		friendDetails := map[string]string{
			"id":   friendKey,
			"name": user.Name,
		}
		friendsDetails = append(friendsDetails, friendDetails)
	}

	// Marshal the friends details into JSON
	friendsDetailsJSON, err := json.Marshal(friendsDetails)
	if err != nil {
		return "", fmt.Errorf("failed to marshal friends details: %v", err)
	}

	return string(friendsDetailsJSON), nil
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
