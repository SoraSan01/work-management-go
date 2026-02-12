package controllers

import (
	"net/http"
	"strings"
	"time"
	"work-management-system/repositories"
	"work-management-system/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthenticationController struct {
	Repo *repositories.AuthenticationRepository
}

func NewAuthenticationController(repo *repositories.AuthenticationRepository) *AuthenticationController {
	return &AuthenticationController{Repo: repo}
}

// --------------------
// Login POST
// --------------------
func (ac *AuthenticationController) Login(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	// Call repo login (returns user + tokens)
	user, access, refresh, err := ac.Repo.Login(email, password)
	if err != nil {
		c.HTML(http.StatusUnauthorized, "authentication/login.html", gin.H{
			"error": err.Error(),
		})
		return
	}

	// Store tokens in secure HttpOnly cookies
	c.SetCookie("access_token", access, 900, "/", "", false, true)         // 15 min
	c.SetCookie("refresh_token", refresh, 7*24*3600, "/", "", false, true) // 7 days

	// Redirect based on user role
	role := strings.ToLower(user.Role.Name)
	switch role {
	case "admin":
		c.Redirect(http.StatusSeeOther, "/admin/dashboard")
	case "manager":
		c.Redirect(http.StatusSeeOther, "/manager/dashboard")
	case "employee":
		c.Redirect(http.StatusSeeOther, "/employee/dashboard")
	case "customer":
		c.Redirect(http.StatusSeeOther, "/customer/dashboard")
	case "supervisor":
		c.Redirect(http.StatusSeeOther, "/supervisor/dashboard")
	case "qualityassurance":
		c.Redirect(http.StatusSeeOther, "/qa/dashboard")
	default:
		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}

func (ac *AuthenticationController) SignupPage(c *gin.Context) {
	c.HTML(http.StatusOK, "authentication/register.html", gin.H{
		"title": "Sign Up",
	})
}

func (ac *AuthenticationController) Signup(c *gin.Context) {
	firstName := strings.TrimSpace(c.PostForm("first_name"))
	lastName := strings.TrimSpace(c.PostForm("last_name"))
	email := strings.TrimSpace(c.PostForm("email"))
	companyName := strings.TrimSpace(c.PostForm("company_name"))
	contactNumber := strings.TrimSpace(c.PostForm("contact_number"))
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")

	formData := gin.H{
		"first_name":     firstName,
		"last_name":      lastName,
		"email":          email,
		"company_name":   companyName,
		"contact_number": contactNumber,
	}

	if password != confirmPassword {
		c.HTML(http.StatusBadRequest, "authentication/register.html", gin.H{
			"error": "Passwords do not match.",
			"form":  formData,
		})
		return
	}

	if len(password) < 8 {
		c.HTML(http.StatusBadRequest, "authentication/register.html", gin.H{
			"error": "Password must be at least 8 characters.",
			"form":  formData,
		})
		return
	}

	if err := ac.Repo.RegisterCustomer(firstName, lastName, email, password, companyName, contactNumber); err != nil {
		c.HTML(http.StatusBadRequest, "authentication/register.html", gin.H{
			"error": err.Error(),
			"form":  formData,
		})
		return
	}

	c.HTML(http.StatusOK, "authentication/login.html", gin.H{
		"message": "Registration successful. You can now sign in.",
	})
}

// --------------------
// Logout
// --------------------
func (ac *AuthenticationController) Logout(c *gin.Context) {
	userID := c.GetString("user_id")
	user, err := ac.Repo.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}

	// Clear refresh token in DB
	user.RefreshToken = ""
	ac.Repo.Update(user)

	// Clear cookies
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)

	// Redirect to login page
	c.Redirect(http.StatusSeeOther, "/login")
}

// --------------------
// Refresh token endpoint
// --------------------
func (ac *AuthenticationController) RefreshToken(c *gin.Context) {
	// Get refresh token from cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// Validate and generate new tokens
	newAccess, newRefresh, err := ac.Repo.RefreshTokens(refreshToken)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// Set new cookies
	c.SetCookie("access_token", newAccess, 900, "/", "", false, true)
	c.SetCookie("refresh_token", newRefresh, 7*24*3600, "/", "", false, true)

	// Redirect to dashboard (could also decode role from JWT if needed)
	c.Redirect(http.StatusSeeOther, "/dashboard")
}

// --------------------
// Forgot Password
// --------------------
func (ac *AuthenticationController) ForgotPasswordPage(c *gin.Context) {
	c.HTML(http.StatusOK, "authentication/forgot_password.html", gin.H{
		"title": "Forgot Password",
	})
}

func (ac *AuthenticationController) ForgotPassword(c *gin.Context) {
	email := c.PostForm("email")

	user, err := ac.Repo.FindByEmail(email)
	if err != nil {
		// Don’t reveal if user exists or not
		c.HTML(http.StatusOK, "authentication/forgot_password.html", gin.H{
			"message": "If the email exists, a reset link has been sent.",
		})
		return
	}

	// Generate a reset token (JWT or random string)
	resetToken, resetExp := utils.GenerateResetToken(user.ID.String())

	user.ResetToken = resetToken
	user.ResetExp = resetExp
	ac.Repo.Update(user)

	// Send email with link
	resetLink := "http://localhost:8080/reset-password?token=" + resetToken
	utils.SendEmail(user.Email, "Password Reset", "Click here to reset your password: "+resetLink)

	c.HTML(http.StatusOK, "authentication/forgot_password.html", gin.H{
		"message": "If the email exists, a reset link has been sent.",
	})
}

// --------------------
// Reset Password
// --------------------
func (ac *AuthenticationController) ResetPasswordPage(c *gin.Context) {
	token := c.Query("token")
	c.HTML(http.StatusOK, "authentication/reset_password.html", gin.H{
		"token": token,
		"title": "Reset Password",
	})
}

func (ac *AuthenticationController) ResetPassword(c *gin.Context) {
	token := c.PostForm("token")
	newPassword := c.PostForm("password")

	user, err := ac.Repo.FindByResetToken(token)
	if err != nil || time.Now().After(user.ResetExp) {
		c.HTML(http.StatusBadRequest, "authentication/reset_password.html", gin.H{
			"error": "Invalid or expired token",
		})
		return
	}

	// Update password
	hash, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	user.PasswordHash = string(hash)
	user.ResetToken = ""
	user.ResetExp = time.Time{} // clear token
	ac.Repo.Update(user)

	c.HTML(http.StatusOK, "authentication/login.html", gin.H{
		"message": "Password reset successfully. Please login.",
	})
}
