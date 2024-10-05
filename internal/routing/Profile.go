package routing

import (
	"UrlShortener/internal/models"
	"UrlShortener/pkg/singleton"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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

	user.HashedPassword = string(hashedPassword)

	if err = db.Save(&user).Error; err != nil {
		log.Printf("Failed to update user profile: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user profile")
	}

	log.Printf("User profile with ID %d successfully updated", user.ID)
	return c.JSON(http.StatusOK, user)
	//return nil
}

func HandleHome(c echo.Context) error {
	conf, ok := singleton.GetAndConvertSingleton[*oauth2.Config]("google_data_profile")
	if !ok {
		return c.String(http.StatusInternalServerError, "singleton error")
	}
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func HandleCallback(c echo.Context) error {

	conf, ok := singleton.GetAndConvertSingleton[*oauth2.Config]("google_data_profile")
	if !ok {
		return c.String(http.StatusInternalServerError, "singleton error")
	}

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

func HandleHomeVK(c echo.Context) error {
	conf, ok := singleton.GetAndConvertSingleton[*oauth2.Config]("vk_data_profile")
	if !ok {
		return c.String(http.StatusInternalServerError, "singleton error")
	}
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func HandleCallbackVK(c echo.Context) error {
	conf, ok := singleton.GetAndConvertSingleton[*oauth2.Config]("vk_data_profile")
	if !ok {
		return c.String(http.StatusInternalServerError, "singleton error")
	}
	ctx := context.Background()
	code := c.QueryParam("code")
	if code == "" {
		return c.String(http.StatusBadRequest, "No code in the URL")
	}

	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Unable to exchange code for token: "+err.Error())
	}
	fmt.Println(tok)
	client := conf.Client(ctx, tok)
	resp, err := client.Get("https://api.vk.com/method/users.get?fields=photo_400_orig&access_token=" + tok.AccessToken + "&v=5.131")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Unable to get user info: "+err.Error())
	}
	defer resp.Body.Close()

	var vkResponse struct {
		Response []struct {
			ID        int64  `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Photo     string `json:"photo_400_orig"`
		} `json:"response"`
	}
	fmt.Println(vkResponse)
	if err := json.NewDecoder(resp.Body).Decode(&vkResponse); err != nil {
		return c.String(http.StatusInternalServerError, "Unable to decode user info response: "+err.Error())
	}

	if len(vkResponse.Response) == 0 {
		return c.String(http.StatusInternalServerError, "No user info found in the response")
	}

	user := c.Get("user").(*models.User)
	var userProfile models.UserProfile
	result := db.Where("user_id = ?", user.ID).First(&userProfile)
	if result.Error != nil {
		log.Printf("User profile not found: %s", result.Error)
		return echo.NewHTTPError(http.StatusNotFound, "User profile not found")
	}

	userProfile.VkId = strconv.FormatInt(vkResponse.Response[0].ID, 10)

	if err := db.Save(&userProfile).Error; err != nil {
		log.Printf("Failed to update user profile: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user profile")
	}

	return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/users/profile/%v", user.ID))
}
func VKUnlink(c echo.Context) error {
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
	userProfile.VkId = ""
	if err := db.Save(&userProfile).Error; err != nil {
		log.Printf("Failed to update user profile: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user profile")
	}
	log.Printf("User profile with ID %d successfully updated", user.ID)
	return c.JSON(http.StatusCreated, user)
}
func TGLINK(c echo.Context) error {
	user := c.Get("user").(*models.User)
	log.Printf("Received request from user ID: %d, Username: %s", user.ID, user.Username)

	if !user.IsActive {
		log.Printf("User ID %d is not activated", user.ID)
		return echo.NewHTTPError(http.StatusForbidden, "User is not activated")
	}

	var userProfileID models.UserProfile
	result := db.Where("user_id = ?", user.ID).First(&userProfileID)
	if result.Error != nil {
		log.Printf("User profile not found for user ID %d: %s", user.ID, result.Error)
		return echo.NewHTTPError(http.StatusNotFound, "User profile not found")
	}
	log.Printf("User profile found for user ID %d", user.ID)

	input := new(models.TgVerification)
	if err := c.Bind(input); err != nil {
		log.Printf("Failed to bind user input for user ID %d: %s", user.ID, err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	log.Printf("User input for TG ID %d received for user ID %d", input.TgID, user.ID)
	check, err := TgVerificiation(input)
	if err != nil || !check {
		log.Printf("Failed to verify TG ID for user ID %d: %s", user.ID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify TG ID")
	}

	fmt.Println(input)
	var userProfile models.UserProfile
	result1 := db.Where("user_id = ?", user.ID).First(&userProfile)
	if result1.Error != nil {
		log.Printf("User profile not found on second fetch for user ID %d: %s", user.ID, result1.Error)
		return echo.NewHTTPError(http.StatusNotFound, "User profile not found")
	}
	log.Printf("User profile found on second fetch for user ID %d", user.ID)

	userProfile.TgId = strconv.Itoa(input.TgID)
	log.Printf("Updating TG ID to %v for user ID %d", input.TgID, user.ID)

	if err = db.Save(&userProfile).Error; err != nil {
		log.Printf("Failed to update user profile for user ID %v: %s", user.ID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user profile")
	}
	log.Printf("User profile with ID %d successfully updated with TG ID %d", user.ID, input.TgID)
	return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/users/profile/%v", user.ID))
}
func TgStringBuilder(v *models.TgVerification) string {
	var dataParts []string

	if v.AuthDate != 0 {
		dataParts = append(dataParts, fmt.Sprintf("auth_date=%v", v.AuthDate))
	}
	if v.FirstName != "" {
		dataParts = append(dataParts, fmt.Sprintf("first_name=%v", v.FirstName))
	}
	if v.TgID != 0 {
		dataParts = append(dataParts, fmt.Sprintf("id=%v", v.TgID))
	}
	if v.LastName != "" {
		dataParts = append(dataParts, fmt.Sprintf("last_name=%v", v.LastName))
	}
	if v.PhotoUrl != "" {
		dataParts = append(dataParts, fmt.Sprintf("photo_url=%v", v.PhotoUrl))
	}
	if v.Username != "" {
		dataParts = append(dataParts, fmt.Sprintf("username=%v", v.Username))
	}

	return strings.Join(dataParts, "\n")
}
func TgVerificiation(v *models.TgVerification) (bool, error) {
	botToken := os.Getenv("BOT_TOKEN")
	token256 := sha256.Sum256([]byte(botToken))
	checkstring := TgStringBuilder(v)
	h := hmac.New(sha256.New, token256[:])
	h.Write([]byte(checkstring))
	calculatedhash := hex.EncodeToString(h.Sum(nil))

	if calculatedhash != v.Hash {
		log.Printf("Invalid TG verification data: calculated hash=%s, received hash=%s", calculatedhash, v.Hash)
		return false, errors.New("invalid hash, data may be tampered with")
	}
	log.Printf("TG verification data is valid: calculated hash=%s, received hash=%s", calculatedhash, v.Hash)
	return true, nil
}
func TgUnlink(c echo.Context) error {
	user := c.Get("user").(*models.User)
	log.Printf("Received request from user ID: %d, Username: %s", user.ID, user.Username)

	if !user.IsActive {
		log.Printf("User ID %d is not activated", user.ID)
		return echo.NewHTTPError(http.StatusForbidden, "User is not activated")
	}

	var userProfileID models.UserProfile
	result := db.Where("user_id = ?", user.ID).First(&userProfileID)
	if result.Error != nil {
		log.Printf("User profile not found for user ID %d: %s", user.ID, result.Error)
		return echo.NewHTTPError(http.StatusNotFound, "User profile not found")
	}
	log.Printf("User profile found for user ID %d", user.ID)
	userProfileID.TgId = ""
	if err := db.Save(&userProfileID).Error; err != nil {
		log.Printf("Failed to update user profile for user ID %d: %s", user.ID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user profile")
	}
	log.Printf("User profile with ID %d successfully updated with TG ID removed", user.ID)
	return c.JSON(http.StatusCreated, user)
}
