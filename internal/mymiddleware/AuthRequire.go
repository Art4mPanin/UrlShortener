package mymiddleware

import (
	"UrlShortener/internal/models"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"time"
)

var db *gorm.DB

func InitializeDB(database *gorm.DB) {
	db = database
}

func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("Authorization")
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Missing or invalid cookie")
		}

		tokenString := cookie.Value
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if exp, ok := claims["exp"].(float64); ok {
				if float64(time.Now().Unix()) > exp {
					return echo.NewHTTPError(http.StatusUnauthorized, "Token expired")
				}
			} else {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims")
			}

			sub, ok := claims["sub"].(float64)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token subject")
			}

			var user models.User
			db.First(&user, uint(sub))
			if user.ID == 0 {
				return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
			}
			result := db.Preload("Profiles").First(&user, user.ID)
			if result.Error != nil {
				if errors.Is(result.Error, gorm.ErrRecordNotFound) {
					return echo.NewHTTPError(http.StatusNotFound, "User not found")
				}
				return echo.NewHTTPError(http.StatusBadRequest, "Incorrect ID")
			}

			if len(user.Profiles) == 0 {
				userProfile := models.UserProfile{
					UserID: user.ID,
				}
				db.Create(&userProfile)
				user.Profiles = append(user.Profiles, userProfile)
			}
			c.Set("profile", &user.Profiles[0])
			c.Set("user", &user)
			c.Set("token", token)
			fmt.Println("Authenticated user ID:", user.ID)
			return UpdateTimeIp(next)(c)
		} else {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims")
		}
	}
}
func UpdateTimeIp(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Получаем пользователя из контекста
		user := c.Get("user").(*models.User)

		// Обновляем время и IP
		var userProfile models.UserProfile
		result := db.Where("user_id = ?", user.ID).First(&userProfile)
		if result.Error != nil {
			log.Printf("User profile not found: %s", result.Error)
			return echo.NewHTTPError(http.StatusNotFound, "User profile not found")
		}
		userProfile.LastVisitDate = time.Now()
		userProfile.LastIP = getRealIP(c)

		if err := db.Save(&userProfile).Error; err != nil {
			log.Printf("Failed to update user profile: %s", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user profile")
		}

		log.Printf("User profile with ID %d successfully updated", user.ID)

		// Выполняем следующий обработчик
		return next(c)
	}

}
func getRealIP(c echo.Context) string {
	realIP := c.Request().Header.Get("X-Real-IP")
	if realIP == "" {
		realIP = c.Request().Header.Get("X-Forwarded-For")
	}
	if realIP == "" {
		realIP = c.Request().RemoteAddr
	}
	return realIP
}
