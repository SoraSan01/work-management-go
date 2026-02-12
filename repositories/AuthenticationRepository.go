package repositories

import (
	"errors"
	"strings"
	"time"
	"work-management-system/models"
	"work-management-system/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthenticationRepository struct {
	DB *gorm.DB
}

func NewAuthenticationRepository(db *gorm.DB) *AuthenticationRepository {
	return &AuthenticationRepository{DB: db}
}

// Login verifies email + password
func (ar *AuthenticationRepository) Login(email, password string) (*models.User, string, string, error) {
	var user models.User

	err := ar.DB.
		Preload("Role").
		Preload("Role.Permissions").
		Where("email = ?", email).
		First(&user).Error
	if err != nil {
		return nil, "", "", errors.New("invalid credentials")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, "", "", errors.New("invalid credentials")
	}

	permissions := []string{}
	for _, p := range user.Role.Permissions {
		permissions = append(permissions, p.Name)
	}

	// Generate the full name
	fullName := user.FirstName + " " + user.LastName

	accessToken, err := utils.GenerateAccessToken(
		user.ID.String(),
		user.Role.Name,
		fullName, // Combine first and last name
		user.Email,
		permissions,
	)

	refreshToken, refreshExp := utils.GenerateRefreshToken(user.ID.String())

	user.RefreshToken = refreshToken
	user.RefreshExp = refreshExp
	ar.DB.Save(&user)

	return &user, accessToken, refreshToken, nil
}

func (ar *AuthenticationRepository) FindByID(id string) (*models.User, error) {
	var user models.User
	if err := ar.DB.Preload("Role").Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (ar *AuthenticationRepository) Update(user *models.User) error {
	return ar.DB.Save(user).Error
}

func (ar *AuthenticationRepository) RefreshTokens(refreshToken string) (string, string, error) {
	user := &models.User{}
	err := ar.DB.
		Preload("Role").
		Preload("Role.Permissions").
		Where("refresh_token = ?", refreshToken).
		First(user).Error
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	if time.Now().After(user.RefreshExp) {
		return "", "", errors.New("refresh token expired")
	}

	permissions := []string{}
	for _, p := range user.Role.Permissions {
		permissions = append(permissions, p.Name)
	}

	// Generate new tokens
	fullName := user.FirstName + " " + user.LastName

	access, err := utils.GenerateAccessToken(
		user.ID.String(),
		user.Role.Name,
		fullName, // Combine first and last name
		user.Email,
		permissions,
	)
	newRefresh, newExp := utils.GenerateRefreshToken(user.ID.String())

	// Save new refresh token in DB
	user.RefreshToken = newRefresh
	user.RefreshExp = newExp
	ar.DB.Save(user)

	return access, newRefresh, nil
}

func (ar *AuthenticationRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := ar.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (ar *AuthenticationRepository) FindByResetToken(token string) (*models.User, error) {
	var user models.User
	if err := ar.DB.Where("reset_token = ?", token).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (ar *AuthenticationRepository) RegisterCustomer(firstName, lastName, email, password, companyName, contactNumber string) error {
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	email = strings.ToLower(strings.TrimSpace(email))
	companyName = strings.TrimSpace(companyName)
	contactNumber = strings.TrimSpace(contactNumber)

	if firstName == "" || lastName == "" || email == "" || password == "" || companyName == "" {
		return errors.New("all required fields must be provided")
	}

	var existing models.User
	if err := ar.DB.Where("email = ?", email).First(&existing).Error; err == nil {
		return errors.New("email already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("failed to validate email")
	}

	var role models.Role
	if err := ar.DB.Where("LOWER(name) = ?", "customer").First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("customer role is not configured")
		}
		return errors.New("failed to load customer role")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to process password")
	}

	return ar.DB.Transaction(func(tx *gorm.DB) error {
		user := models.User{
			ID:           uuid.New(),
			FirstName:    firstName,
			LastName:     lastName,
			Email:        email,
			PasswordHash: string(hash),
			RoleID:       role.ID,
			IsActive:     true,
		}
		if err := tx.Create(&user).Error; err != nil {
			return errors.New("failed to create account")
		}

		customer := models.Customer{
			ID:            uuid.New(),
			UserID:        user.ID,
			CompanyName:   companyName,
			ContactNumber: contactNumber,
		}
		if err := tx.Create(&customer).Error; err != nil {
			return errors.New("failed to create customer profile")
		}

		return nil
	})
}
