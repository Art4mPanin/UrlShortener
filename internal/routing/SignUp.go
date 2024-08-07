package routing

import (
	"UrlShortener/internal/models"
	"UrlShortener/pkg/utils"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"
)

func SetDB(database *gorm.DB) {
	db = database
}
func SignUpTPL(c echo.Context) error {
	Tplname := "assets/templates/signup.html"
	return utils.ExecuteTemplate(c, Tplname)
}
func SignUp(c echo.Context) error {

	input := new(models.UserInput)
	if err := c.Bind(input); err != nil {
		log.Printf("Failed to bind user input: %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	if err := c.Validate(input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to hash password")
	}
	user := models.User{
		Username:       input.Username,
		Email:          input.Email,
		HashedPassword: string(hashedPassword),
		BirthDate:      input.BirthDate,
		IsActive:       false,
		IsAdmin:        false,
		RegisterDate:   time.Now(),
	}

	result := db.Create(&user)
	if result.Error != nil {
		log.Printf("Failed to insert user: %s", result.Error)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to insert user")
	} else {
		log.Printf("User successfully inserted with ID: %d", user.ID)
	}
	code, tstmp := utils.VerificationCode(input.Email)
	Verif := models.Verification{
		UserEmail: input.Email,
		Code:      code,
		Timestamp: tstmp,
	}
	//no vibes
	result1 := db.Create(&Verif)
	if result1.Error != nil {
		log.Printf("Failed to insert verification: %s", result1.Error)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to insert verification")
	}
	return c.JSON(http.StatusCreated, user)
}
