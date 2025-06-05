package middleware

import (
	"strings"
	"v01_system_backend/apps/utils"

	"github.com/labstack/echo/v4"
)

func JWTAuth(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return utils.UnauthorizedResponse(c, "Missing authorization header")
			}

			tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
			claims, err := utils.ValidateToken(tokenString, secret)
			if err != nil {
				return utils.UnauthorizedResponse(c, "Invalid token")
			}

			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("email", claims.Email)

			return next(c)
		}
	}
}
