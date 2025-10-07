package usecase_test

import (
	"testing"

	"github.com/raufhm/rra/internal/domain"
	"github.com/raufhm/rra/internal/usecase"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	users map[int64]*domain.User
}

func (m *mockUserRepo) Create(user *domain.User) error {
	user.ID = int64(len(m.users) + 1)
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetByID(id int64) (*domain.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *mockUserRepo) GetByEmail(email string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) GetAll() ([]*domain.User, error) {
	var users []*domain.User
	for _, u := range m.users {
		users = append(users, u)
	}
	return users, nil
}

func (m *mockUserRepo) Update(user *domain.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) Delete(id int64) error {
	delete(m.users, id)
	return nil
}

func (m *mockUserRepo) UpdateRole(id int64, role domain.UserRole) error {
	if u, ok := m.users[id]; ok {
		u.Role = role
	}
	return nil
}

type mockAPIKeyRepo struct {
	keys map[int64]*domain.APIKey
}

func (m *mockAPIKeyRepo) Create(key *domain.APIKey) error {
	key.ID = int64(len(m.keys) + 1)
	m.keys[key.ID] = key
	return nil
}

func (m *mockAPIKeyRepo) GetByUserID(userID int64) ([]*domain.APIKey, error) {
	var keys []*domain.APIKey
	for _, k := range m.keys {
		if k.UserID == userID {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func (m *mockAPIKeyRepo) GetByHash(hash string) (*domain.APIKey, error) {
	for _, k := range m.keys {
		if k.KeyHash == hash {
			return k, nil
		}
	}
	return nil, nil
}

func (m *mockAPIKeyRepo) Delete(id int64) error {
	delete(m.keys, id)
	return nil
}

func (m *mockAPIKeyRepo) UpdateLastUsed(id int64) error {
	return nil
}

func (m *mockAPIKeyRepo) GetByID(id int64) (*domain.APIKey, error) {
	if k, ok := m.keys[id]; ok {
		return k, nil
	}
	return nil, nil
}

func TestRegister(t *testing.T) {
	userRepo := &mockUserRepo{users: make(map[int64]*domain.User)}
	apiKeyRepo := &mockAPIKeyRepo{keys: make(map[int64]*domain.APIKey)}
	auditRepo := &mockAuditRepo{}

	uc := usecase.NewAuthUseCase(userRepo, apiKeyRepo, auditRepo, "test-secret")

	// Test successful registration
	user, token, err := uc.Register("john@example.com", "password123", "John Doe", domain.RoleUser)
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
	_, _, err = uc.Register("john@example.com", "password456", "Jane Doe", domain.RoleUser)
	if err == nil {
		t.Error("Expected error for duplicate email")
	}
}

func TestLogin(t *testing.T) {
	userRepo := &mockUserRepo{users: make(map[int64]*domain.User)}
	apiKeyRepo := &mockAPIKeyRepo{keys: make(map[int64]*domain.APIKey)}
	auditRepo := &mockAuditRepo{}

	uc := usecase.NewAuthUseCase(userRepo, apiKeyRepo, auditRepo, "test-secret")

	// Register a user
	_, _, err := uc.Register("john@example.com", "password123", "John Doe", domain.RoleUser)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Test successful login
	user, token, err := uc.Login("john@example.com", "password123")
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
	_, _, err = uc.Login("john@example.com", "wrongpassword")
	if err == nil {
		t.Error("Expected error for wrong password")
	}

	// Test non-existent user
	_, _, err = uc.Login("nonexistent@example.com", "password")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

func TestGenerateAPIKey(t *testing.T) {
	userRepo := &mockUserRepo{users: make(map[int64]*domain.User)}
	apiKeyRepo := &mockAPIKeyRepo{keys: make(map[int64]*domain.APIKey)}
	auditRepo := &mockAuditRepo{}

	uc := usecase.NewAuthUseCase(userRepo, apiKeyRepo, auditRepo, "test-secret")

	// Create a user
	user := &domain.User{Name: "John", Email: "john@example.com", Role: domain.RoleUser}
	userRepo.Create(user)

	// Generate API key
	rawKey, keyID, err := uc.CreateAPIKey(user.ID, "Test Key")
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
	keys, _ := uc.GetAPIKeys(user.ID)
	if len(keys) != 1 {
		t.Errorf("Expected 1 API key, got %d", len(keys))
	}
}

func TestVerifyAPIKey(t *testing.T) {
	userRepo := &mockUserRepo{users: make(map[int64]*domain.User)}
	apiKeyRepo := &mockAPIKeyRepo{keys: make(map[int64]*domain.APIKey)}
	auditRepo := &mockAuditRepo{}

	uc := usecase.NewAuthUseCase(userRepo, apiKeyRepo, auditRepo, "test-secret")

	// Create user and API key
	user := &domain.User{Name: "John", Email: "john@example.com", Role: domain.RoleUser}
	userRepo.Create(user)

	rawKey, keyID, _ := uc.CreateAPIKey(user.ID, "Test Key")

	// Verify valid API key
	verifiedUser, err := uc.VerifyAPIKey(rawKey)
	if err != nil {
		t.Fatalf("Failed to verify API key: %v", err)
	}
	if verifiedUser.ID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, verifiedUser.ID)
	}

	// Revoke API key
	uc.RevokeAPIKey(user.ID, keyID)

	// Verify revoked key fails
	_, err = uc.VerifyAPIKey(rawKey)
	if err == nil {
		t.Error("Expected error for revoked API key")
	}

	// Verify invalid key fails
	_, err = uc.VerifyAPIKey("invalid-key")
	if err == nil {
		t.Error("Expected error for invalid API key")
	}
}
