package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/raufhm/fairflow/internal/domain"
	"github.com/raufhm/fairflow/pkg/crypto"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailExists        = errors.New("email already exists")
)

type AuthUseCase struct {
	userRepo     domain.UserRepository
	apiKeyRepo   domain.APIKeyRepository
	auditRepo    domain.AuditLogRepository
	jwtSecret    string
	tokenService *crypto.TokenService
}

func NewAuthUseCase(
	userRepo domain.UserRepository,
	apiKeyRepo domain.APIKeyRepository,
	auditRepo domain.AuditLogRepository,
	jwtSecret string,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:     userRepo,
		apiKeyRepo:   apiKeyRepo,
		auditRepo:    auditRepo,
		jwtSecret:    jwtSecret,
		tokenService: crypto.NewTokenService(jwtSecret),
	}
}

// Register creates a new user account
func (uc *AuthUseCase) Register(ctx context.Context, email, password, name string, role domain.UserRole) (*domain.User, string, error) {
	// Check if user exists
	existingUser, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", err
	}
	if existingUser != nil {
		return nil, "", ErrEmailExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	// Create user
	user := &domain.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         name,
		Role:         role,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}

	// Generate JWT token
	token, err := uc.tokenService.GenerateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	// Audit log
	_ = uc.auditRepo.Create(ctx, &domain.AuditLog{
		UserID:       &user.ID,
		UserName:     user.Name,
		Action:       "User registered",
		ResourceType: stringPtr("user"),
		ResourceID:   &user.ID,
		IPAddress:    "unknown",
	})

	return user, token, nil
}

// Login authenticates a user and returns a JWT token
func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", ErrInvalidCredentials
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := uc.tokenService.GenerateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	// Audit log
	_ = uc.auditRepo.Create(ctx, &domain.AuditLog{
		UserID:       &user.ID,
		UserName:     user.Name,
		Action:       "User logged in",
		ResourceType: stringPtr("user"),
		ResourceID:   &user.ID,
		IPAddress:    "unknown",
	})

	return user, token, nil
}

// UpdateUserSettings updates user name and/or password
func (uc *AuthUseCase) UpdateUserSettings(ctx context.Context, userID int64, name, currentPassword, newPassword *string) (*domain.User, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	updated := false

	if name != nil && *name != "" {
		user.Name = *name
		updated = true
	}

	if newPassword != nil && *newPassword != "" {
		if currentPassword == nil || *currentPassword == "" {
			return nil, errors.New("current password is required to set a new password")
		}

		// Verify current password
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(*currentPassword)); err != nil {
			return nil, errors.New("invalid current password")
		}

		// Hash new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*newPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = string(hashedPassword)
		updated = true
	}

	if !updated {
		return nil, errors.New("no valid fields provided for update")
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Audit log
	_ = uc.auditRepo.Create(ctx, &domain.AuditLog{
		UserID:       &user.ID,
		UserName:     user.Name,
		Action:       "User settings updated",
		ResourceType: stringPtr("user"),
		ResourceID:   &user.ID,
		IPAddress:    "unknown",
	})

	return user, nil
}

// CreateAPIKey generates a new API key for a user
func (uc *AuthUseCase) CreateAPIKey(ctx context.Context, userID int64, name string) (string, int64, error) {
	// Generate random key
	rawKey := crypto.GenerateAPIKey()
	keyHash := crypto.HashAPIKey(rawKey, uc.jwtSecret)

	apiKey := &domain.APIKey{
		UserID:  userID,
		Name:    name,
		KeyHash: keyHash,
		Active:  true,
	}

	if err := uc.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return "", 0, err
	}

	// Audit log
	user, _ := uc.userRepo.GetByID(ctx, userID)
	if user != nil {
		_ = uc.auditRepo.Create(ctx, &domain.AuditLog{
			UserID:       &userID,
			UserName:     user.Name,
			Action:       "API key created",
			ResourceType: stringPtr("api_key"),
			ResourceID:   &apiKey.ID,
			Details:      stringPtr(fmt.Sprintf(`{"name":"%s"}`, name)),
			IPAddress:    "unknown",
		})
	}

	return rawKey, apiKey.ID, nil
}

// GetAPIKeys returns all API keys for a user
func (uc *AuthUseCase) GetAPIKeys(ctx context.Context, userID int64) ([]*domain.APIKey, error) {
	return uc.apiKeyRepo.GetByUserID(ctx, userID)
}

// RevokeAPIKey deletes an API key
func (uc *AuthUseCase) RevokeAPIKey(ctx context.Context, userID, keyID int64) error {
	apiKey, err := uc.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		return err
	}
	if apiKey == nil || apiKey.UserID != userID {
		return errors.New("API key not found or does not belong to user")
	}

	if err := uc.apiKeyRepo.Delete(ctx, keyID); err != nil {
		return err
	}

	// Audit log
	user, _ := uc.userRepo.GetByID(ctx, userID)
	if user != nil {
		_ = uc.auditRepo.Create(ctx, &domain.AuditLog{
			UserID:       &userID,
			UserName:     user.Name,
			Action:       "API key revoked",
			ResourceType: stringPtr("api_key"),
			ResourceID:   &keyID,
			Details:      stringPtr(fmt.Sprintf(`{"name":"%s"}`, apiKey.Name)),
			IPAddress:    "unknown",
		})
	}

	return nil
}

// VerifyAPIKey validates an API key and returns the associated user
func (uc *AuthUseCase) VerifyAPIKey(ctx context.Context, rawKey string) (*domain.User, error) {
	keyHash := crypto.HashAPIKey(rawKey, uc.jwtSecret)
	apiKey, err := uc.apiKeyRepo.GetByHash(ctx, keyHash)
	if err != nil {
		return nil, err
	}
	if apiKey == nil || !apiKey.Active {
		return nil, errors.New("invalid or revoked API key")
	}

	user, err := uc.userRepo.GetByID(ctx, apiKey.UserID)
	if err != nil {
		return nil, err
	}

	// Update last used timestamp (async, ignore errors)
	go uc.apiKeyRepo.UpdateLastUsed(context.Background(), apiKey.ID)

	return user, nil
}

// GetUserByID retrieves a user by ID
func (uc *AuthUseCase) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}
