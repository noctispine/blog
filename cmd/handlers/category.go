package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgerrcode"
	"github.com/noctispine/blog/cmd/models"
	"github.com/noctispine/blog/pkg/utils"
	"github.com/noctispine/blog/pkg/wrappers"
	"gorm.io/gorm"
)

type CategoryHandler struct {
	ctx context.Context
	db *gorm.DB
}

func NewCategoryHandler(ctx context.Context, db *gorm.DB) *CategoryHandler {
	return &CategoryHandler{
		ctx,
		db,
	}
}

func (h *CategoryHandler) GetAll(c *gin.Context) {
	var categories []models.Category
	result := h.db.Order("id desc").Find(&categories)

	if result.RowsAffected == 0 {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	if result.Error != nil {
		log.Println(result.Error.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, categories)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var newCategory models.Category
	
	if err := c.ShouldBindJSON(&newCategory); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := validate.Struct(newCategory); err != nil {
		errs := translateError(err, enTrans)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"errors": stringfyJSONErrArr(errs)})
		return
	}

	newCategory.Slug = utils.ConstructSlug(newCategory.Title)

	if err := h.db.Omit("id").Save(&newCategory).Error; err != nil {
		if utils.CheckPostgreError(err, pgerrcode.UniqueViolation) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "The category already exists"})
			return
		}

		log.Println(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	categoryId := c.Params.ByName("id")
	if categoryId == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := h.db.Delete(&models.Category{}, categoryId).Error; err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	var updateCategory models.Category

	if err := c.ShouldBindJSON(&updateCategory); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := validate.Struct(updateCategory); err != nil {
		errs := translateError(err, enTrans)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": stringfyJSONErrArr(errs)})
		return
	}

	var category models.Category
	if err := h.db.Model(&models.Category{}).Where("id = ?", updateCategory.ID).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": wrappers.NewErrDoesNotExist("category").Error()})
			return
		}

		log.Println(err.Error())
		c.AbortWithStatus(http.StatusBadGateway)
		return
	}

	if category.Title != updateCategory.Title {
		updateCategory.Slug = utils.ConstructSlug(updateCategory.Title)
	}

	if err := h.db.Model(&models.Category{}).Where("id = ?", updateCategory.ID).Omit("id").Updates(&updateCategory).Error; err != nil {
		log.Println(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}