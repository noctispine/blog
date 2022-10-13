package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgerrcode"
	"github.com/noctispine/blog/cmd/models"
	"github.com/noctispine/blog/pkg/utils"
	"gorm.io/gorm"
)

type TagHandler struct {
	ctx context.Context
	db *gorm.DB
}

func NewTagHandler(ctx context.Context, db *gorm.DB) *TagHandler {
	return &TagHandler{
		ctx,
		db,
	}
}

func (h *TagHandler) GetAll(c *gin.Context) {
	var tags []models.Tag

	result := h.db.Find(&tags); 

	if result.RowsAffected == 0 {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	
	if result.Error != nil {
		log.Println(result.Error.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, tags)
}

func (h *TagHandler) Create(c *gin.Context) {
	var newTag models.Tag

	if err := c.ShouldBindJSON(&newTag); err != nil {
		log.Println(err.Error())
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := validate.Struct(&newTag); err != nil {
		errs := translateError(err, enTrans)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"errors": stringfyJSONErrArr(errs)})
		return
	}

	newTag.Slug = utils.ConstructSlug(newTag.Title)

	if err := h.db.Omit("id").Save(&newTag).Error; err != nil {
		if utils.CheckPostgreError(err, pgerrcode.UniqueViolation) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "this tag already exists"})
			return
		}

		log.Println(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *TagHandler) Delete(c *gin.Context) {
	tagId := c.Params.ByName("id")
	if tagId == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := h.db.Delete(&models.Tag{}, tagId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "tag doesnt exist"})
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TagHandler) Update(c *gin.Context) {
	var updateTag models.Tag

	if err := c.ShouldBindJSON(&updateTag); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := validate.Struct(&updateTag); err != nil {
		errs := translateError(err, enTrans)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"errors": errs})
		return
	}

	var tag models.Tag
	if err := h.db.Where("id", updateTag.ID).First(&tag).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "Tag doesnt exist"})
			return
		}

		log.Println(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if updateTag.Title != tag.Title {
		updateTag.Slug = utils.ConstructSlug(updateTag.Title)
	}

	if err := h.db.Where("id = ?", updateTag.ID).Updates(&updateTag).Error; err != nil {
		log.Println(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)

}