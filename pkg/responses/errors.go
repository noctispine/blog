package responses

import (
	"github.com/gin-gonic/gin"
	"github.com/noctispine/blog/internal/translateErrors"
)

func AbortWithStatusJSONError(c *gin.Context, code int, err error) {
    c.AbortWithStatusJSON(code , gin.H{
		"error": err.Error()})
}

func AbortWithStatusJSONValidationErrors(c *gin.Context, code int, err error) {
	errs := translateErrors.TranslateError(err, translateErrors.EnTrans)
    c.AbortWithStatusJSON(code , gin.H{
		"errors": translateErrors.StringfyJSONErrArr(errs)})
}



