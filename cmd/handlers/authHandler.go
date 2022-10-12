package handlers

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/noctispine/blog/cmd/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	ctx context.Context
	db *gorm.DB
}

type Claims struct {
	Email string `json:"email"`
	UserID int64 
	jwt.RegisteredClaims
	Role int
}

type JWTOutput struct {
	Token string `json:"token"`
	Expires time.Time `json:"expires"`
}

func NewAuthHandler(ctx context.Context, db *gorm.DB) *AuthHandler{
	return &AuthHandler{
		ctx,
		db,
	}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (h *AuthHandler) SignInHandler(c *gin.Context) {
	var user models.LoginUser
	var dbUser models.UserAccount

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})

		return
	}
	
	if err := validate.Struct(user); err != nil {
		errs := translateError(err, enTrans)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": stringfyJSONErrArr(errs)})
		return
	}

	if err := h.db.Where("email = ?", user.Email).First(&dbUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "user not found"})
			return
		}
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error()})
		return
	}

	if !checkPasswordHash(user.Password, dbUser.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Wrong Credentials"})
		return
	}

	expireInMinutes, err := strconv.Atoi(os.Getenv("JWT_EXPIRE_MINUTES"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error()})
		return
	}

	expirationTime := time.Now().Add(time.Duration(expireInMinutes) * time.Minute)
	claims := &Claims{
		Email: user.Email,
		UserID: dbUser.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
		Role: dbUser.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error()})
		return
	}

	dbUser.LastLoginAt = time.Now()

	if err := h.db.Save(&dbUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error()})	
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": models.UserAccount{
			Email: dbUser.Email,
			FirstName: dbUser.FirstName,
			LastName: dbUser.LastName,
			Role: dbUser.Role,
			IntroDesc: dbUser.IntroDesc,
			ProfileDesc: dbUser.ProfileDesc,
		},
		"token": tokenString})
}

func (h *AuthHandler) RefreshHandler(c *gin.Context) {
	tokenValue := c.GetHeader("Authorization")

	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenValue, claims, 
		func(token *jwt.Token) (interface{}, error){
			return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token is expired"})
			return
		} 

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error()})
		return
	}

	if tkn == nil || !tkn.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token"})
		return
	}

	if  time.Until(claims.ExpiresAt.Time).Minutes()  > 5 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Token is not expired yet"})
		return
	}

	expirationTime := time.Now().Add(10 * time.Minute)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var user models.SignUpUser

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}


	if err := validate.Struct(user); err != nil {
		errs := translateError(err, enTrans)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": stringfyJSONErrArr(errs)})
		return
	}

	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error()})
		return
	}

	newUser := models.UserAccount{
		FirstName: user.FirstName,
		LastName: user.LastName,
		Email: user.Email,
		PasswordHash: hashedPassword,
		IntroDesc: user.IntroDesc,
		ProfileDesc: user.ProfileDesc,
		RegisteredAt: time.Now(),
	}

	if err := h.db.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}
