package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgerrcode"
	"github.com/noctispine/blog/cmd/models"
	"github.com/noctispine/blog/pkg/responses"
	"github.com/noctispine/blog/pkg/utils"
	"github.com/noctispine/blog/pkg/wrappers"
	log "github.com/sirupsen/logrus"
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
		if utils.CheckPostgreError(err, pgerrcode.UniqueViolation){
			responses.AbortWithStatusJSONError(c, http.StatusBadRequest, wrappers.NewErrAlreadyExists("category"))
			return
		}
		
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *PostCategoryHandler) BatchCreate(c *gin.Context) {
	var postId string
	var ok bool

	if postId, ok = c.GetQuery("postId"); !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	postIdInt, err := strconv.ParseInt(postId, 10, 64)

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}


	postCategoryBatch := struct {
		CategoryIds []int64 `json:"categoryIds"`
	}{}

	if err := c.ShouldBindJSON(&postCategoryBatch); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	categoryIds := postCategoryBatch.CategoryIds

	var newPostCategories []models.PostCategory

	for _, categoryId := range categoryIds {
		newPostCategories = append(newPostCategories, models.PostCategory{PostID: postIdInt, CategoryID: categoryId})
	}

	if err := h.db.Table("post_category").Create(&newPostCategories).Error; err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
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
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)

}
