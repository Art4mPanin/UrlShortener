package routing

import (
	"UrlShortener/internal/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

func ProfileTPL(c echo.Context) error {
	profileData := c.Get("profile").(*models.UserProfile)
	templateFile := "profile.html"
	return c.Render(http.StatusOK, templateFile, map[string]interface{}{
		"profile": profileData,
	})
}
func AvatarUpdate(c echo.Context) error {
	//ANOTHER STRUCT
	user := c.Get("user").(*models.User)
	urlUserID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	if int(user.ID) != urlUserID {
		log.Printf("User ID mismatch: tokenUserID=%d, urlUserID=%d", user.ID, urlUserID)
		return echo.NewHTTPError(http.StatusForbidden, "User ID mismatch")
	}
	if !user.IsActive {
		return echo.NewHTTPError(http.StatusForbidden, "User is not activated")
	}
	file, err := c.FormFile("avatar")
	if err != nil {
		log.Printf("Invalid file upload: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid file upload")
	}

	src, err := file.Open()
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error opening file")
	}
	defer src.Close()

	dstPath := "uploads/" + file.Filename

	dst, err := os.Create(dstPath)
	if err != nil {
		log.Printf("Error saving file: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error saving file")
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		log.Printf("Error saving file: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error saving file")
	}

	avatarURL := "/uploads/" + file.Filename

	var userProfile models.UserProfile
	result := db.Where("user_id = ?", user.ID).First(&userProfile)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			userProfile = models.UserProfile{
				UserID:    user.ID,
				AvatarURL: avatarURL,
			}
			result = db.Create(&userProfile)
			if result.Error != nil {
				log.Printf("Error creating user profile in database: %s", result.Error)
				return echo.NewHTTPError(http.StatusInternalServerError, "Error creating user profile in database")
			}
		} else {
			log.Printf("Error retrieving user profile from database: %s", result.Error)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error retrieving user profile from database")
		}
	} else {
		userProfile.AvatarURL = avatarURL
		result = db.Save(&userProfile)
		if result.Error != nil {
			log.Printf("Error updating avatar URL in database: %s", result.Error)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error updating database")
		}
	}
	return c.JSON(http.StatusOK, map[string]string{
		"message":    "File uploaded successfully",
		"avatar_url": avatarURL,
	})
}

func ProfileHandler(c echo.Context) error {
	userProfile := c.Get("profile").(*models.UserProfile)
	user := c.Get("user").(*models.User)
	fmt.Printf("USER FROM PROFILE HANLDER %v", user)
	urlUserID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	if int(user.ID) != urlUserID {
		log.Printf("User ID mismatch: tokenUserID=%d, urlUserID=%d", user.ID, urlUserID)
		return echo.NewHTTPError(http.StatusForbidden, "User ID mismatch")
	}
	if !user.IsActive {
		return echo.NewHTTPError(http.StatusForbidden, "User is not activated")
	}
	return c.JSON(http.StatusOK, userProfile)
}

func DataUpdate(c echo.Context) error {
	user := c.Get("user").(*models.User)
	urlUserID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}
	if int(user.ID) != urlUserID {
		log.Printf("User ID mismatch: tokenUserID=%d, urlUserID=%d", user.ID, urlUserID)
		return echo.NewHTTPError(http.StatusForbidden, "User ID mismatch")
	}
	if !user.IsActive {
		return echo.NewHTTPError(http.StatusForbidden, "User is not activated")
	}
	///
	input := new(models.UserProfile)
	if err := c.Bind(input); err != nil {
		log.Printf("Failed to bind user input: %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	///
	var userProfile models.UserProfile
	result := db.Where("user_id = ?", user.ID).First(&userProfile)
	if result.Error != nil {
		log.Printf("User profile not found: %s", result.Error)
		return echo.NewHTTPError(http.StatusNotFound, "User profile not found")
	}
	userProfile.DisplayedName = input.DisplayedName
	userProfile.ProfileTitle = input.ProfileTitle
	userProfile.Bio = input.Bio
	userProfile.Email = input.Email

	if err := db.Save(&userProfile).Error; err != nil {
		log.Printf("Failed to update user profile: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user profile")
	}
	log.Printf("User profile with ID %d successfully updated", user.ID)
	return c.JSON(http.StatusCreated, user)
}

func PassUpdate(c echo.Context) error {
	user := c.Get("user").(*models.User)
	urlUserID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	if int(user.ID) != urlUserID {
		log.Printf("User ID mismatch: tokenUserID=%d, urlUserID=%d", user.ID, urlUserID)
		return echo.NewHTTPError(http.StatusForbidden, "User ID mismatch")
	}
	if !user.IsActive {
		return echo.NewHTTPError(http.StatusForbidden, "User is not activated")
	}
	// Bind input data
	input := new(models.PassConfirmation)
	if err := c.Bind(input); err != nil {
		log.Printf("Failed to bind user input: %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	//id check
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(input.OldPassword))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, "Incorrect email or password")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to hash password")
	}
	result := db.Where("id = ?", user.ID).First(&user)
	if result.Error != nil {
		log.Printf("User profile not found: %s", result.Error)
		return echo.NewHTTPError(http.StatusNotFound, "User profile not found")
	}

	// Update fields
	user.HashedPassword = string(hashedPassword)
	// добавлено для обновления email

	if err := db.Save(&user).Error; err != nil {
		log.Printf("Failed to update user profile: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user profile")
	}

	log.Printf("User profile with ID %d successfully updated", user.ID)
	return c.JSON(http.StatusOK, user)
	//return nil
}

var (
	conf = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/auth/google/callback/login",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
)

func HandleHome(c echo.Context) error {
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func HandleCallback(c echo.Context) error {
	ctx := context.Background()
	code := c.QueryParam("code")
	if code == "" {
		return c.String(http.StatusBadRequest, "No code in the URL")
	}

	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Unable to exchange code for token: "+err.Error())
	}

	client := conf.Client(ctx, tok)
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

	user := c.Get("user").(*models.User)
	var userProfile models.UserProfile
	result := db.Where("user_id = ?", user.ID).First(&userProfile)
	if result.Error != nil {
		log.Printf("User profile not found: %s", result.Error)
		return echo.NewHTTPError(http.StatusNotFound, "User profile not found")
	}
	userProfile.GoogleID = userInfo.UserID

	if err := db.Save(&userProfile).Error; err != nil {
		log.Printf("Failed to update user profile: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user profile")
	}

	return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/users/profile/%v", user.ID))
}
func GoogleUnlink(c echo.Context) error {
	user := c.Get("user").(*models.User)
	urlUserID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}
	if int(user.ID) != urlUserID {
		log.Printf("User ID mismatch: tokenUserID=%d, urlUserID=%d", user.ID, urlUserID)
		return echo.NewHTTPError(http.StatusForbidden, "User ID mismatch")
	}
	if !user.IsActive {
		return echo.NewHTTPError(http.StatusForbidden, "User is not activated")
	}
	///

	var userProfile models.UserProfile
	result := db.Where("user_id = ?", user.ID).First(&userProfile)
	if result.Error != nil {
		log.Printf("User profile not found: %s", result.Error)
		return echo.NewHTTPError(http.StatusNotFound, "User profile not found")
	}
	userProfile.GoogleID = ""
	if err := db.Save(&userProfile).Error; err != nil {
		log.Printf("Failed to update user profile: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user profile")
	}
	log.Printf("User profile with ID %d successfully updated", user.ID)
	return c.JSON(http.StatusCreated, user)
}
