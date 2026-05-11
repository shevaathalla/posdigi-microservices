# Testing Guide

This document explains how to run and write tests for the Posdigi Microservices project.

## 📁 Test Structure

```
services/
├── auth/
│   └── tests/
│       ├── auth_handler_test.go       # Auth endpoint tests
│       ├── dto_validation_test.go     # DTO validation tests
│       └── integration_test.go        # Integration tests
├── user/
│   └── tests/
│       └── integration_test.go        # User service tests
├── attendance/
│   └── tests/
│       └── integration_test.go        # Attendance service tests
├── gateway/
│   └── tests/
│       └── integration_test.go        # Gateway tests
└── shared/
    └── testing/
        └── test_helpers.go             # Shared testing utilities
```

## 🧪 Running Tests

### Run All Tests
```bash
make test-all
```

### Run Individual Service Tests
```bash
make test-auth        # Auth Service tests
make test-user        # User Service tests
make test-attendance  # Attendance Service tests
make test-gateway     # Gateway Service tests
```

### Run Tests with Coverage
```bash
make test-coverage
```
This generates HTML coverage reports in each service directory.

### Run Tests with Race Detection
```bash
make test-race
```

### Run Benchmark Tests
```bash
make test-bench
```

## 📝 Test Categories

### 1. **DTO Validation Tests**
Test data validation rules and constraints:
- Email format validation
- Password requirements
- Required field validation
- Employee data validation
- Date format validation

### 2. **Integration Tests**
Test complete HTTP workflows:
- Endpoint routing
- Request/response handling
- Error handling
- Authentication/authorization
- Query parameter handling

### 3. **Middleware Tests**
Test custom middleware components:
- Request ID generation
- Logging functionality
- CORS handling
- Rate limiting
- Service authentication

### 4. **Error Handling Tests**
Test error scenarios:
- Invalid JSON requests
- Missing required fields
- Invalid query parameters
- Service authentication failures
- Rate limit violations

## ✅ Test Examples

### Example: Testing an Endpoint
```go
func TestRoute_Register_MissingEmail(t *testing.T) {
    e := SetupTestRouter()

    requestBody := map[string]interface{}{
        "password": "password123",
    }
    bodyJSON, _ := json.Marshal(requestBody)
    req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(bodyJSON))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    e.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusBadRequest, rec.Code)
}
```

### Example: Testing Validation
```go
func TestRegisterRequest_Validation(t *testing.T) {
    validate := validator.New()

    request := dto.RegisterRequest{
        Email:    "test@example.com",
        Password: "SecurePass123",
    }

    err := validate.Struct(request)
    assert.NoError(t, err)
}
```

## 🎯 Best Practices

### 1. **Test Naming**
- Use descriptive names: `TestRoute_Register_MissingEmail`
- Group related tests: `TestAuthHandler_Register_*`
- Use table-driven tests for multiple scenarios

### 2. **Test Structure**
- Arrange: Set up test data and mocks
- Act: Execute the function being tested
- Assert: Verify expected outcomes

### 3. **Error Handling**
- Test both success and failure cases
- Verify proper error messages
- Check appropriate HTTP status codes

### 4. **Assertions**
- Use specific assertions (`assert.Equal`, `assert.Contains`)
- Provide clear failure messages
- Use `require` for setup that must succeed

## 🔧 Testing Utilities

The shared testing utilities in `services/shared/testing/test_helpers.go` provide:

- `MakeRequest()` - Creates HTTP requests
- `AssertJSONSuccess()` - Validates successful JSON responses
- `AssertJSONError()` - Validates error JSON responses
- `AssertStatusCode()` - Checks HTTP status codes
- `AssertValidationError()` - Validates validation error responses

## 🚀 Writing New Tests

### 1. **Create Test File**
```bash
# In the service directory
mkdir -p tests
touch tests/my_feature_test.go
```

### 2. **Import Required Packages**
```go
package tests

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)
```

### 3. **Write Test Functions**
```go
func TestMyFeature_Scenario(t *testing.T) {
    // Arrange
    e := SetupTestRouter()

    // Act
    req := httptest.NewRequest(http.MethodGet, "/api/v1/my-endpoint", nil)
    rec := httptest.NewRecorder()
    e.ServeHTTP(rec, req)

    // Assert
    assert.Equal(t, http.StatusOK, rec.Code)
}
```

### 4. **Run Your Tests**
```bash
make test-auth  # or appropriate service
```

## 📊 Test Coverage Goals

### Minimum Coverage Targets:
- **Handlers**: 70%+ coverage
- **Services**: 80%+ coverage
- **DTOs**: 90%+ coverage (validation rules)
- **Middleware**: 75%+ coverage

### What to Test:
- ✅ All public API endpoints
- ✅ Validation rules
- ✅ Error handling paths
- ✅ Middleware functionality
- ✅ Integration between components

### What NOT to Test:
- ❌ Third-party libraries
- ❌ Standard library functions
- ❌ Trivial getters/setters
- ❌ Generated code

## 🐛 Debugging Failed Tests

### Run Tests Verbosely
```bash
cd services/auth
go test -v ./tests/...
```

### Run Specific Test
```bash
go test -v ./tests/... -run TestRoute_Register
```

### Run with Debug Output
```bash
go test -v ./tests/... -run TestRoute_Register > test_output.txt
```

## 🔄 Continuous Integration

Tests are designed to run in CI/CD pipelines:
- Fast execution (< 30 seconds for all tests)
- No external dependencies (database mocking)
- Deterministic results
- Clear pass/fail indications

## 📚 Additional Resources

- [Go Testing Guide](https://golang.org/pkg/testing/)
- [Testify Assertions](https://github.com/stretchr/testify)
- [Table Driven Tests](https://dave.cheney.net/2019/03/02/table-driven-tests-in-go/)
- [HTTP Testing in Go](https://pkg.go.dev/net/http/httptest)

## 🎓 Training Notes

### Common Testing Patterns

1. **Table-Driven Tests**: Test multiple scenarios in a single test function
2. **Setup/Teardown**: Use `TestMain` for global setup if needed
3. **Test Helpers**: Create reusable helper functions in `test_helpers.go`
4. **Mock Interfaces**: Use interfaces for external dependencies
5. **Subtests**: Use `t.Run()` for grouping related test cases

### Error Message Best Practices

❌ Bad: `assert.Equal(t, 200, rec.Code)`
✅ Good: `assert.Equal(t, 200, rec.Code, "Status code should be OK")`

## 🚨 Troubleshooting

### Import Errors
```bash
# Fix missing testify imports
go mod tidy
```

### Path Issues
```bash
# Run tests from service root directory
cd services/auth
go test ./tests/...
```

### Database Connection Errors
Tests are designed to work without database connections. If you see database errors:
1. Check if you need to add mocking
2. Verify environment variables for tests
3. Ensure test configuration is correct

---

**Remember:** Tests should be fast, reliable, and easy to understand. They're your safety net when making changes to the codebase! 🛡️