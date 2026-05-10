package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"posdigi-user/config"
	"posdigi-user/dto"
	"posdigi-user/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(req *dto.CreateUserRequest) (*repository.User, error)
	GetUserByID(userID string) (*repository.User, error)
	GetUserByEmail(email string) (*repository.User, error)
	AuthenticateUser(email, password string) (*repository.User, error)
	UpdateUser(userID string, req *dto.UpdateUserRequest) (*repository.User, error)
	DeleteUser(userID string) error
	ListUsers(req *dto.ListUsersRequest) (*dto.ListUsersResponse, error)
}

type userService struct {
	userRepo repository.UserRepository
	config   *config.Config
}

// NewUserService creates a new user service instance
func NewUserService(userRepo repository.UserRepository, cfg *config.Config) UserService {
	return &userService{
		userRepo: userRepo,
		config:   cfg,
	}
}

// CreateUser creates a new user profile with hashed password
func (s *userService) CreateUser(req *dto.CreateUserRequest) (*repository.User, error) {
	config.Debug("Creating new user: " + req.Email)

	// Check if user already exists (including soft-deleted records which still
	// hold the unique email constraint in the DB)
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		config.Errorf("Database error checking existing user: %v", err)
		return nil, fmt.Errorf("database error: %w", err)
	}
	if existingUser != nil {
		config.Warn("User already exists: " + req.Email)
		return nil, errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		config.Errorf("Error hashing password: %v", err)
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Set default role
	role := req.Role
	if role == "" {
		role = "user"
	}

	// Create new user
	user := &repository.User{
		ID:        uuid.NewString(),
		Email:     req.Email,
		Password:  string(hashedPassword),
		FullName:  req.FullName,
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		config.Errorf("Error creating user: %v", err)
		// Catch DB-level duplicate key (e.g. soft-deleted user with same email)
		if strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "duplicate key") {
			return nil, errors.New("user already exists")
		}
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	config.Info("User created successfully: " + req.Email)
	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *userService) GetUserByID(userID string) (*repository.User, error) {
	config.Debug("Getting user by ID: " + userID)

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		config.Errorf("Database error finding user: %v", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	if user == nil {
		config.Warn("User not found: " + userID)
		return nil, errors.New("user not found")
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *userService) GetUserByEmail(email string) (*repository.User, error) {
	config.Debug("Getting user by email: " + email)

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		config.Errorf("Database error finding user by email: %v", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	if user == nil {
		config.Warn("User not found by email: " + email)
		return nil, errors.New("user not found")
	}

	return user, nil
}

// AuthenticateUser validates email + password credentials
func (s *userService) AuthenticateUser(email, password string) (*repository.User, error) {
	config.Debug("Authenticating user: " + email)

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		config.Errorf("Database error finding user: %v", err)
		return nil, errors.New("invalid credentials")
	}

	if user == nil {
		config.Warn("Authentication failed - user not found: " + email)
		return nil, errors.New("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		config.Warn("Authentication failed - invalid password for: " + email)
		return nil, errors.New("invalid credentials")
	}

	config.Info("User authenticated successfully: " + email)
	return user, nil
}

// UpdateUser updates a user profile
func (s *userService) UpdateUser(userID string, req *dto.UpdateUserRequest) (*repository.User, error) {
	config.Debug("Updating user: " + userID)

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		config.Errorf("Database error finding user: %v", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	if user == nil {
		config.Warn("User not found for update: " + userID)
		return nil, errors.New("user not found")
	}

	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		config.Errorf("Error updating user: %v", err)
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	config.Info("User updated successfully: " + userID)
	return user, nil
}

// DeleteUser deletes a user
func (s *userService) DeleteUser(userID string) error {
	config.Debug("Deleting user: " + userID)

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		config.Errorf("Database error finding user: %v", err)
		return fmt.Errorf("database error: %w", err)
	}

	if user == nil {
		config.Warn("User not found for deletion: " + userID)
		return errors.New("user not found")
	}

	if err := s.userRepo.Delete(userID); err != nil {
		config.Errorf("Error deleting user: %v", err)
		return fmt.Errorf("error deleting user: %w", err)
	}

	config.Info("User deleted successfully: " + userID)
	return nil
}

// ListUsers retrieves a paginated list of users
func (s *userService) ListUsers(req *dto.ListUsersRequest) (*dto.ListUsersResponse, error) {
	config.Debug("Listing users")

	users, total, err := s.userRepo.List(req.Page, req.Limit, req.Search)
	if err != nil {
		config.Errorf("Error listing users: %v", err)
		return nil, fmt.Errorf("error listing users: %w", err)
	}

	userResponses := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		userResponses = append(userResponses, dto.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FullName:  user.FullName,
			Role:      user.Role,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
			UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &dto.ListUsersResponse{
		Users: userResponses,
		Total: total,
		Page:  req.Page,
		Limit: req.Limit,
	}, nil
}
