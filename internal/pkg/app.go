package pkg

import (
	"UrlShortener/internal/routing"
	"UrlShortener/internal/storage/users"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"log"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func InitServer() {
	// Инициализация базы данных
	users.InitDB()

	// Создание Echo instance
	e := echo.New()
	//e.Use(middleware.Logger())
	//e.Use(middleware.Recover())
	//
	//// Добавление кастомного валидатора
	e.Validator = &CustomValidator{validator: validator.New()}

	e.POST("/users/signup/", routing.SignUp)
	e.POST("/users/login/", routing.LogIn)

	log.Fatal(e.Start(":8080"))
}

//func getUsersHandler(c echo.Context) error {
//	return c.String(http.StatusOK, "Get Users Handler")
//}
