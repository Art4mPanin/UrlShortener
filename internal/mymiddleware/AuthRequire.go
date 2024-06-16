package mymiddleware

import (
	"UrlShortener/internal/models"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	"os"
	"time"
)

var db *gorm.DB

func InitializeDB(database *gorm.DB) {
	db = database
}

//	func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
//		return func(c echo.Context) error {
//			tokenstring, err := c.Cookie("Authorization")
//			if err != nil {
//				return echo.NewHTTPError(http.StatusUnauthorized)
//			}
//			token, err := jwt.Parse(tokenstring, func(token *jwt.Token) (interface{}, error) {
//				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
//					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
//				}
//				return []byte(os.Getenv("JWT_SECRET")), nil
//			})
//			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
//				if float64(time.Now().Unix()) > claims["exp"].(float64) {
//					return echo.NewHTTPError(http.StatusUnauthorized)
//				}
//				var user models.User
//				db.First(&user, claims["sub"])
//				if user.ID == 0 {
//					return echo.NewHTTPError(http.StatusUnauthorized)
//				}
//				c.Set("user", user)
//				return next(c)
//				fmt.Println(claims["sub"])
//			} else {
//				return echo.NewHTTPError(http.StatusUnauthorized)
//			}
//			// Здесь вы можете добавить свою логику проверки аутентификации
//			// Например, проверка токена или сессии пользователя
//			// Вызов следующего обработчика
//
//		}
//	}
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

			c.Set("user", user)
			fmt.Println("Authenticated user ID:", user.ID)
			return next(c)
		} else {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims")
		}
	}
}
