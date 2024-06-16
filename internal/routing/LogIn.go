package routing

import (
	"UrlShortener/internal/models"
	"UrlShortener/pkg/utils"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"time"
)

var db *gorm.DB

func LoginTPL(c echo.Context) error {
	templateFile := "assets/templates/login.html"
	return utils.ExecuteTemplate(c, templateFile)
}
func Validate(c echo.Context) error {
	user := c.Get("user")
	return c.JSON(http.StatusOK, echo.Map{"message": user})
}
func LogIn(c echo.Context) error {
	input := new(models.UserData)
	if err := c.Bind(input); err != nil {
		log.Printf("Failed to bind user input: %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	if err := c.Validate(input); err != nil {
		log.Printf("Failed to validate user input: %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	var user models.User
	result := db.First(&user, "email = ?", input.Email)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, "Incorrect email or password")
		}
		log.Printf("Error finding user: %s", result.Error)
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(input.Password))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, "Incorrect email or password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	secret := os.Getenv("JWT_SECRET")
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Printf("Failed to sign token: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to sign token")
	}
	c.SetCookie(&http.Cookie{
		Name:     "Authorization",
		Value:    tokenString,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24 * 30),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true})
	return c.JSON(http.StatusOK, echo.Map{"message": "Login successful"})
}
