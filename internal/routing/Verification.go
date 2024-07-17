package routing

import (
	"UrlShortener/internal/models"
	"UrlShortener/pkg/utils"
	"fmt"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"time"
)

// c.bind
// entered code = code in db; entered code<= 30 min
// registration successfully, IsActive = true
func VerificationTPL(c echo.Context) error {
	Tplname := "assets/templates/verification.html"
	return utils.ExecuteTemplate(c, Tplname)
}
func Verification(c echo.Context) error {
	input := new(models.VerificationUserInput)
	err := c.Bind(input)
	if err != nil {
		log.Printf("Binding error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	log.Printf("Received input: %+v", input)
	var verification models.Verification
	fmt.Println(input)
	result := db.Where("user_email = ?", input.Email).First(&verification)
	if result.Error != nil {
		log.Printf("User profile not found: %s", result.Error)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Incorrect code. A new code has been sent to your email.",
		})
	}
	if input.Code != verification.Code {
		log.Printf("Incorrect code: got %d, expected %d", input.Code, verification.Code)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Code has expired. A new code has been sent to your email.",
		})
	}
	if time.Since(verification.Timestamp) > 30*time.Minute {
		log.Printf("Code expired: timestamp %v", verification.Timestamp)
		return c.JSON(http.StatusBadRequest, "Code has expired. A new code has been sent to your email.")
	}
	var user models.User
	if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		log.Printf("User not found: %s", err.Error())
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}
	user.IsActive = true
	db.Save(&user)
	//add row deleting
	db.Where("user_email =?", input.Email).Delete(&verification)
	log.Printf("User with email %s has been verified", input.Email)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Verification successfully",
	})
}
