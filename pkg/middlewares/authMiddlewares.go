package middlewares

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/noctispine/blog/cmd/handlers"
	"github.com/noctispine/blog/pkg/constants/keys"
	"github.com/noctispine/blog/pkg/utils"
)
func ValidateToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenValue := c.GetHeader("Authorization")
		claims := &handlers.Claims{}

		tkn, err := jwt.ParseWithClaims(tokenValue, claims,
			func(token *jwt.Token) (interface{}, error){
				return []byte(os.Getenv("JWT_SECRET")), nil
		})


		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if tkn == nil || !tkn.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}


		c.Set(keys.UserID, claims.UserID)
		c.Set(keys.UserRole, claims.Role)
		c.Next()
	}
}

func Authorization(avaliableRoles []int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if utils.Contains(avaliableRoles, c.GetInt(keys.UserRole)){
			c.Next()
			return
		}

		c.AbortWithStatus(http.StatusUnauthorized)
	}
}