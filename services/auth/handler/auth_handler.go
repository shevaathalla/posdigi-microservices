package handler

import (
	"net/http"
	"strings"

	"posdigi-auth/config"
	"posdigi-auth/dto"
	"posdigi-auth/service"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new auth handler instance
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
// @Summary Register new user
// @Description Register a new user account with optional employee profile
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration details (can include employee_data)"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} dto.AuthResponse
// @Failure 409 {object} dto.AuthResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		config.Warn("Invalid request body for registration")
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	// Validate employee data if provided
	if req.EmployeeData != nil {
		if err := c.Validate(req.EmployeeData); err != nil {
			return err
		}
		config.Info("Registration includes employee profile data")
	}

	userProfile, token, err := h.authService.Register(req.Email, req.Password, req.EmployeeData)
	if err != nil {
		errMsg := err.Error()
		config.Warn("Register failed for " + req.Email + ": " + errMsg)
		switch {
		case strings.Contains(errMsg, "already exists"):
			return c.JSON(http.StatusConflict, dto.NewErrorResponse("User already exists"))
		case strings.Contains(errMsg, "service unavailable"), strings.Contains(errMsg, "upstream service"):
			return c.JSON(http.StatusServiceUnavailable, dto.NewErrorResponse("Service temporarily unavailable, please try again later"))
		default:
			return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Registration failed"))
		}
	}

	userResponse := dto.NewUserResponse(userProfile.ID, userProfile.Email, userProfile.Role)
	loginResponse := dto.LoginResponse{
		User:  userResponse,
		Token: token,
	}
	config.Info("User registered successfully: " + userProfile.Email)

	return c.JSON(http.StatusCreated, dto.NewSuccessResponse("User registered successfully", loginResponse))
}

// Login handles user login
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.AuthResponse
// @Failure 401 {object} dto.AuthResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		config.Warn("Invalid request body for login")
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	userProfile, token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		errMsg := err.Error()
		config.Warn("Login failed for " + req.Email + ": " + errMsg)
		switch {
		case strings.Contains(errMsg, "invalid credentials"):
			return c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Invalid email or password"))
		case strings.Contains(errMsg, "service unavailable"), strings.Contains(errMsg, "upstream service"):
			return c.JSON(http.StatusServiceUnavailable, dto.NewErrorResponse("Service temporarily unavailable, please try again later"))
		default:
			return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Login failed"))
		}
	}

	userResponse := dto.NewUserResponse(userProfile.ID, userProfile.Email, userProfile.Role)
	loginResponse := dto.LoginResponse{
		User:  userResponse,
		Token: token,
	}

	config.Info("User logged in successfully: " + userProfile.Email)
	return c.JSON(http.StatusOK, dto.NewSuccessResponse("Login successful", loginResponse))
}

// ValidateToken handles token validation
// @Summary Validate JWT token
// @Description Validate JWT token and return user information
// @Tags auth
// @Accept json
// @Produce json
// @Param token query string false "Token (optional if provided in body)"
// @Param request body dto.TokenRequest false "Token in request body"
// @Success 200 {object} dto.AuthResponse
// @Failure 401 {object} dto.AuthResponse
// @Router /auth/validate [post]
// @Router /auth/validate [get]
func (h *AuthHandler) ValidateToken(c echo.Context) error {
	var tokenString string

	// Try to get token from query parameter
	tokenString = c.QueryParam("token")

	// If not in query, try to get from request body
	if tokenString == "" {
		var req dto.TokenRequest
		if err := c.Bind(&req); err == nil && req.Token != "" {
			tokenString = req.Token
		}
	}

	// If still no token, try Authorization header
	if tokenString == "" {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}
	}

	if tokenString == "" {
		config.Warn("No token provided for validation")
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Token is required"))
	}

	claims, err := h.authService.ValidateToken(tokenString)
	if err != nil {
		config.Warn("Token validation failed: " + err.Error())
		return c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Invalid token"))
	}

	config.Info("Token validated successfully for user: " + (*claims)["email"].(string))
	return c.JSON(http.StatusOK, dto.NewSuccessResponse("Token is valid", *claims))
}