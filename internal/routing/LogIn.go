package routing

import (
	"UrlShortener/internal/models"
	"UrlShortener/pkg/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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
	//token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
	//	"sub": user.ID,
	//	"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	//})
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user.ID
	accessTokenExpireTime := time.Now().Add(time.Hour * 24).Unix()
	claims["exp"] = accessTokenExpireTime

	secret := os.Getenv("JWT_SECRET")
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Printf("Failed to sign token: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to sign token")
	}
	return c.JSON(http.StatusOK, echo.Map{
		"token": tokenString, // Возвращаем токен в ответе
	})
}

var (
	conf1 = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/auth/google/callback/login",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
)

func HandleHomeLogin(c echo.Context) error {
	url := conf1.AuthCodeURL("state", oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func HandleCallbackLogin(c echo.Context) error {
	ctx := context.Background()
	code := c.QueryParam("code")
	if code == "" {
		return c.String(http.StatusBadRequest, "No code in the URL")
	}

	tok, err := conf1.Exchange(ctx, code)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Unable to exchange code for token: "+err.Error())
	}

	client := conf1.Client(ctx, tok)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Unable to get user info: "+err.Error())
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email  string `json:"email"`
		UserID string `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return c.String(http.StatusInternalServerError, "Unable to decode user info response: "+err.Error())
	}

	var userProfile models.UserProfile
	result := db.Where("google_id = ?", userInfo.UserID).First(&userProfile)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {

			return c.String(http.StatusUnauthorized, "User not found")
		}
		return c.String(http.StatusInternalServerError, "Database error: "+result.Error.Error())
	}
	var user models.User
	result = db.Where("id = ?", userProfile.UserID).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {

			return c.String(http.StatusUnauthorized, "User not found")
		}
		return c.String(http.StatusInternalServerError, "Database error: "+result.Error.Error())
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user.ID
	accessTokenExpireTime := time.Now().Add(time.Hour * 24).Unix()
	claims["exp"] = accessTokenExpireTime

	secret := os.Getenv("JWT_SECRET")
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Printf("Failed to sign token: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to sign token")
	}

	cookie := new(http.Cookie)
	cookie.Name = "Authorization"
	cookie.MaxAge = 24 * 60 * 60 * 30
	cookie.Value = tokenString
	cookie.Path = "/"
	cookie.HttpOnly = false
	c.SetCookie(cookie)

	return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/users/profile/%v", user.ID))
}
