package auth

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Query("token")
		if tokenString == "" {
			tokenString = c.PostForm("token")
		}
		// Token is missing, returns with error code 401 Unauthorized
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, "API token required")
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			SECRET_KEY, err := getSecretKey()
			if err != nil {
				panic(err)
			}
			return []byte(SECRET_KEY), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, "Invalid API token")
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Store user information from token into context.
			c.Set("userId", claims["userId"])
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, "Invalid API token")
			c.Abort()
		}
	}
}

func getSecretKey() (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", err
	}
	return os.Getenv("SECRET_KEY"), nil
}
