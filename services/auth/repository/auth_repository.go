package repository

import (
	"gorm.io/gorm"
)

// AuthUser represents a lightweight user model for authentication
type AuthUser struct {
	ID       string `gorm:"type:uuid;primaryKey;"`
	Email    string `gorm:"type:varchar(255);uniqueIndex;not null"`
	Password string `gorm:"type:varchar(255);not null"`
	Role     string `gorm:"type:varchar(50);default:'user'"`
}

// AuthRepository handles authentication-specific database operations
type AuthRepository interface {
	FindByEmail(email string) (*AuthUser, error)
	CreateUser(email, hashedPassword string) error
}

type authRepository struct {
	db *gorm.DB
}

// NewAuthRepository creates a new auth repository instance
func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}

// FindByEmail finds a user by email for authentication
func (r *authRepository) FindByEmail(email string) (*AuthUser, error) {
	var user AuthUser
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// CreateUser creates a new user with email and password
func (r *authRepository) CreateUser(email, hashedPassword string) error {
	user := AuthUser{
		Email:    email,
		Password: hashedPassword,
		Role:     "user",
	}
	return r.db.Create(&user).Error
}