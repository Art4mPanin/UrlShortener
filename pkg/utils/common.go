package utils

import (
	"github.com/labstack/echo/v4"
	"html/template"
	"net/http"
	"time"
)

func DoWithTries(f func() error, attempts int, delay time.Duration) error {
	var err error
	for attempts > 0 {
		if err = f(); err == nil {
			return nil
		}
		time.Sleep(delay)
		attempts--
	}
	return err
}
func ExecuteTemplate(c echo.Context, templateFile string) error {
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could not load template: "+err.Error())
	}

	// Исполняем шаблон
	err = tmpl.Execute(c.Response(), nil)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could not execute template: "+err.Error())
	}

	return nil
}
