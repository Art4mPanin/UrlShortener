package routing

import (
	"UrlShortener/internal/models"
	"UrlShortener/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"io"
	"log"
	"net/http"
	"os"
)

func ProfileTPL(c echo.Context) error {
	templateFile := "assets/templates/profile.html"
	return utils.ExecuteTemplate(c, templateFile)
}
func AvatarUpdate(c echo.Context) error {
	// Получение файла
	file, err := c.FormFile("avatar")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid file upload")
	}

	// Открытие файла
	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error opening file")
	}
	defer src.Close()

	// Определение пути для сохранения файла
	dstPath := "uploads/" + file.Filename

	// Создание файла на сервере (в новой директории)
	dst, err := os.Create(dstPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error saving file")
	}
	defer dst.Close()

	// Копирование содержимого файла
	if _, err = io.Copy(dst, src); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error saving file")
	}

	// Получение ID пользователя из токена
	userID := c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["sub"].(float64)

	// Создание URL для файла
	avatarURL := "http://localhost:8080/uploads/" + file.Filename

	// Обновление URL аватара в базе данных
	result := db.Model(&models.UserProfile{}).Where("user_id = ?", userID).Update("avatar_url", avatarURL)
	if result.Error != nil {
		log.Printf("Error updating avatar URL in database: %s", result.Error)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error updating database")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message":    "File uploaded successfully",
		"avatar_url": avatarURL,
	})
}
