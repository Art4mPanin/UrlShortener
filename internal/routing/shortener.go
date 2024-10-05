package routing

import (
	"UrlShortener/internal/models"
	"UrlShortener/internal/mymiddleware"
	shorten "UrlShortener/pkg/shortener"
	"fmt"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"time"
)

func Shortener(c echo.Context) error {
	user := c.Get("user").(*models.User)

	var LI models.LinkInput
	if err := c.Bind(&LI); err != nil {
		log.Printf("Failed to bind link input: %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	fmt.Println(LI)
	var LD models.LinkDB

	LD.UserID = user.ID
	LD.LongLink = LI.Link
	LD.Availability = LI.Availability
	LD.UniqCounter = LI.UniqCounter
	LD.ShortLink = shorten.RandomString(5)
	switch LI.Duration {
	case "1day":
		LD.ExpiresAt = time.Now().Add(24 * time.Hour)
	case "7days":
		LD.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	case "30days":
		LD.ExpiresAt = time.Now().Add(30 * 24 * time.Hour)
	default:
		LD.ExpiresAt = time.Now().Add(24 * 365 * time.Hour)
	}
	switch LI.MaxRedirects {
	case -1:
		{
			LD.MaxRedirects = -1
		}
	case 0:
		{
			LD.MaxRedirects = 0
			log.Printf("The maximum redirect value is 0, therefore the link is unavailable")
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "The maximum redirect value is 0")
		}
	default:
		if LI.MaxRedirects > 0 {
			LD.MaxRedirects = LI.MaxRedirects
		} else {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid max redirects value")
		}
	}
	switch LI.TimeBeforeRedirect {
	case "immediately":
		{
			LD.TimeBeforeRedirect = 0
		}
	case "3sec":
		{
			LD.TimeBeforeRedirect = 3
		}
	case "5sec":
		{
			LD.TimeBeforeRedirect = 5
		}
	case "10sec":
		{
			LD.TimeBeforeRedirect = 10
		}
	case "15sec":
		{
			LD.TimeBeforeRedirect = 15
		}
	}

	if err := db.Save(&LD).Error; err != nil {
		log.Printf("Failed to save link: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save link")
	}
	baseURL := c.Request().Host
	shortURL := fmt.Sprintf("http://%s/s/%s", baseURL, LD.ShortLink)

	return c.JSON(http.StatusCreated, map[string]string{"shortenedLink": shortURL})
}
func ShortLinkHandler(c echo.Context) error {

	shortLink := c.Param("shortLink")

	var LD models.LinkDB
	if err := db.Where("short_link =?", shortLink).First(&LD).Error; err != nil {
		log.Printf("Failed to find link: %s", err)
		return echo.NewHTTPError(http.StatusNotFound, "Short link not found")
	}

	if LD.ExpiresAt.Before(time.Now()) {
		log.Printf("Link with ID %d has expired", LD.ID)
		return echo.NewHTTPError(http.StatusGone, "Link has expired")
	}

	if LD.MaxRedirects == 0 {
		log.Printf("Link with ID %d is blocked due to incorrect ammount of redirects avaliable", LD.ID)
		return echo.NewHTTPError(http.StatusTooManyRequests)
	}
	if LD.MaxRedirects == -1 {
		LD.MaxRedirects = -1
	} else if LD.MaxRedirects > 0 {
		LD.MaxRedirects = LD.MaxRedirects - 1
		if err := db.Save(&LD).Error; err != nil {
			log.Printf("Failed to update max redirects for link: %s", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update max redirects")
		}
	}
	if !LD.Availability {
		log.Printf("Link with ID %d is blocked due to unavailability", LD.ID)
		return echo.NewHTTPError(http.StatusForbidden, "Link is not available")
	}
	if LD.UniqCounter {
		clientIP := mymiddleware.GetRealIP(c)
		var ipLink models.IPLINK
		if err := db.Where("ip = ? AND link_id = ?", clientIP, LD.ID).First(&ipLink).Error; err == nil {
			log.Printf("Link with ID %d already visited by IP: %s", LD.ID, clientIP)
			return echo.NewHTTPError(http.StatusConflict, "Link already visited by this IP")
		} else if err = db.Create(&models.IPLINK{IP: clientIP, LinkID: LD.ID}).Error; err != nil {
			log.Printf("Failed to record IP visit for link: %s", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to record IP visit")
		}
	}

	var userProfile models.UserProfile
	if err := db.Where("user_id = ?", LD.UserID).First(&userProfile).Error; err != nil {
		log.Printf("Failed to find user profile: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "User profile not found")
	}

	profile := c.Get("profile").(*models.UserProfile)

	lastVisitDate := userProfile.LastVisitDate.Format("2006-01-02")
	currentDate := time.Now().Format("2006-01-02")

	userProfile.TotalRedirects++

	if profile.UserID == LD.UserID {
		userProfile.TotalRedirected++

		if lastVisitDate == currentDate {
			userProfile.DailyRedirects++
			userProfile.DailyRedirected++
		} else {
			userProfile.DailyRedirects = 1
			userProfile.DailyRedirected = 1
		}
	} else {
		if lastVisitDate == currentDate {
			userProfile.DailyRedirects++
		} else {
			userProfile.DailyRedirects = 1
		}
	}

	if err := db.Save(&userProfile).Error; err != nil {
		log.Printf("Failed to update user profile: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user profile")
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/redirect/%s", shortLink))
}
func RedirectPage(c echo.Context) error {
	shortLink := c.Param("shortLink")

	var LD models.LinkDB
	if err := db.Where("short_link = ?", shortLink).First(&LD).Error; err != nil {
		log.Printf("Failed to find link: %s", err)
		return echo.NewHTTPError(http.StatusNotFound, "Short link not found")
	}

	data := map[string]interface{}{
		"ShortLink":  shortLink,
		"LongLink":   LD.LongLink,
		"RedirectIn": LD.TimeBeforeRedirect,
	}

	return c.Render(http.StatusOK, "redirect_page.html", data)
}
func ShortenerPublic(c echo.Context) error {

	var LI models.LinkDBPublic
	if err := c.Bind(&LI); err != nil {
		log.Printf("Failed to bind link input: %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	var LD models.LinkDB

	LD.LongLink = LI.LongLink
	LD.Availability = true
	LD.UniqCounter = false
	LD.ShortLink = shorten.RandomString(5)
	switch LI.ExpiresAt {
	case "1day":
		LD.ExpiresAt = time.Now().Add(24 * time.Hour)
	case "7days":
		LD.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	case "30days":
		LD.ExpiresAt = time.Now().Add(30 * 24 * time.Hour)
	default:
		LD.ExpiresAt = time.Now().Add(24 * 365 * time.Hour)
	}
	LD.MaxRedirects = -1
	LD.TimeBeforeRedirect = 5
	fmt.Println("LD from unauthorized:", LD)
	if err := db.Save(&LD).Error; err != nil {
		log.Printf("Failed to save link: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save link")
	}
	baseURL := c.Request().Host
	shortURL := fmt.Sprintf("http://%s/s/%s", baseURL, LD.ShortLink)

	return c.JSON(http.StatusCreated, map[string]string{"shortenedLink": shortURL})
}
func ShortLinkPublicHandler(c echo.Context) error {

	shortLink := c.Param("shortLink")

	var LD models.LinkDB
	if err := db.Where("short_link =?", shortLink).First(&LD).Error; err != nil {
		log.Printf("Failed to find link: %s", err)
		return echo.NewHTTPError(http.StatusNotFound, "Short link not found")
	}
	if LD.ExpiresAt.Before(time.Now()) {
		log.Printf("Link with ID %d has expired", LD.ID)
		return echo.NewHTTPError(http.StatusGone, "Link has expired")
	}
	return c.Redirect(http.StatusFound, fmt.Sprintf("/redirect/%s", shortLink))
}
