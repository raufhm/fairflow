package usecase_test

import (
	"context"
	"testing"

	"github.com/raufhm/fairflow/internal/domain"
	"github.com/raufhm/fairflow/internal/usecase"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister(t *testing.T) {
	userRepo := newMockUserRepo()
	apiKeyRepo := newMockAPIKeyRepo()
	auditRepo := &mockAuditRepo{}

	uc := usecase.NewAuthUseCase(userRepo, apiKeyRepo, auditRepo, "test-secret")

	// Test successful registration
	user, token, err := uc.Register(context.Background(), "john@example.com", "password123", "John Doe", domain.RoleUser)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	if user.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%s'", user.Name)
	}
	if user.Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got '%s'", user.Email)
	}
	if user.Role != domain.RoleUser {
		t.Errorf("Expected role 'user', got '%s'", user.Role)
	}
	if token == "" {
		t.Error("Expected token but got empty string")
	}

	// Verify password is hashed
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123")); err != nil {
		t.Error("Password was not properly hashed")
	}

	// Test duplicate email
	_, _, err = uc.Register(context.Background(), "john@example.com", "password456", "Jane Doe", domain.RoleUser)
	if err == nil {
		t.Error("Expected error for duplicate email")
	}
}

func TestLogin(t *testing.T) {
	userRepo := newMockUserRepo()
	apiKeyRepo := newMockAPIKeyRepo()
	auditRepo := &mockAuditRepo{}

	uc := usecase.NewAuthUseCase(userRepo, apiKeyRepo, auditRepo, "test-secret")

	// Register a user
	_, _, err := uc.Register(context.Background(), "john@example.com", "password123", "John Doe", domain.RoleUser)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Test successful login
	user, token, err := uc.Login(context.Background(), "john@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	if token == "" {
		t.Error("Expected token but got empty string")
	}
	if user.Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got '%s'", user.Email)
	}

	// Test wrong password
	_, _, err = uc.Login(context.Background(), "john@example.com", "wrongpassword")
	if err == nil {
		t.Error("Expected error for wrong password")
	}

	// Test non-existent user
	_, _, err = uc.Login(context.Background(), "nonexistent@example.com", "password")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

func TestGenerateAPIKey(t *testing.T) {
	userRepo := newMockUserRepo()
	apiKeyRepo := newMockAPIKeyRepo()
	auditRepo := &mockAuditRepo{}

	uc := usecase.NewAuthUseCase(userRepo, apiKeyRepo, auditRepo, "test-secret")

	// Create a user
	user := &domain.User{Name: "John", Email: "john@example.com", Role: domain.RoleUser}
	userRepo.Create(context.Background(), user)

	// Generate API key
	rawKey, keyID, err := uc.CreateAPIKey(context.Background(), user.ID, "Test Key")
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	if rawKey == "" {
		t.Error("Expected API key but got empty string")
	}
	if keyID == 0 {
		t.Error("Expected key ID but got 0")
	}

	// Verify the key was stored
	keys, _ := uc.GetAPIKeys(context.Background(), user.ID)
	if len(keys) != 1 {
		t.Errorf("Expected 1 API key, got %d", len(keys))
	}
}

func TestVerifyAPIKey(t *testing.T) {
	userRepo := newMockUserRepo()
	apiKeyRepo := newMockAPIKeyRepo()
	auditRepo := &mockAuditRepo{}

	uc := usecase.NewAuthUseCase(userRepo, apiKeyRepo, auditRepo, "test-secret")

	// Create user and API key
	user := &domain.User{Name: "John", Email: "john@example.com", Role: domain.RoleUser}
	userRepo.Create(context.Background(), user)

	rawKey, keyID, _ := uc.CreateAPIKey(context.Background(), user.ID, "Test Key")

	// Verify valid API key
	verifiedUser, err := uc.VerifyAPIKey(context.Background(), rawKey)
	if err != nil {
		t.Fatalf("Failed to verify API key: %v", err)
	}
	if verifiedUser.ID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, verifiedUser.ID)
	}

	// Revoke API key
	uc.RevokeAPIKey(context.Background(), user.ID, keyID)

	// Verify revoked key fails
	_, err = uc.VerifyAPIKey(context.Background(), rawKey)
	if err == nil {
		t.Error("Expected error for revoked API key")
	}

	// Verify invalid key fails
	_, err = uc.VerifyAPIKey(context.Background(), "invalid-key")
	if err == nil {
		t.Error("Expected error for invalid API key")
	}
}
