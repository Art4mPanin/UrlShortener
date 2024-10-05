package routing

import (
	"UrlShortener/internal/models"
	"UrlShortener/pkg/logger"
	logcfg "UrlShortener/pkg/logger/config"
	"UrlShortener/pkg/singleton"
	"UrlShortener/pkg/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var db *gorm.DB

func LoginTPL(c echo.Context) error {
	templateFile := "assets/templates/login.html"
	return utils.ExecuteTemplateFunc(c, templateFile)
}
func Validate(c echo.Context) error {
	user := c.Get("user")
	return c.JSON(http.StatusOK, echo.Map{"message": user})
}

type LoginService struct {
	db *gorm.DB
}

func NewLoginService(db2 *gorm.DB) *LoginService {
	return &LoginService{
		db: db2,
	}
}

func (l *LoginService) LoginValidateRequest(c echo.Context) (*models.UserData, error) {
	input := new(models.UserData)
	if err := c.Bind(input); err != nil {
		log.Printf("Failed to bind user input: %s", err)
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	if err := c.Validate(input); err != nil {
		log.Printf("Failed to validate user input: %s", err)
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	return input, nil
}

func (l *LoginService) LoginFetchUser(email string) (*models.User, error) {
	var user models.User
	result := l.db.First(&user, "email = ?", email)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, UserNotFoundError{}
		}
		log.Printf("Error finding user: %s", result.Error)
		return nil, UnknownDBError{}
	}
	return &user, nil
}

func (l *LoginService) LoginAuthenticateUser(user *models.User, password string) (string, error) {
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		return "", InvalidPasswordError{}
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user.ID
	accessTokenExpireTime := time.Now().Add(time.Hour * 24 * 30).Unix()
	claims["exp"] = accessTokenExpireTime

	secret := os.Getenv("JWT_SECRET")
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Printf("Failed to sign token: %s", err)
		return "", SignTokenError{}
	}

	return tokenString, nil
}

type UserNotFoundError struct{}

func (e UserNotFoundError) Error() string { return "Incorrect email or password" }

type UnknownDBError struct{}

func (e UnknownDBError) Error() string { return "Database error" }

type InvalidPasswordError struct{}

func (e InvalidPasswordError) Error() string { return "Incorrect email or password" }

type SignTokenError struct{}

func (e SignTokenError) Error() string { return "Failed to sign token" }

func LogIn(c echo.Context) error {
	service := NewLoginService(db)

	// input
	input, err := service.LoginValidateRequest(c)
	if err != nil {
		return err
	}

	// get user
	user, err := service.LoginFetchUser(input.Email)
	if err != nil {
		if errors.As(err, &UserNotFoundError{}) {
			return c.JSON(http.StatusNotFound, err.Error())
		}
		if errors.As(err, &UnknownDBError{}) {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusInternalServerError, "unknown error")
	}

	//token
	tokenString, err := service.LoginAuthenticateUser(user, input.Password)
	if err != nil {
		if errors.As(err, &InvalidPasswordError{}) {
			return c.JSON(401, err.Error())
		}
		if errors.As(err, &SignTokenError{}) {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusInternalServerError, "unknown error")
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": tokenString, // Возвращаем токен в ответе
	})
}

func HandleHomeLogin(c echo.Context) error {
	conf, ok := singleton.GetAndConvertSingleton[*oauth2.Config]("google_data_login")
	if !ok {
		return c.String(http.StatusInternalServerError, "singleton error")
	}
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

type NoCodeInURLError struct{}

func (e NoCodeInURLError) Error() string { return "No code in the URL" }
func (s *LoginService) HandleCallbackloginCode(c echo.Context) (string, error) {
	code := c.QueryParam("code")
	if code == "" {
		log.Println("No code in the URL")
		return "", NoCodeInURLError{}
	}
	return code, nil
}
func (s *LoginService) getOAuthConfig() (*oauth2.Config, error) {
	conf, ok := singleton.GetAndConvertSingleton[*oauth2.Config]("google_data_login")
	if !ok {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "singleton error")
	}
	return conf, nil
}

