package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgerrcode"
	"github.com/noctispine/blog/cmd/models"
	"github.com/noctispine/blog/pkg/constants/keys"
	"github.com/noctispine/blog/pkg/pagination"
	"github.com/noctispine/blog/pkg/responses"
	"github.com/noctispine/blog/pkg/scopes"
	"github.com/noctispine/blog/pkg/utils"
	"github.com/noctispine/blog/pkg/wrappers"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PostHandler struct {
	ctx context.Context
	db *gorm.DB
}

func NewPostHandler(ctx context.Context, db *gorm.DB) *PostHandler {
	return &PostHandler{
		ctx: ctx,
		db: db,
	}
}

func (h *PostHandler) GetAll(c *gin.Context) {
	var posts []models.Post
	result := h.db.Order("created_at desc").Find(&posts)
	
	if result.RowsAffected == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	if result.Error != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Error(result.Error)
		return
	}

	c.JSON(http.StatusOK, &posts)
}

func (h *PostHandler) GetPage(c *gin.Context) {
	var posts []models.Post
	var pagination pagination.Pagination	

	result := h.db.Scopes(scopes.Paginate(posts, &pagination, h.db, c)).Find(&posts)
	if result.RowsAffected == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	if result.Error != nil {
		log.Error(result.Error)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	pagination.Rows = posts
	c.JSON(http.StatusOK, pagination)
}

func (h *PostHandler) GetPageByCategory(c *gin.Context) {
	var posts []models.Post
	var pagination pagination.Pagination

	categoryId := c.Params.ByName("id")
	if categoryId == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	h.db.Scopes(scopes.Paginate(posts, &pagination, h.db, c)).Joins("post_category", h.db.Model("post_category").Where("category_id = ?", categoryId)).Find(&posts)
	log.Println(posts)
}

func (h *PostHandler) Create(c *gin.Context) {
	var post models.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	post.UserID = c.GetInt64("userId")

	if err := validate.Struct(post); err != nil {
		responses.AbortWithStatusJSONValidationErrors(c, http.StatusBadRequest, err)
		return
	}

	post.Slug = utils.ConstructSlug(post.Title)

	if err := h.db.Omit("id", "created_at", "updated_at", "published_at", "is_published").Create(&post).Error; err != nil {
		if utils.CheckPostgreError(err, pgerrcode.UniqueViolation) {
			responses.AbortWithStatusJSONError(c, http.StatusBadRequest, wrappers.NewErrAlreadyExists("post"))
			return
		}
		
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, post)
}

func (h *PostHandler) Update(c *gin.Context) {
	var updatePost models.Post


	if err := c.ShouldBindJSON(&updatePost); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := validate.Struct(updatePost); err != nil {
		responses.AbortWithStatusJSONValidationErrors(c, http.StatusBadRequest, err)
		return
	}

	var post models.Post
	if err := h.db.Model(&models.Post{}).Where("id = ?", updatePost.ID).First(&post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			responses.AbortWithStatusJSONError(c, http.StatusNotFound, wrappers.NewErrDoesNotExist("post"))
			return
		}

		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	updatePost.UpdatedAt = time.Now()

	if post.Title != updatePost.Title {
		updatePost.Slug = utils.ConstructSlug(updatePost.Title)
	}
	
	if err := h.db.Model(&models.Post{}).Where("user_id = ?", c.GetInt64(keys.UserID)).Omit("id", "created_at", "user_id").Updates(&updatePost).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			responses.AbortWithStatusJSONError(c, http.StatusNotFound, wrappers.NewErrDoesNotExist("post"))
			return
		}

		c.AbortWithStatus(http.StatusInternalServerError)
		log.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *PostHandler) TogglePublish(c *gin.Context) {
	postId := c.Params.ByName("id")

	if postId == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var post models.Post
	if err := h.db.Model(&models.Post{}).Where("user_id = ? AND id = ?", c.GetInt64(keys.UserID), postId).First(&post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			responses.AbortWithStatusJSONError(c, http.StatusNotFound, wrappers.NewErrDoesNotExist("post"))
			return
		}

		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if err := h.db.Model(&models.Post{}).Where("user_id = ? AND id = ?", c.GetInt64(keys.UserID), postId).Update("is_published", !post.IsPublished).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			responses.AbortWithStatusJSONError(c, http.StatusBadRequest, wrappers.NewErrNotFound("post"))
			return
		}

		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func (h *PostHandler) Delete(c *gin.Context) {
	postId := c.Params.ByName("id")
	
	if postId == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	
	result := h.db.Where("user_id = ?", c.GetInt64(keys.UserID)).Delete(&models.Post{}, postId)

	if result.RowsAffected == 0 {
		responses.AbortWithStatusJSONError(c, http.StatusNotFound, wrappers.NewErrNotFound("post"))
		return
	}

	if result.Error != nil {
		log.Error(result.Error)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	
	c.Status(http.StatusNoContent)
}