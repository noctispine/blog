package middlewares

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/noctispine/blog/cmd/models"
	"github.com/noctispine/blog/pkg/constants/keys"
	"github.com/noctispine/blog/pkg/responses"
	"github.com/noctispine/blog/pkg/wrappers"
	"gorm.io/gorm"
)

func PostMatchesWithUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var post models.Post
		var postId string
		var ok bool
		
		if postId, ok = c.GetQuery("postId"); !ok{
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if err := db.First(&post, postId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				responses.AbortWithStatusJSONError(c, http.StatusNotFound, wrappers.NewErrNotFound("post"))
				return
			}
	
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	
		if post.UserID != c.GetInt64(keys.UserID) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	}
}