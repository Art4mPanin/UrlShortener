package utils

import (
	"github.com/labstack/echo/v4"
	"html/template"
	"net/http"
)

var ExecuteTemplateFunc = ExecuteTemplate

func ExecuteTemplate(c echo.Context, templateFile string) error {
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could not load template: "+err.Error())
	}

	err = tmpl.Execute(c.Response(), nil)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could not execute template: "+err.Error())
	}

	return nil
}
