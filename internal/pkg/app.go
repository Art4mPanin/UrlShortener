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
	"net/http"
	"os"
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
	newEnv := env.NewEnv(&config.EnvRoot)
	baseConfig := config.NewBaseConfig(newEnv)
	logger.InitLoggers(baseConfig)

	handler := singleton.InitializeSingletonHandler()
	handler.RegisterSingleton("google_data_profile", GetGoogleDataProfile())
	handler.RegisterSingleton("vk_data_profile", GetVKDataProfile())
	handler.RegisterSingleton("google_data_login", GetGoogleDataLogin())
	handler.RegisterSingleton("vk_data_login", GetVKDataLogin())
	handler.RegisterSingleton("global_db", database)

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
	e.GET("/auth/vk/login", routing.HandleHomeVKLogin)
	e.GET("/auth/vk/callback/login", routing.HandleCallbackVKLogin)
	e.POST("/users/tg_link/login", routing.TGLogin)

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
	e.GET("/auth/google", routing.HandleHome)
	e.GET("/auth/google/callback", routing.HandleCallback, mymiddleware.RequireAuth)
	e.PUT("/users/unlink_google/:id", routing.GoogleUnlink, mymiddleware.RequireAuth) //
	e.GET("/auth/vk/", routing.HandleHomeVK)
	e.GET("/auth/vk/callback", routing.HandleCallbackVK, mymiddleware.RequireAuth)
	e.PUT("/users/unlink_vk/:id", routing.VKUnlink, mymiddleware.RequireAuth)
	e.PUT("/users/tg_link/", routing.TGLINK, mymiddleware.RequireAuth)
	e.DELETE("/users/tg_unlink", routing.TgUnlink, mymiddleware.RequireAuth)

	e.GET("/api/url_shorten/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "shortener.html", nil)
	})
	e.POST("/api/shorten", routing.Shortener, mymiddleware.RequireAuth)
	e.GET("/s/:shortLink", routing.ShortLinkHandler, mymiddleware.RequireAuth)
	e.GET("/redirect/:shortLink", routing.RedirectPage, mymiddleware.RequireAuth)

	e.GET("/api/url_shorten_public/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "shortener_public.html", nil)
	})
	e.POST("/api/shorten_public", routing.ShortenerPublic)
	e.GET("/s/:shortLink", routing.ShortLinkPublicHandler)
	e.GET("/redirect/:shortLink", routing.RedirectPage)
	//e.Depression("/life")
	log.Fatal(e.Start(":8080"))
}
