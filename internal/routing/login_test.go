package routing

import (
	"UrlShortener/pkg/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type MockUtils struct {
	mock.Mock
}

func (m *MockUtils) ExecuteTemplate(c echo.Context, templateFile string) error {
	args := m.Called(c, templateFile)
	return args.Error(0)
}

func TestLoginTPL(t *testing.T) {
	e := echo.New()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	c := e.NewContext(req, rec)

	mockUtils := new(MockUtils)
	mockUtils.On("ExecuteTemplate", c, "assets/templates/login.html").Return(nil)

	originalExecuteTemplate := utils.ExecuteTemplate
	utils.ExecuteTemplateFunc = mockUtils.ExecuteTemplate
	defer func() { utils.ExecuteTemplateFunc = originalExecuteTemplate }()

	err := LoginTPL(c)

	assert.NoError(t, err)
	mockUtils.AssertExpectations(t)
}
func TestValidate(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/validate", nil)
	c := e.NewContext(req, rec)
	expectedUser := "testuser"
	c.Set("user", expectedUser)
	err := Validate(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	expectedResponse := `{"message":"testuser"}`
	assert.JSONEq(t, expectedResponse, rec.Body.String())
}
func TestBind_Success(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	stringreq := `{"email":"example@example.com", "password":"password"}`
	req := httptest.NewRequest(http.MethodGet, "/login", strings.NewReader(stringreq))
	req.Header.Set("Content-Type", "application/json")
	c := e.NewContext(req, rec)
	err := LogIn(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
