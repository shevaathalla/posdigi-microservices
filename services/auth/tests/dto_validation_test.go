package tests

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"posdigi-auth/dto"
)

// TestDTO_RegisterRequest_Validation tests RegisterRequest validation
func TestDTO_RegisterRequest_Validation(t *testing.T) {
	validate := validator.New()

	t.Run("Valid registration request", func(t *testing.T) {
		req := dto.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		err := validate.Struct(req)
		assert.NoError(t, err)
	})

	t.Run("Missing email", func(t *testing.T) {
		req := dto.RegisterRequest{
			Password: "password123",
		}

		err := validate.Struct(req)
		assert.Error(t, err)
	})

	t.Run("Invalid email format", func(t *testing.T) {
		req := dto.RegisterRequest{
			Email:    "not-an-email",
			Password: "password123",
		}

		err := validate.Struct(req)
		assert.Error(t, err)
	})

	t.Run("Password too short (5 chars)", func(t *testing.T) {
		req := dto.RegisterRequest{
			Email:    "test@example.com",
			Password: "12345",
		}

		err := validate.Struct(req)
		assert.Error(t, err)
	})

	t.Run("Password way too short (3 chars)", func(t *testing.T) {
		req := dto.RegisterRequest{
			Email:    "test@example.com",
			Password: "123",
		}

		err := validate.Struct(req)
		assert.Error(t, err)
	})

	t.Run("With valid employee data", func(t *testing.T) {
		empData := &dto.EmployeeDataRequest{
			FullName:  "John Doe",
			Phone:     "1234567890",
			Department: "Engineering",
			HireDate:  "2024-01-15",
		}
		req := dto.RegisterRequest{
			Email:        "john@example.com",
			Password:     "password123",
			EmployeeData: empData,
		}

		err := validate.Struct(req)
		assert.NoError(t, err)
	})
}

// TestDTO_LoginRequest_Validation tests LoginRequest validation
func TestDTO_LoginRequest_Validation(t *testing.T) {
	validate := validator.New()

	t.Run("Valid login request", func(t *testing.T) {
		req := dto.LoginRequest{
			Email:    "test@example.com",
			Password: "any_password",
		}

		err := validate.Struct(req)
		assert.NoError(t, err)
	})

	t.Run("Missing email", func(t *testing.T) {
		req := dto.LoginRequest{
			Password: "password123",
		}

		err := validate.Struct(req)
		assert.Error(t, err)
	})

	t.Run("Invalid email", func(t *testing.T) {
		req := dto.LoginRequest{
			Email:    "invalid-email",
			Password: "password123",
		}

		err := validate.Struct(req)
		assert.Error(t, err)
	})

	t.Run("Missing password", func(t *testing.T) {
		req := dto.LoginRequest{
			Email: "test@example.com",
		}

		err := validate.Struct(req)
		assert.Error(t, err)
	})
}

// TestDTO_TokenRequest_Validation tests TokenRequest validation
func TestDTO_TokenRequest_Validation(t *testing.T) {
	validate := validator.New()

	t.Run("Valid token", func(t *testing.T) {
		req := dto.TokenRequest{
			Token: "valid.jwt.token",
		}

		err := validate.Struct(req)
		assert.NoError(t, err)
	})

	t.Run("Empty token", func(t *testing.T) {
		req := dto.TokenRequest{}

		err := validate.Struct(req)
		assert.Error(t, err)
	})
}

// TestDTO_EmployeeDataRequest_Validation tests EmployeeDataRequest validation
func TestDTO_EmployeeDataRequest_Validation(t *testing.T) {
	validate := validator.New()

	t.Run("Valid employee data", func(t *testing.T) {
		req := dto.EmployeeDataRequest{
			FullName:  "John Doe",
			Phone:     "1234567890",
			Department: "Engineering",
			Position:  "Software Developer",
			Salary:    75000,
			HireDate:  "2024-01-15",
		}

		err := validate.Struct(req)
		assert.NoError(t, err)
	})

	t.Run("Name too short", func(t *testing.T) {
		req := dto.EmployeeDataRequest{
			FullName: "J",
		}

		err := validate.Struct(req)
		assert.Error(t, err)
	})

	t.Run("Invalid employment status", func(t *testing.T) {
		req := dto.EmployeeDataRequest{
			FullName:         "John Doe",
			EmploymentStatus: "invalid_status",
		}

		err := validate.Struct(req)
		assert.Error(t, err)
	})

	t.Run("Negative salary", func(t *testing.T) {
		req := dto.EmployeeDataRequest{
			FullName: "John Doe",
			Salary:   -1000,
		}

		err := validate.Struct(req)
		assert.Error(t, err)
	})

	t.Run("Valid employment statuses", func(t *testing.T) {
		validStatuses := []string{"active", "terminated", "on_leave", "suspended"}

		for _, status := range validStatuses {
			req := dto.EmployeeDataRequest{
				FullName:         "John Doe",
				EmploymentStatus: status,
			}

			err := validate.Struct(req)
			assert.NoError(t, err, "Status should be valid: "+status)
		}
	})
}

// TestEmailValidation_EdgeCases tests edge cases for email validation
func TestEmailValidation_EdgeCases(t *testing.T) {
	validate := validator.New()

	t.Run("Valid email formats", func(t *testing.T) {
		validEmails := []string{
			"test@example.com",
			"user.name@example.com",
			"user+tag@example.com",
			"user123@test-site.com",
		}

		for _, email := range validEmails {
			t.Run("Valid: "+email, func(t *testing.T) {
				type EmailTest struct {
					Email string `validate:"required,email"`
				}
				req := EmailTest{Email: email}
				err := validate.Struct(req)
				assert.NoError(t, err, "Email should be valid: "+email)
			})
		}
	})

	t.Run("Invalid email formats", func(t *testing.T) {
		invalidEmails := []string{
			"plainaddress",
			"@missinglocal.com",
			"username@",
			"username@.com",
		}

		for _, email := range invalidEmails {
			t.Run("Invalid: "+email, func(t *testing.T) {
				type EmailTest struct {
					Email string `validate:"required,email"`
				}
				req := EmailTest{Email: email}
				err := validate.Struct(req)
				assert.Error(t, err, "Email should be invalid: "+email)
			})
		}
	})
}

// TestPasswordValidation tests password strength validation
func TestPasswordValidation(t *testing.T) {
	validate := validator.New()

	t.Run("Valid passwords", func(t *testing.T) {
		validPasswords := []string{
			"password123",     // exactly 6 chars
			"SecurePass123!",  // longer password
			"123456",           // exactly 6 chars
		}

		for _, password := range validPasswords {
			t.Run("Valid: "+password, func(t *testing.T) {
				type PasswordTest struct {
					Password string `validate:"required,min=6"`
				}
				req := PasswordTest{Password: password}
				err := validate.Struct(req)
				assert.NoError(t, err, "Password should be valid: "+password)
			})
		}
	})

	t.Run("Invalid passwords", func(t *testing.T) {
		invalidPasswords := []string{
			"",       // empty
			"12345",   // too short
			"1",       // way too short
		}

		for _, password := range invalidPasswords {
			t.Run("Invalid: length "+fmt.Sprintf("%d", len(password)), func(t *testing.T) {
				type PasswordTest struct {
					Password string `validate:"required,min=6"`
				}
				req := PasswordTest{Password: password}
				err := validate.Struct(req)
				assert.Error(t, err, "Password should be invalid: length "+fmt.Sprintf("%d", len(password)))
			})
		}
	})
}