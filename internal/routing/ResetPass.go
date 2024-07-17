package routing

import (
	"UrlShortener/internal/models"
	"UrlShortener/pkg/utils"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

func PasswordResetTPL(c echo.Context) error {
	return c.Render(http.StatusOK, "password_reset.html", nil)
}
func PasswordResetSetNewPassTPL(c echo.Context) error {
	return c.Render(http.StatusOK, "password_reset_setnewpass.html", nil)
}
func SendCodePass(c echo.Context) error {
	input := new(models.ResetPassVerification)
	if err := c.Bind(input); err != nil {
		log.Printf("Failed to bind user input: %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	code, tstmp := utils.VerificationCode(input.Email)
	Verif := models.Verification{
		UserEmail: input.Email,
		Code:      code,
		Timestamp: tstmp,
	}
	//no vibes
	result := db.Create(&Verif)
	if result.Error != nil {
		log.Printf("Failed to insert verification: %s", result.Error)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to insert verification")
	}
	return echo.NewHTTPError(http.StatusOK, "Verification successfully")
}
func ResetPassword(c echo.Context) error {
	input := new(models.ResetPassInput)
	if err := c.Bind(input); err != nil {
		log.Printf("Failed to bind user input: %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	var verification models.Verification
	result := db.Where("user_email =?", input.Email).First(&verification)
	if result.Error != nil {
		log.Printf("User profile not found: %s", result.Error)
		return echo.NewHTTPError(http.StatusNotFound, "User profile not found")
	}
	if verification.Code != input.Code || verification.Timestamp.Before(time.Now().Add(-time.Minute*30)) {
		log.Printf("Invalid verification code or expired")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid verification code or expired")
	}
	//no vibes
	var user models.User
	if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		log.Printf("User not found: %s", err.Error())
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}
	pass, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to hash password")
	}
	user.HashedPassword = string(pass)
	db.Save(&user)
	return echo.NewHTTPError(http.StatusOK, "Password successfully reset")
}
