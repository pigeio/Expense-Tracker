package utils

import (
	"testing"
)

func TestGenerateToken(t *testing.T) {
	// 1. Arrange
	userID := uint(1)

	// 2. Act
	token, err := GenerateToken(userID)

	// 3. Assert
	if err != nil {
		t.Errorf("Error generating token: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty string")
	}
}

func TestValidateToken(t *testing.T) {
	// Create a token first
	token, _ := GenerateToken(1)

	// Verify it
	claims, err := ValidateToken(token)

	if err != nil {
		t.Errorf("Failed to validate token: %v", err)
	}

	// Check if the user_id inside matches what we put in
	id := uint(claims["user_id"].(float64))
	if id != 1 {
		t.Errorf("Expected user_id 1, got %v", id)
	}
}