type CodeExchangeError struct{}

func (e CodeExchangeError) Error() string { return "Unable to exchange code for token" }
func (s *LoginService) exchangeCodeForToken(ctx context.Context, conf *oauth2.Config, code string) (*oauth2.Token, error) {
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Printf("Unable to exchange code for token: %s\n", err)
		return nil, CodeExchangeError{}
	}
	return tok, nil
}

type UserInfo struct {
	Email  string `json:"email"`
	UserID string `json:"id"`
}

func (s *LoginService) fetchUserInfo(ctx context.Context, tok *oauth2.Token, conf *oauth2.Config) (*UserInfo, error) {
	client := conf.Client(ctx, tok)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Printf("Unable to get user info: %s\n", err)
		return nil, fmt.Errorf("unable to get user info: %w", err)
	}
	defer resp.Body.Close()

	var userInfo UserInfo

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Printf("Unable to decode user info response: %s\n", err)
		return nil, fmt.Errorf("unable to decode user info response: %w", err)
	}

	return &userInfo, nil
}
func (s *LoginService) FindUserInDb(ctx context.Context, userInfo *UserInfo) (*models.User, error) {
	var userProfile models.UserProfile
	result := s.db.Where("google_id = ?", userInfo.UserID).First(&userProfile)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Printf("User not found with Google ID: %s\n", userInfo.UserID)
			return nil, fmt.Errorf("user not found")
		}
		log.Printf("Database error when searching for user profile: %s\n", result.Error)
		return nil, fmt.Errorf("database error: %w", result.Error)
	}

	log.Printf("Found user profile: %+v\n", userProfile)

	var user models.User
	result = s.db.Where("id = ?", userProfile.UserID).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Printf("User not found with ID: %d\n", userProfile.UserID)
			return nil, fmt.Errorf("user not found")
		}
		log.Printf("Database error when searching for user: %s\n", result.Error)
		return nil, fmt.Errorf("database error: %w", result.Error)
	}

	return &user, nil
}
func (s *LoginService) CreateToken(user *models.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user.ID
	accessTokenExpireTime := time.Now().Add(time.Hour * 24).Unix()
	claims["exp"] = accessTokenExpireTime

	secret := os.Getenv("JWT_SECRET")
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Printf("Failed to sign token: %s\n", err)
		return "", echo.NewHTTPError(http.StatusInternalServerError, "Failed to sign token")
	}
	return tokenString, nil
}
func (s *LoginService) CreateCookie(c echo.Context, tokenString string, user *models.User) error {
	cookie := new(http.Cookie)
	cookie.Name = "Authorization"
	cookie.MaxAge = 24 * 60 * 60 * 30
	cookie.Value = tokenString
	cookie.Path = "/"
	cookie.HttpOnly = false
	c.SetCookie(cookie)

	log.Printf("Set cookie for user ID: %d\n", user.ID)
	return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/users/profile/%v", user.ID))
}

