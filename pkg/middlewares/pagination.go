package middlewares

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/noctispine/blog/pkg/constants/keys"
)

func Pagination() gin.HandlerFunc {
	return func(c *gin.Context) {
		Page, isPageOk := c.GetQuery(keys.PageKey)
		PageSize, isPageSizeOk := c.GetQuery(keys.PageSizeKey)
		if !isPageOk || !isPageSizeOk {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		intPage := 0
		intPageSize := 0
		var err error
		if Page != "" || PageSize != "" {
			if intPage, err = strconv.Atoi(Page); err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}

			if intPageSize, err = strconv.Atoi(PageSize); err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
		}

		c.Set(keys.PageKey, intPage)
		c.Set(keys.PageSizeKey, intPageSize)
		c.Next()
	}
} 
