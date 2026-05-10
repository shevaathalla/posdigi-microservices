package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"posdigi-auth/client"
	"posdigi-auth/config"
	"posdigi-auth/dto"
	"posdigi-auth/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService defines the authentication service interface
type AuthService interface {
	Register(email, password string) (*repository.AuthUser, error)
	Login(email, password string) (*repository.AuthUser, string, error)
	ValidateToken(tokenString string) (*jwt.MapClaims, error)
}

type authService struct {
	authRepo   repository.AuthRepository
	userClient *client.UserClient
	config     *config.Config
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.MapClaims
}

// NewAuthService creates a new auth service with User client
func NewAuthService(authRepo repository.AuthRepository, userClient *client.UserClient, cfg *config.Config) AuthService {
	return &authService{
		authRepo:   authRepo,
		userClient: userClient,
		config:     cfg,
	}
}

// Register creates a new user account with HTTP communication to User Service
func (s *authService) Register(email, password string) (*repository.AuthUser, error) {
	config.Debug("Attempting to register user: " + email)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if user exists in auth database
	existingUser, err := s.authRepo.FindByEmail(email)
	if err != nil {
		config.Errorf("Database error checking existing user: %v", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	if existingUser != nil {
		config.Warn("User already exists: " + email)
		return nil, errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		config.Errorf("Error hashing password: %v", err)
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Create user in auth database (for authentication)
	if err := s.authRepo.CreateUser(email, string(hashedPassword)); err != nil {
		config.Errorf("Error creating auth user: %v", err)
		return nil, fmt.Errorf("error creating auth user: %w", err)
	}

	// Create user profile in User Service via HTTP
	createUserReq := &dto.CreateUserRequest{
		Email:    email,
		FullName: email, // Using email as full_name initially
		Role:     "user",
	}

	userProfile, err := s.userClient.CreateUser(ctx, createUserReq)
	if err != nil {
		config.Errorf("Error creating user profile: %v", err)
		// Rollback auth user creation
		return nil, fmt.Errorf("error creating user profile: %w", err)
	}

	config.Infof("User registered successfully: %s (Profile ID: %s)", email, userProfile.ID)

	// Return the created user from auth database
	user, err := s.authRepo.FindByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("error retrieving created user: %w", err)
	}

	return user, nil
}

// Login authenticates a user
func (s *authService) Login(email, password string) (*repository.AuthUser, string, error) {
	config.Debug("Login attempt for: " + email)

	user, err := s.authRepo.FindByEmail(email)
	if err != nil {
		config.Errorf("Database error finding user: %v", err)
		return nil, "", fmt.Errorf("database error: %w", err)
	}

	if user == nil {
		config.Warn("Login failed - user not found: " + email)
		return nil, "", errors.New("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		config.Warn("Login failed - invalid password for: " + email)
		return nil, "", errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		config.Errorf("Error generating token: %v", err)
		return nil, "", fmt.Errorf("error generating token: %w", err)
	}

	config.Info("User logged in successfully: " + email)
	return user, token, nil
}

// ValidateToken validates a JWT token
func (s *authService) ValidateToken(tokenString string) (*jwt.MapClaims, error) {
	config.Debug("Validating token")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		config.Warn("Token validation failed: " + err.Error())
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		config.Debug("Token validated successfully")
		return &claims, nil
	}

	config.Warn("Invalid token claims")
	return nil, errors.New("invalid token claims")
}

// generateToken generates a JWT token for a user
func (s *authService) generateToken(user *repository.AuthUser) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		MapClaims: jwt.MapClaims{
			"user_id": user.ID,
			"email":   user.Email,
			"role":    user.Role,
			"exp":     time.Now().Add(time.Duration(s.config.JWTExpiry) * time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}