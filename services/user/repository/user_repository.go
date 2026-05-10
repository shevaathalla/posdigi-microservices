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
	List(page, limit int, search string) ([]*User, int, error)
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

// FindByEmail finds a user by email, including soft-deleted records.
// Using Unscoped so we catch emails still held by the unique constraint
// even after soft-deletion, preventing spurious duplicate-key DB errors.
func (r *userRepository) FindByEmail(email string) (*User, error) {
	var user User
	err := r.db.Unscoped().Where("email = ?", email).First(&user).Error
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

// List retrieves a paginated list of users with optional search
func (r *userRepository) List(page, limit int, search string) ([]*User, int, error) {
	var users []*User
	var total int64

	query := r.db.Model(&User{})
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("email ILIKE ? OR full_name ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, int(total), nil
}