// chaincode/social_media/social_media.go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type User struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Password       string `json:"password"`
	ProfilePicture string `json:"profilePicture,omitempty"`
	DataCID        string `json:"dataCID"`
}

type Post struct {
	ID            int               `json:"id"`
	UserEmail     string            `json:"userEmail"`
	ContentCID    string            `json:"contentCID"`
	Reactions     map[string]string `json:"reactions,omitempty"`
	ReactionCount int               `json:"reactionCount"`
}

func (s *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface, email string, dataCID string) error {
	exists, err := s.UserExists(ctx, email)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("user already exists: %s", email)
	}

	user := User{
		Email:   email,
		DataCID: dataCID,
	}
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(email, userJSON)
}

func (s *SmartContract) UserExists(ctx contractapi.TransactionContextInterface, email string) (bool, error) {
	userJSON, err := ctx.GetStub().GetState(email)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return userJSON != nil, nil
}

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
	return &user, nil
}

func (s *SmartContract) CreatePost(ctx contractapi.TransactionContextInterface, postID int, userEmail string, contentCID string) error {
	exists, err := s.UserExists(ctx, userEmail)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("user does not exist: %s", userEmail)
	}

	post := Post{
		ID:            postID,
		UserEmail:     userEmail,
		ContentCID:    contentCID,
		Reactions:     make(map[string]string),
		ReactionCount: 0,
	}
	postJSON, err := json.Marshal(post)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(fmt.Sprintf("POST_%d", postID), postJSON)
}

func (s *SmartContract) GetPost(ctx contractapi.TransactionContextInterface, postID int) (*Post, error) {
	postJSON, err := ctx.GetStub().GetState(fmt.Sprintf("POST_%d", postID))
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if postJSON == nil {
		return nil, fmt.Errorf("post does not exist: %d", postID)
	}

	var post Post
	err = json.Unmarshal(postJSON, &post)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (s *SmartContract) AddReaction(ctx contractapi.TransactionContextInterface, postID int, userEmail string, reactionType string) error {
	post, err := s.GetPost(ctx, postID)
	if err != nil {
		return err
	}

	post.Reactions[userEmail] = reactionType
	post.ReactionCount = len(post.Reactions)

	postJSON, err := json.Marshal(post)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(fmt.Sprintf("POST_%d", postID), postJSON)
}

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