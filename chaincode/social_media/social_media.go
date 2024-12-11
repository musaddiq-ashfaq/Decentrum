package main

import (
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
	IPFSHash   string `json:"ipfsHash"`
	Signature  string `json:"signature"`
	Sender     string `json:"sender"`
	Receiver   string `json:"receiver"`
	Timestamp  string `json:"timestamp"`
}
type Chat struct {
    Participants [2]string `json:"participants"` // Public keys of the two participants
    Messages     []Message `json:"messages"`     // List of messages exchanged
}

// Group represents a group structure in the blockchain
type Group struct {
	ID      string   `json:"id"`
	GroupName    string   `json:"groupname"`
	Members []string `json:"members"`
}

type GroupMessage struct {
	IPFSHash   string `json:"ipfsHash"`
	Signature  string `json:"signature"`
	Sender     string `json:"sender"`
	Receiver[]   string `json:"receiver"`
	Timestamp  string `json:"timestamp"`
}
type GroupChat struct {
    Participants []string `json:"participants"` // Public keys of the two participants
    GroupMessages  []GroupMessage `json:"groupmessages"`     // List of messages exchanged
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

// CreateGroup creates a new group with the given name and members
func (s *SmartContract) CreateGroup(ctx contractapi.TransactionContextInterface, id string, groupname string, members []string) error {
    // Check if the group already exists
    exists, err := s.GroupExists(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to check if group exists: %v", err)
    }
    if exists {
        return fmt.Errorf("group with ID %s already exists", id)
    }

    // Validate input
    if len(members) == 0 {
        return fmt.Errorf("group must have at least one member")
    }

    // Validate user names
    for _, userName := range members {
        if userName == "" {
            return fmt.Errorf("empty user name is not allowed")
        }
        // Optional: Add additional validation for user names
    }

    // Create the group object
    group := Group{
        ID:      id,
        GroupName:    groupname,
        Members: members, // Now expects user names
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

// AddMemberToGroup adds a new user's name to an existing group
func (s *SmartContract) AddMemberToGroup(ctx contractapi.TransactionContextInterface, id string, userName string) error {
    // Retrieve the existing group
    group, err := s.ReadGroup(ctx, id)
    if err != nil {
        return err
    }

    // Check if the user name already exists
    for _, existingMember := range group.Members {
        if existingMember == userName {
            return fmt.Errorf("user name already exists in the group")
        }
    }

    // Add the new user name
    group.Members = append(group.Members, userName)

    // Serialize the updated group
    groupJSON, err := json.Marshal(group)
    if err != nil {
        return fmt.Errorf("failed to serialize updated group: %v", err)
    }

    // Update the group on the blockchain
    err = ctx.GetStub().PutState(id, groupJSON)
    if err != nil {
        return fmt.Errorf("failed to update group: %v", err)
    }

    return nil
}

// GetAllGroups retrieves all groups from the ledger
func (s *SmartContract) GetAllGroups(ctx contractapi.TransactionContextInterface) ([]*Group, error) {
    // Initialize a slice to store all groups
    var allGroups []*Group

    // Fetch all records from the ledger
    resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
    if err != nil {
        return nil, fmt.Errorf("failed to get state by range: %v", err)
    }
    defer resultsIterator.Close()

    // Iterate through all records
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, fmt.Errorf("failed to iterate query results: %v", err)
        }

        // Deserialize each record into a Group struct
        var group Group
        err = json.Unmarshal(queryResponse.Value, &group)
        if err != nil {
            return nil, fmt.Errorf("failed to deserialize group: %v", err)
        }

        // Skip empty or null groups (you can adjust the condition as needed)
        if group.ID == "" || len(group.Members) == 0 {
            continue
        }

        // Add the group to the result
        allGroups = append(allGroups, &group)
    }

    // Log the retrieved groups as JSON for better readability
    jsonGroups, err := json.Marshal(allGroups)
    if err != nil {
        log.Printf("Failed to marshal groups: %v", err)
    } else {
        log.Printf("Retrieved groups: %s", string(jsonGroups))
    }

    return allGroups, nil
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