//	func HandleCallbackLogin(c echo.Context) error {
//		code := c.QueryParam("code")
//		if code == "" {
//			log.Println("No code in the URL")
//			return c.String(http.StatusBadRequest, "No code in the URL")
//		}
//
//		conf, ok := singleton.GetAndConvertSingleton[*oauth2.Config]("google_data_login")
//		if !ok {
//			return c.String(http.StatusInternalServerError, "singleton error")
//		}
//
//		log.Printf("Received code: %s\n", code)
//		ctx := context.Background()
//		tok, err := conf.Exchange(ctx, code)
//		if err != nil {
//			log.Printf("Unable to exchange code for token: %s\n", err)
//			return c.String(http.StatusInternalServerError, "Unable to exchange code for token: "+err.Error())
//		}
//
//		log.Printf("Received token: %v\n", tok)
//		client := conf.Client(ctx, tok)
//		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
//		if err != nil {
//			log.Printf("Unable to get user info: %s\n", err)
//			return c.String(http.StatusInternalServerError, "Unable to get user info: "+err.Error())
//		}
//		defer resp.Body.Close()
//
//		var userInfo struct {
//			Email  string `json:"email"`
//			UserID string `json:"id"`
//		}
//
//		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
//			log.Printf("Unable to decode user info response: %s\n", err)
//			return c.String(http.StatusInternalServerError, "Unable to decode user info response: "+err.Error())
//		}
//
//		log.Printf("Received user info: %+v\n", userInfo)
//		var userProfile models.UserProfile
//		result := db.Where("google_id = ?", userInfo.UserID).First(&userProfile)
//		if result.Error != nil {
//			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
//				log.Printf("User not found with Google ID: %s\n", userInfo.UserID)
//				return c.String(http.StatusUnauthorized, "User not found")
//			}
//			log.Printf("Database error when searching for user profile: %s\n", result.Error)
//			return c.String(http.StatusInternalServerError, "Database error: "+result.Error.Error())
//		}
//
//		log.Printf("Found user profile: %+v\n", userProfile)
//		var user models.User
//		result = db.Where("id = ?", userProfile.UserID).First(&user)
//		if result.Error != nil {
//			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
//				log.Printf("User not found with ID: %d\n", userProfile.UserID)
//				return c.String(http.StatusUnauthorized, "User not found")
//			}
//			log.Printf("Database error when searching for user: %s\n", result.Error)
//			return c.String(http.StatusInternalServerError, "Database error: "+result.Error.Error())
//		}
//
//		log.Printf("Found user: %+v\n", user)
//		token := jwt.New(jwt.SigningMethodHS256)
//		claims := token.Claims.(jwt.MapClaims)
//		claims["sub"] = user.ID
//		accessTokenExpireTime := time.Now().Add(time.Hour * 24).Unix()
//		claims["exp"] = accessTokenExpireTime
//
//		secret := os.Getenv("JWT_SECRET")
//		tokenString, err := token.SignedString([]byte(secret))
//		if err != nil {
//			log.Printf("Failed to sign token: %s\n", err)
//			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to sign token")
//		}
//
//		log.Printf("Generated JWT token: %s\n", tokenString)
//		cookie := new(http.Cookie)
//		cookie.Name = "Authorization"
//		cookie.MaxAge = 24 * 60 * 60 * 30
//		cookie.Value = tokenString
//		cookie.Path = "/"
//		cookie.HttpOnly = false
//		c.SetCookie(cookie)
//
//		log.Printf("Set cookie for user ID: %d\n", user.ID)
//		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/users/profile/%v", user.ID))
//	}
func HandleCallbackLogin(c echo.Context) error {
	service := NewLoginService(db)

	ctx := c.Request().Context()

	code, err := service.HandleCallbackloginCode(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Unable to handle callback login code: "+err.Error())
	}

	conf, err := service.getOAuthConfig()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Unable to get OAuth config: "+err.Error())
	}

	token, err := service.exchangeCodeForToken(ctx, conf, code)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Unable to exchange code for token: "+err.Error())
	}

	userInfo, err := service.fetchUserInfo(ctx, token, conf)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Unable to fetch user info: "+err.Error())
	}

	user, err := service.FindUserInDb(ctx, userInfo)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Unable to find user in DB: "+err.Error())
	}

	tokenString, err := service.CreateToken(user)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Unable to create token: "+err.Error())
	}

	err = service.CreateCookie(c, tokenString, user)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Unable to create cookie: "+err.Error())
	}

	return nil
}
func HandleHomeVKLogin(c echo.Context) error {
	conf, ok := singleton.GetAndConvertSingleton[*oauth2.Config]("vk_data_login")
	if !ok {
		return c.String(http.StatusInternalServerError, "singleton error")
	}
	logger.FastDebug(logcfg.DebugLoggerName, "This is conf struct: %v", conf)

	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)

	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func HandleCallbackVKLogin(c echo.Context) error {
	conf, ok := singleton.GetAndConvertSingleton[*oauth2.Config]("vk_data_login")
	if !ok {
		return c.String(http.StatusInternalServerError, "singleton error")
	}
	service := NewLoginService(db)
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
	if err = json.NewDecoder(resp.Body).Decode(&vkResponse); err != nil {
		return c.String(http.StatusInternalServerError, "Unable to decode user info response: "+err.Error())
	}

	if len(vkResponse.Response) == 0 {
		return c.String(http.StatusInternalServerError, "No user info found in the response")
	}

	var userProfile models.UserProfile
	result := db.Where("vk_id = ?", strconv.FormatInt(vkResponse.Response[0].ID, 10)).First(&userProfile)
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
	tokenString, err := service.CreateToken(&user)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Unable to create token: "+err.Error())
	}
	//token := jwt.New(jwt.SigningMethodHS256)
	//claims := token.Claims.(jwt.MapClaims)
	//claims["sub"] = user.ID
	//accessTokenExpireTime := time.Now().Add(time.Hour * 24).Unix()
	//claims["exp"] = accessTokenExpireTime
	//
	//secret := os.Getenv("JWT_SECRET")
	//tokenString, err := token.SignedString([]byte(secret))
	//if err != nil {
	//	log.Printf("Failed to sign token: %s", err)
	//	return echo.NewHTTPError(http.StatusInternalServerError, "Failed to sign token")
	//}

	err = service.CreateCookie(c, tokenString, &user)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Unable to create cookie: "+err.Error())
	}

	//cookie := new(http.Cookie)
	//cookie.Name = "Authorization"
	//cookie.MaxAge = 24 * 60 * 60 * 30
	//cookie.Value = tokenString
	//cookie.Path = "/"
	//cookie.HttpOnly = false
	//c.SetCookie(cookie)

	return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/users/profile/%v", user.ID))
}
func TGLogin(c echo.Context) error {
	service := NewLoginService(db)
	input := new(models.TgVerification)
	if err := c.Bind(input); err != nil {

		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	fmt.Println(input)
	check, err := TgVerificiation(input)
	if err != nil || !check {
		log.Printf("Failed to verify TG ID: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify TG ID")
	}
	fmt.Printf("tg id:%v\n, tg_id_type:%T\n", input.TgID, input.TgID)
	var userProfile models.UserProfile
	result := db.Where("tg_id = ?", strconv.Itoa(input.TgID)).First(&userProfile)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Printf("User not found with TG ID: %d\n", input.TgID)
			return c.String(http.StatusUnauthorized, "User not found")
		}
		log.Printf("Database error when searching for user profile: %s\n", result.Error)
		return c.String(http.StatusInternalServerError, "Database error: "+result.Error.Error())
	}

	log.Printf("Found user profile: %+v\n", userProfile)
	var user models.User
	result = db.Where("id = ?", userProfile.UserID).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Printf("User not found with ID: %d\n", userProfile.UserID)
			return c.String(http.StatusUnauthorized, "User not found")
		}
		log.Printf("Database error when searching for user: %s\n", result.Error)
		return c.String(http.StatusInternalServerError, "Database error: "+result.Error.Error())
	}

	log.Printf("Found user: %+v\n", user)
	tokenString, err := service.CreateToken(&user)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Unable to create token: "+err.Error())
	}
	log.Printf("Generated JWT token: %s\n", tokenString)

	err = service.CreateCookie(c, tokenString, &user)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Unable to create cookie: "+err.Error())
	}
	log.Printf("Set cookie for user ID: %d\n", user.ID)
	return c.JSON(http.StatusOK, user.ID)
}
