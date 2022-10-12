package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/noctispine/blog/cmd/constants/roles"
	dbPackage "github.com/noctispine/blog/cmd/db"
	"github.com/noctispine/blog/cmd/handlers"
	"github.com/noctispine/blog/pkg/middlewares"
	"gorm.io/gorm"
)

var db *gorm.DB
var authHandler *handlers.AuthHandler
var postHandler *handlers.PostHandler
var ctx context.Context


func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db = dbPackage.GetDatabase()
	authHandler = handlers.NewAuthHandler(ctx, db)
	postHandler = handlers.NewPostHandler(ctx, db)
}



func main() {
	r := gin.Default()
	user := r.Group("/user")
	{
		user.POST("/sign-in", authHandler.SignInHandler)
		user.POST("/register", authHandler.Register)
		user.POST("/refresh/:id", authHandler.RefreshHandler)
	}

	posts := r.Group("/posts")
	{
		posts.GET("/all", postHandler.GetAll)
		posts.GET("/", middlewares.Pagination(), postHandler.GetPage)
	}

	

	blogger := r.Group("/", middlewares.ValidateToken(), middlewares.Authorization(roles.BLOGGER_PERMS))
	{
		bloggerPost := blogger.Group("/posts")
		{
			bloggerPost.POST("", postHandler.Create)
			bloggerPost.PATCH("", postHandler.Update)
			bloggerPost.DELETE(":id", postHandler.Delete)
		}
	}

	// admin := r.Group("/", middlewares.ValidateToken(), middlewares.Authorization(constants.ADMIN_PERMS) )
	// {
	// 	_ = admin.Group("posts")
	// }
	
	if os.Getenv("APP_ENV") == "PROD" {
		log.Fatal(r.Run(":" + os.Getenv("PROD_PORT")))
	} else {
		log.Fatal(r.Run(":" + os.Getenv("DEV_PORT")))
	}
}

