package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"posdigi-auth/client"
	"posdigi-auth/config"
	"posdigi-auth/dto"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shevaathalla/posdigi-microservice/shared/activitylogger"
)

// AuthService defines the authentication service interface
type AuthService interface {
	Register(email, password string, employeeData *dto.EmployeeDataRequest) (*dto.UserProfileResponse, string, error)
	Login(email, password string) (*dto.UserProfileResponse, string, error)
	ValidateToken(tokenString string) (*jwt.MapClaims, error)
}

type authService struct {
	userClient     *client.UserClient
	config         *config.Config
	activityLogger *activitylogger.Logger
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.MapClaims
}

// NewAuthService creates a new auth service
func NewAuthService(userClient *client.UserClient, cfg *config.Config, activityLogger *activitylogger.Logger) AuthService {
	return &authService{
		userClient:     userClient,
		config:         cfg,
		activityLogger: activityLogger,
	}
}

// Register creates a new user account with optional employee profile creation
func (s *authService) Register(email, password string, employeeData *dto.EmployeeDataRequest) (*dto.UserProfileResponse, string, error) {
	config.Debug("Attempting to register user: " + email)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	requestID := uuid.New().String()

	// Check if user already exists
	existingUser, err := s.userClient.GetUserByEmail(ctx, email)
	if err == nil && existingUser != nil {
		config.Warn("User already exists: " + email)

		// Log failed registration attempt
		if s.activityLogger != nil {
			metadata := &activitylogger.ActivityMetadata{
				Extra: map[string]interface{}{
					"email":        email,
					"error_reason": "user already exists",
				},
			}
			go s.activityLogger.LogUserAction(context.Background(), requestID, "", activitylogger.ActionRegister, "POST", "/auth/register", "", "", false, metadata)
		}

		return nil, "", errors.New("user already exists")
	}

	// Create user in User Service (password hashing handled by User Service)
	createUserReq := &dto.CreateUserRequest{
		Email:    email,
		Password: password,
		FullName: email,
		Role:     "user",
	}

	// Use employee full name if provided
	if employeeData != nil && employeeData.FullName != "" {
		createUserReq.FullName = employeeData.FullName
	}

	userProfile, err := s.userClient.CreateUser(ctx, createUserReq)
	if err != nil {
		config.Errorf("Error creating user: %v", err)

		// Log failed user creation
		if s.activityLogger != nil {
			metadata := &activitylogger.ActivityMetadata{
				Extra: map[string]interface{}{
					"email":        email,
					"error_reason": err.Error(),
				},
			}
			go s.activityLogger.LogUserAction(context.Background(), requestID, "", activitylogger.ActionRegister, "POST", "/auth/register", "", "", false, metadata)
		}

		return nil, "", fmt.Errorf("error creating user: %w", err)
	}

	// Create employee profile if employee data is provided
	if employeeData != nil {
		config.Info("Creating employee profile for user: " + email)

		if employeeData.HireDate == "" {
			employeeData.HireDate = time.Now().Format("2006-01-02")
		}
		if employeeData.EmploymentStatus == "" {
			employeeData.EmploymentStatus = "active"
		}

		if err := s.userClient.CreateEmployee(ctx, userProfile.ID, employeeData); err != nil {
			config.Errorf("Error creating employee profile: %v", err)
			// Rollback user creation
			_ = s.userClient.DeleteUser(ctx, userProfile.ID)

			// Log failed employee creation
			if s.activityLogger != nil {
				metadata := &activitylogger.ActivityMetadata{
					Extra: map[string]interface{}{
						"email":         email,
						"user_id":       userProfile.ID,
						"error_reason":  "employee creation failed",
						"employee_data": employeeData,
					},
				}
				go s.activityLogger.LogUserAction(context.Background(), requestID, userProfile.ID, activitylogger.ActionRegister, "POST", "/auth/register", "", "", false, metadata)
			}

			return nil, "", fmt.Errorf("error creating employee profile: %w", err)
		}

		config.Infof("Employee profile created successfully for user: %s", email)
	}

	// Generate JWT token
	token, err := s.generateToken(userProfile.ID, userProfile.Email, userProfile.Role)
	if err != nil {
		// Log failed token generation
		if s.activityLogger != nil {
			metadata := &activitylogger.ActivityMetadata{
				Extra: map[string]interface{}{
					"email":        email,
					"user_id":      userProfile.ID,
					"error_reason": "token generation failed",
				},
			}
			go s.activityLogger.LogUserAction(context.Background(), requestID, userProfile.ID, activitylogger.ActionRegister, "POST", "/auth/register", "", "", false, metadata)
		}
		return nil, "", fmt.Errorf("error generating token: %w", err)
	}

	config.Infof("User registered successfully: %s (Profile ID: %s)", email, userProfile.ID)

	// Log successful registration
	if s.activityLogger != nil {
		metadata := &activitylogger.ActivityMetadata{
			After: map[string]interface{}{
				"user_id":   userProfile.ID,
				"email":     userProfile.Email,
				"full_name": userProfile.FullName,
				"role":      userProfile.Role,
			},
			Extra: map[string]interface{}{
				"employee_created": employeeData != nil,
			},
		}
		go s.activityLogger.LogUserAction(context.Background(), requestID, userProfile.ID, activitylogger.ActionRegister, "POST", "/auth/register", "", "", true, metadata)
	}

	return userProfile, token, nil
}

// Login authenticates a user via User Service and returns JWT token
func (s *authService) Login(email, password string) (*dto.UserProfileResponse, string, error) {
	config.Debug("Login attempt for: " + email)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	requestID := uuid.New().String()

	// Authenticate via User Service
	userProfile, err := s.userClient.AuthenticateUser(ctx, email, password)
	if err != nil {
		config.Warn("Login failed for: " + email + " - " + err.Error())

		// Log failed login attempt
		if s.activityLogger != nil {
			go s.activityLogger.LogLogin(context.Background(), requestID, "", email, "", "", false, err.Error())
		}

		return nil, "", errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := s.generateToken(userProfile.ID, userProfile.Email, userProfile.Role)
	if err != nil {
		config.Errorf("Error generating token: %v", err)

		// Log failed token generation
		if s.activityLogger != nil {
			go s.activityLogger.LogLogin(context.Background(), requestID, userProfile.ID, email, "", "", false, "token generation failed")
		}

		return nil, "", fmt.Errorf("error generating token: %w", err)
	}

	config.Info("User logged in successfully: " + email)

	// Log successful login
	if s.activityLogger != nil {
		go s.activityLogger.LogLogin(context.Background(), requestID, userProfile.ID, email, "", "", true, "")
	}

	return userProfile, token, nil
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
func (s *authService) generateToken(userID, email, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		MapClaims: jwt.MapClaims{
			"user_id": userID,
			"email":   email,
			"role":    role,
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
