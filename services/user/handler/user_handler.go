package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"posdigi-user/config"
	"posdigi-user/dto"
	"posdigi-user/service"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new user handler instance
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser handles user creation
// @Summary Create new user
// @Description Create a new user profile
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.CreateUserRequest true "User details"
// @Success 201 {object} dto.APIResponse
// @Failure 400 {object} dto.APIResponse
// @Failure 409 {object} dto.APIResponse
// @Router /users [post]
func (h *UserHandler) CreateUser(c echo.Context) error {
	var req dto.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		config.Warn("Invalid request body for user creation")
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	user, err := h.userService.CreateUser(&req)
	if err != nil {
		if err.Error() == "user already exists" {
			return c.JSON(http.StatusConflict, dto.NewErrorResponse("User already exists"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Internal server error"))
	}

	userResponse := dto.NewUserResponse(
		user.ID,
		user.Email,
		user.FullName,
		user.Role,
		user.CreatedAt.Format(time.RFC3339),
		user.UpdatedAt.Format(time.RFC3339),
	)

	config.Info("User created successfully: " + user.Email)
	return c.JSON(http.StatusCreated, dto.NewSuccessResponse("User created successfully", userResponse))
}

// GetUserByID handles getting a user by ID
// @Summary Get user by ID
// @Description Get user profile by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetUserByID(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("User ID is required"))
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		if err.Error() == "user not found" {
			return c.JSON(http.StatusNotFound, dto.NewErrorResponse("User not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Internal server error"))
	}

	userResponse := dto.NewUserResponse(
		user.ID,
		user.Email,
		user.FullName,
		user.Role,
		user.CreatedAt.Format(time.RFC3339),
		user.UpdatedAt.Format(time.RFC3339),
	)

	return c.JSON(http.StatusOK, dto.NewSuccessResponse("User retrieved successfully", userResponse))
}

// GetUserByEmail handles getting a user by email (used by internal services)
func (h *UserHandler) GetUserByEmail(c echo.Context) error {
	email := c.Param("email")
	if email == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Email is required"))
	}

	user, err := h.userService.GetUserByEmail(email)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, dto.NewErrorResponse("User not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Internal server error"))
	}

	userResponse := dto.NewUserResponse(
		user.ID,
		user.Email,
		user.FullName,
		user.Role,
		user.CreatedAt.Format(time.RFC3339),
		user.UpdatedAt.Format(time.RFC3339),
	)

	return c.JSON(http.StatusOK, dto.NewSuccessResponse("User retrieved successfully", userResponse))
}

// AuthenticateUser validates email + password and returns user profile (used by auth service)
func (h *UserHandler) AuthenticateUser(c echo.Context) error {
	var req dto.AuthenticateUserRequest
	if err := c.Bind(&req); err != nil {
		config.Warn("Invalid request body for authentication")
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	user, err := h.userService.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		config.Warn("Authentication failed for: " + req.Email)
		return c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Invalid credentials"))
	}

	userResponse := dto.NewUserResponse(
		user.ID,
		user.Email,
		user.FullName,
		user.Role,
		user.CreatedAt.Format(time.RFC3339),
		user.UpdatedAt.Format(time.RFC3339),
	)

	config.Info("User authenticated: " + user.Email)
	return c.JSON(http.StatusOK, dto.NewSuccessResponse("Authentication successful", userResponse))
}

// UpdateUser handles updating a user
// @Summary Update user
// @Description Update user profile
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body dto.UpdateUserRequest true "Update details"
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("User ID is required"))
	}

	var req dto.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		config.Warn("Invalid request body for user update")
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	user, err := h.userService.UpdateUser(userID, &req)
	if err != nil {
		if err.Error() == "user not found" {
			return c.JSON(http.StatusNotFound, dto.NewErrorResponse("User not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Internal server error"))
	}

	userResponse := dto.NewUserResponse(
		user.ID,
		user.Email,
		user.FullName,
		user.Role,
		user.CreatedAt.Format(time.RFC3339),
		user.UpdatedAt.Format(time.RFC3339),
	)

	config.Info("User updated successfully: " + userID)
	return c.JSON(http.StatusOK, dto.NewSuccessResponse("User updated successfully", userResponse))
}

// DeleteUser handles deleting a user
// @Summary Delete user
// @Description Delete user profile
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("User ID is required"))
	}

	if err := h.userService.DeleteUser(userID); err != nil {
		if err.Error() == "user not found" {
			return c.JSON(http.StatusNotFound, dto.NewErrorResponse("User not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Internal server error"))
	}

	config.Info("User deleted successfully: " + userID)
	return c.JSON(http.StatusOK, dto.NewSuccessResponse("User deleted successfully", nil))
}

// ListUsers handles listing users with pagination
// @Summary List users
// @Description Get paginated list of users
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search term"
// @Success 200 {object} dto.APIResponse
// @Router /users [get]
func (h *UserHandler) ListUsers(c echo.Context) error {
	// Parse query parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	search := c.QueryParam("search")

	req := &dto.ListUsersRequest{
		Page:   page,
		Limit:  limit,
		Search: search,
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error()))
	}

	response, err := h.userService.ListUsers(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Internal server error"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse("Users retrieved successfully", response))
}
