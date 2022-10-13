package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/noctispine/blog/cmd/models"
	"gorm.io/gorm"
)


type PostCategoryHandler struct {
	ctx context.Context
	db *gorm.DB
}

func NewPostCategoryHandler(ctx context.Context, db *gorm.DB) *PostCategoryHandler {
	return &PostCategoryHandler{
		ctx,
		db,
	}
}

func (h *PostCategoryHandler) Create(c *gin.Context) {
	var categoryId, postId string
	var ok bool

	if categoryId, ok = c.GetQuery("categoryId"); !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if postId, ok = c.GetQuery("postId"); !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := h.db.Table("post_category").Create(map[string]interface{}{
		"post_id": postId, "category_id": categoryId,
	  }).Error; err != nil {
		log.Println(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *PostCategoryHandler) Delete(c *gin.Context) {
	var categoryId, postId string
	var ok bool

	if categoryId, ok = c.GetQuery("categoryId"); !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if postId, ok = c.GetQuery("postId"); !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}


	if err := h.db.Table("post_category").Where("post_id = ? AND category_id = ?", postId, categoryId).Delete(&models.PostCategory{}).Error; err != nil {
		log.Println(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)

}



// func (h *PostCategoryHandler) Delete(c *gin.Context) {
// 	postCategoryId := c.Params.ByName("id")
// 	if postCategoryId == "" {
// 		c.AbortWithStatus(http.StatusBadRequest)
// 		return
// 	}

// 	result := h.db.Model(&models.PostCategory{}).Delete(&models.PostCategory{}, postCategoryId)
	
// 	if result.RowsAffected == 0 {
// 		responses.AbortWithStatusJSONError(c, http.StatusNotFound, wrappers.NewErrNotFound("post-category"))
// 		return
// 	}

// 	if result.Error != nil {
// 		c.AbortWithStatus(http.StatusInternalServerError)
// 		return
// 	}
	
// 	c.Status(http.StatusNoContent)
// }