package tokenstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

var (
	TokenNotFound = errors.New("Token Not Found!")
)

type Token struct {
	Profile      string `json:"profile"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpireAt     int64  `json:"expire_in"` // Changed to int64 for easier time comparisons
}

type TokenStore struct {
	filePath string
}

// New initializes a new TokenStore
func New(path string) *TokenStore {
	// Ensure the directory exists
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		fmt.Println("Error creating directory:", err)
		return nil
	}
	filePath := fmt.Sprintf("%s/tokens.json", path)
	// Ensure the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create an empty file if it doesn't exist
		if err := os.WriteFile(filePath, []byte("[]"), 0644); err != nil {
			fmt.Println("Error creating file:", err)
			return nil
		}
	}
	return &TokenStore{
		filePath: filePath,
	}
}

// AddToken adds or updates a token for the given profile
func (s *TokenStore) AddToken(token *Token) error {

	// Read existing tokens
	tokens, err := s.readTokens()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read tokens: %w", err)
	}

	// Check if the profile already exists and update it
	found := false
	for i, t := range tokens {
		if t.Profile == token.Profile {
			tokens[i] = *token
			found = true
			break
		}
	}

	// If not found, add the new token
	if !found {
		tokens = append(tokens, *token)
	}

	// Write back to the file
	return s.writeTokens(tokens)
}

// RemoveToken removes a token for a specific profile
func (s *TokenStore) RemoveToken(profile string) error {

	// Read existing tokens
	tokens, err := s.readTokens()
	if err != nil {
		return fmt.Errorf("failed to read tokens: %w", err)
	}

	// Filter out the profile to remove
	var updatedTokens []Token
	for _, t := range tokens {
		if t.Profile != profile {
			updatedTokens = append(updatedTokens, t)
		}
	}

	// Write back the updated tokens to the file
	return s.writeTokens(updatedTokens)
}

// ListTokens lists all tokens, optionally removing expired ones
func (s *TokenStore) ListTokens() ([]Token, error) {
	// Read existing tokens
	tokens, err := s.readTokens()
	if err != nil {
		return nil, fmt.Errorf("failed to read tokens: %w", err)
	}

	return tokens, nil
}

func (s *TokenStore) GetToken(profile string) (*Token, error) {
	tokens, err := s.ListTokens()
	if err != nil {
		return nil, err
	}
	profileTokenIndex := 0
	found := false
	for id, token := range tokens {
		if token.Profile == profile {
			profileTokenIndex = id
			found = true
			break
		}
	}
	if !found {
		return nil, nil
	}

	return &tokens[profileTokenIndex], nil
}

// Helper function to read tokens from the file
func (s *TokenStore) readTokens() ([]Token, error) {
	// Read the file content
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}

	// Parse JSON into a slice of Token objects
	var tokens []Token
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return tokens, nil
}

// Helper function to write tokens to the file
func (s *TokenStore) writeTokens(tokens []Token) error {
	// Marshal the tokens into JSON
	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write the JSON data to the file
	return os.WriteFile(s.filePath, data, 0644)
}
