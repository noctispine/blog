package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgerrcode"
	"github.com/noctispine/blog/cmd/models"
	"github.com/noctispine/blog/pkg/responses"
	"github.com/noctispine/blog/pkg/utils"
	"github.com/noctispine/blog/pkg/wrappers"
	log "github.com/sirupsen/logrus"
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
		log.Error(result.Error)
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
		responses.AbortWithStatusJSONValidationErrors(c, http.StatusBadRequest, err)
		return
	}

	newCategory.Slug = utils.ConstructSlug(newCategory.Title)

	if err := h.db.Omit("id").Save(&newCategory).Error; err != nil {
		if utils.CheckPostgreError(err, pgerrcode.UniqueViolation) {
			responses.AbortWithStatusJSONError(c, http.StatusBadRequest, wrappers.NewErrAlreadyExists("category"))
			return
		}

		log.Error(err)
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

	result := h.db.Delete(&models.Category{}, categoryId); 

	if result.RowsAffected == 0 {
		responses.AbortWithStatusJSONError(c, http.StatusBadRequest, wrappers.NewErrNotFound("category"))
		return
	}
	
	if result.Error != nil {
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
		responses.AbortWithStatusJSONValidationErrors(c, http.StatusBadRequest, err)
		return
	}

	var category models.Category
	if err := h.db.Model(&models.Category{}).Where("id = ?", updateCategory.ID).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			responses.AbortWithStatusJSONError(c, http.StatusBadRequest, wrappers.NewErrDoesNotExist("category"))
			return
		}

		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if category.Title != updateCategory.Title {
		updateCategory.Slug = utils.ConstructSlug(updateCategory.Title)
	}

	if err := h.db.Model(&models.Category{}).Where("id = ?", updateCategory.ID).Omit("id").Updates(&updateCategory).Error; err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}