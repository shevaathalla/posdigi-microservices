package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// User represents the user model in the database
type User struct {
	ID        string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Email     string `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password  string `gorm:"type:varchar(255);not null" json:"-"`
	FullName  string `gorm:"type:varchar(255)" json:"full_name"`
	Role      string `gorm:"type:varchar(50);default:'user'" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserRepository interface defines user database operations
type UserRepository interface {
	Create(user *User) error
	FindByEmail(email string) (*User, error)
	FindByID(id string) (*User, error)
	Update(user *User) error
	Delete(id string) error
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user in the database
func (r *userRepository) Create(user *User) error {
	return r.db.Create(user).Error
}

// FindByEmail finds a user by email
func (r *userRepository) FindByEmail(email string) (*User, error) {
	var user User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByID finds a user by ID
func (r *userRepository) FindByID(id string) (*User, error) {
	var user User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Update updates a user in the database
func (r *userRepository) Update(user *User) error {
	return r.db.Save(user).Error
}

// Delete soft deletes a user from the database
func (r *userRepository) Delete(id string) error {
	return r.db.Delete(&User{}, "id = ?", id).Error
}