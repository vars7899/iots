package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/pkg/response"
)

const (
	CSRFTokenCookie = "csrf_token"
	CSRFTokenHeader = "X-CSRF-TOKEN"
)

func CsrfMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if err := VerifyCSRFToken(c); err != nil {
				return err
			}
			return next(c)
		}
	}
}

func SetCSRFCookie(c echo.Context, token string) {
	cookie := &http.Cookie{
		Name:     CSRFTokenCookie,
		Value:    token,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		Secure:   config.IsUnderProduction(),
	}
	c.SetCookie(cookie)
}

func VerifyCSRFToken(c echo.Context) error {
	cookie, err := c.Cookie(CSRFTokenCookie)
	if err != nil {
		return response.ErrBadRequest.WithDetails(echo.Map{
			"error": "CSRF token not found",
		})
	}

	requestToken := c.Request().Header.Get(CSRFTokenHeader)
	if requestToken == "" {
		requestToken = c.FormValue(CSRFTokenHeader)
	}

	if requestToken == "" {
		return response.ErrBadRequest.WithDetails(echo.Map{
			"error": "CSRF token not provided",
		})
	}

	if cookie.Value != requestToken {
		return response.ErrForbidden.WithDetails(echo.Map{
			"error": "CSRF token mismatch",
		})
	}

	return nil
}

func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
