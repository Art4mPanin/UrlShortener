package pkg

import (
	"UrlShortener/internal/mymiddleware"
	"UrlShortener/internal/routing"
	"UrlShortener/internal/storage/users"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
	"html/template"
	"io"
	"log"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

var db *gorm.DB

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func InitServer() {
	users.InitDB()
	//routing.Storeconfig()

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Validator = &CustomValidator{validator: validator.New()}
	e.GET("/users/signup/", routing.SignUpTPL)
	e.POST("/users/signup/", routing.SignUp)

	e.GET("/users/login/", routing.LoginTPL)
	e.POST("/users/login/", routing.LogIn)
	e.GET("/auth/google/login", routing.HandleHomeLogin)
	e.GET("/auth/google/callback/login", routing.HandleCallbackLogin)

	e.GET("/validate/", routing.Validate, mymiddleware.RequireAuth)
	e.Static("/assets", "assets")
	e.Static("/uploads", "uploads")

	e.GET("/api/users/profile/:id", routing.ProfileHandler, mymiddleware.RequireAuth)
	e.GET("/users/profile/:id", routing.ProfileTPL, mymiddleware.RequireAuth)
	e.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseGlob("assets/templates/*.html")),
	}
	e.PUT("/users/update-avatar/:id", routing.AvatarUpdate, mymiddleware.RequireAuth)
	e.PUT("/users/update-data/:id", routing.DataUpdate, mymiddleware.RequireAuth)
	e.PUT("/users/update-password/:id", routing.PassUpdate, mymiddleware.RequireAuth)
	e.GET("/users/verification/", routing.VerificationTPL)
	e.POST("/users/verification/", routing.Verification)
	e.GET("/users/password-reset/", routing.PasswordResetTPL)
	e.POST("/users/password-reset/", routing.SendCodePass)
	e.GET("/users/password-reset-new-pass/", routing.PasswordResetSetNewPassTPL)
	//e.GET("/", routing.Homepage)
	e.GET("/auth/google", routing.HandleHome)
	e.GET("/auth/google/callback", routing.HandleCallback, mymiddleware.RequireAuth)
	e.PUT("/users/unlink_google/:id", routing.GoogleUnlink, mymiddleware.RequireAuth) //

	//e.Depression("/life")
	log.Fatal(e.Start(":8080"))
}
