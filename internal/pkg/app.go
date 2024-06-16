package pkg

import (
	"UrlShortener/internal/mymiddleware"
	"UrlShortener/internal/routing"
	"UrlShortener/internal/storage/users"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func InitServer() {
	users.InitDB()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	//
	//// Добавление кастомного валидатора
	e.Validator = &CustomValidator{validator: validator.New()}
	e.GET("/users/signup/", routing.SignUpTPL)
	e.POST("/users/signup/", routing.SignUp)

	e.GET("/users/login/", routing.LoginTPL)
	e.POST("/users/login/", routing.LogIn)
	e.GET("/validate", routing.Validate, mymiddleware.RequireAuth)
	e.Static("/assets", "assets")
	e.GET("/users/profile/:id", routing.ProfileTPL)
	e.POST("/user/update-avatar/", routing.AvatarUpdate)
	log.Fatal(e.Start(":8080"))
}
