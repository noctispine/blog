package scopes

import (
	"math"

	"github.com/gin-gonic/gin"
	"github.com/noctispine/blog/pkg/constants/keys"
	"github.com/noctispine/blog/pkg/pagination"
	"gorm.io/gorm"
)
func Paginate(value interface{}, pagination *pagination.Pagination, db *gorm.DB , c *gin.Context) func (db *gorm.DB) *gorm.DB {
	pagination.Page = c.GetInt(keys.PageKey)
	pagination.Limit = c.GetInt(keys.PageSizeKey)
	
	var totalRows int64
	db.Model(value).Count(&totalRows)
	pagination.TotalRows = totalRows

	totalPages := int(math.Ceil(float64(totalRows) / float64(pagination.Limit)))
	pagination.TotalPages = totalPages

	return func (db *gorm.DB) *gorm.DB {
		return db.Offset(pagination.GetOffset()).Limit(pagination.GetLimit()).Order(pagination.GetSort())
	}
}