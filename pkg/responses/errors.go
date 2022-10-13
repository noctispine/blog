package responses

import (
	"github.com/gin-gonic/gin"
)

func AbortWithStatusJSONError(c *gin.Context, code int, err error) {
    c.AbortWithStatusJSON(code , gin.H{
		"error": err.Error()})
}

// func AbortWithStatusJSONValidationErrors(c *gin.Context, code int, err error) bool {
// 	errs := TranslateError(err, enTrans)

//     c.AbortWithStatusJSON(code , gin.H{
// 		"errors": stringfyJSONErrArr(errs)})

// }

