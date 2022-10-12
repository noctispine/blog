package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgerrcode"
	"github.com/noctispine/blog/cmd/models"
	"github.com/noctispine/blog/pkg/constants/keys"
	"github.com/noctispine/blog/pkg/pagination"
	"github.com/noctispine/blog/pkg/scopes"
	"github.com/noctispine/blog/pkg/utils"
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error()})
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
		log.Println(result.Error)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	pagination.Rows = posts
	c.JSON(http.StatusOK, pagination)
}

func (h *PostHandler) Create(c *gin.Context) {
	var post models.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}

	post.UserID = c.GetInt64("userId")

	if err := validate.Struct(post); err != nil {
		errs := translateError(err, enTrans)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": stringfyJSONErrArr(errs)})
		return
	}

	post.Slug = utils.ConstructSlug(post.Title)

	if err := h.db.Omit("id", "parent_id", "created_at", "updated_at", "published_at", "is_published").Create(&post).Error; err != nil {
		if utils.CheckPostgreError(err, pgerrcode.UniqueViolation) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "The title is already exists"})
			return
		}
		
		log.Println(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, post)
}

func (h *PostHandler) Update(c *gin.Context) {
	var updatePost models.Post


	if err := c.ShouldBindJSON(&updatePost); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}

	if err := validate.Struct(updatePost); err != nil {
		errs := translateError(err, enTrans)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": stringfyJSONErrArr(errs)})
		return
	}


	var post models.Post
	if err := h.db.Model(&models.Post{}).Where("id = ?", updatePost.ID).First(&post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "post doesnt exist"})
			return
		}

		log.Println(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	updatePost.UpdatedAt = time.Now()

	if post.Title != updatePost.Title {
		updatePost.Slug = utils.ConstructSlug(updatePost.Title)
	}
	
	if err := h.db.Model(&models.Post{}).Where("user_id = ?", c.GetInt64(keys.UserID)).Omit("id", "created_at", "user_id").Updates(&updatePost).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *PostHandler) Publish(c *gin.Context) {
	postId := c.Params.ByName("id")

	if postId == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var isPublished bool
	if err := c.ShouldBindJSON(isPublished); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	
	if err := h.db.Model(&models.Post{}).Where("user_id = ? AND id = ?", c.GetInt64(keys.UserID), postId).Update("is_published", !isPublished).Error; err != nil {
		log.Println(err.Error())
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
	
	if err := h.db.Where("user_id = ?", c.GetInt64(keys.UserID)).Delete(&models.Post{}, postId).Error; err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	
	c.Status(http.StatusNoContent)
}