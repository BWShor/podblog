package csp

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/labstack/echo/v4"
)

const nonceKey = "cspNonce"

func NewNonce() (string, error) {
	b := make([]byte, 16) // 128 bits
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 1. Generate nonce
			n, err := NewNonce()
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "nonce generation failed")
			}

			// 2. Store it in the context
			c.Set(nonceKey, n)

			// 3. Build CSP header
			csp := "default-src 'self'; " +
				"script-src 'self' 'nonce-" + n + "' https://cdn.jsdelivr.net; " +
				"style-src 'self' 'nonce-" + n + "' https://cdn.jsdelivr.net; " +
				"img-src 'self' data:; " +
				"font-src 'self'; " +
				"connect-src 'self'; " +
				"frame-ancestors 'none'; " +
				"base-uri 'self'; " +
				"object-src 'none';"

			c.Response().Header().Set("Content-Security-Policy", csp)

			// 4. Call next handler
			return next(c)
		}
	}
}
