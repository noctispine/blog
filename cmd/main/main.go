package main

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/noctispine/blog/cmd/constants/roles"
	dbPackage "github.com/noctispine/blog/cmd/db"
	"github.com/noctispine/blog/cmd/handlers"
	"github.com/noctispine/blog/pkg/middlewares"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var db *gorm.DB
var authHandler *handlers.AuthHandler
var postHandler *handlers.PostHandler
var categoryHandler *handlers.CategoryHandler
var tagHandler *handlers.TagHandler
var postCategoryHandler *handlers.PostCategoryHandler
var ctx context.Context


func init() {
	log.SetFormatter(&log.JSONFormatter{})
	
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db = dbPackage.GetDatabase()
	authHandler = handlers.NewAuthHandler(ctx, db)
	postHandler = handlers.NewPostHandler(ctx, db)
	categoryHandler = handlers.NewCategoryHandler(ctx, db)
	tagHandler = handlers.NewTagHandler(ctx, db)
	postCategoryHandler = handlers.NewPostCategoryHandler(ctx, db)

}



func main() {
	r := gin.Default()
	r.Use(middlewares.CORSMiddleware())
	user := r.Group("/user")
	{
		user.POST("/sign-in", authHandler.SignInHandler)
		user.POST("/register", authHandler.Register)
		user.POST("/refresh/:id", authHandler.RefreshHandler)
	}

	posts := r.Group("/posts")
	{
		posts.GET("/all", postHandler.GetAll)
		posts.GET(":id", postHandler.GetUserPosts)
		posts.GET("", middlewares.Pagination(), postHandler.GetPage)
		posts.GET(":id", middlewares.Pagination(), postHandler.GetPageByCategory)
	}

	categories := r.Group("/categories")
	{
		categories.GET("", categoryHandler.GetAll)
	}

	tags := r.Group("/tags")
	{
		tags.GET("", tagHandler.GetAll)
	}

	

	blogger := r.Group("/", middlewares.ValidateToken(), middlewares.Authorization(roles.BLOGGER_PERMS))
	{
		bloggerPost := blogger.Group("posts")
		{
			bloggerPost.POST("", postHandler.Create)
			bloggerPost.PATCH("", postHandler.Update)
			bloggerPost.DELETE(":id", postHandler.Delete)
			bloggerPost.PATCH(":id", postHandler.TogglePublish)
		}

		bloggerPostCategory := blogger.Group("post-category", middlewares.PostMatchesWithUser(db))
		{
			bloggerPostCategory.POST("", postCategoryHandler.Create)
			bloggerPostCategory.DELETE("", postCategoryHandler.Delete)
		}
	}

	admin := r.Group("/", middlewares.ValidateToken(), middlewares.Authorization(roles.ADMIN_PERMS) )
	{
		adminCategory := admin.Group("categories")
		{
			adminCategory.POST("", categoryHandler.Create)
			adminCategory.DELETE(":id", categoryHandler.Delete)
			adminCategory.PATCH("", categoryHandler.Update)
		}

		adminTag := admin.Group("tags")
		{
			adminTag.POST("", tagHandler.Create)
			adminTag.DELETE(":id", tagHandler.Delete)
			adminTag.PATCH("", tagHandler.Update)
		}
	}
	
	if os.Getenv("APP_ENV") == "PROD" {
		log.Fatal(r.Run(":" + os.Getenv("PROD_PORT")))
	} else {
		log.Fatal(r.Run(":" + os.Getenv("DEV_PORT")))
	}
}

