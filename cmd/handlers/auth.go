package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgerrcode"
	"github.com/noctispine/blog/cmd/models"
	"github.com/noctispine/blog/pkg/responses"
	"github.com/noctispine/blog/pkg/utils"
	"github.com/noctispine/blog/pkg/wrappers"
	log "github.com/sirupsen/logrus"
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
	if err != nil {
		return string(bytes),  fmt.Errorf("error while hashing password: %w", err)
	}
	return string(bytes), nil
}

func checkPasswordHash(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (h *AuthHandler) SignInHandler(c *gin.Context) {
	var user models.LoginUser
	var dbUser models.UserAccount

	if err := c.ShouldBindJSON(&user); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	
	if err := validate.Struct(user); err != nil {
		responses.AbortWithStatusJSONValidationErrors(c, http.StatusBadRequest, err)
	}

	if err := h.db.Where("email = ?", user.Email).First(&dbUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			responses.AbortWithStatusJSONError(c, http.StatusUnauthorized, wrappers.NewErrNotFound("user"))
			return
		}
		
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Error(err.Error())
		return
	}

	if !checkPasswordHash(user.Password, dbUser.PasswordHash) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Wrong Credentials"})
		return
	}

	expireInMinutes, err := strconv.Atoi(os.Getenv("JWT_EXPIRE_MINUTES"))
	if err != nil {
		log.Error(fmt.Errorf("jwt conversion to int: %w", err))
		c.AbortWithStatus(http.StatusInternalServerError)
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
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Error(fmt.Errorf("jwt signed string: %w", err))
		return
	}

	dbUser.LastLoginAt = time.Now()

	if err := h.db.Save(&dbUser).Error; err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Error(err.Error())
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
			responses.AbortWithStatusJSONError(c, http.StatusUnauthorized, fmt.Errorf("token is expired"))
			return
		} 

		c.AbortWithStatus(http.StatusUnauthorized)
		log.Error(fmt.Errorf("error while parsing claims: %w", err))
		return
	}

	if tkn == nil || !tkn.Valid {
		responses.AbortWithStatusJSONError(c, http.StatusUnauthorized, fmt.Errorf("invalid token"))
		return
	}

	if  time.Until(claims.ExpiresAt.Time).Minutes()  > 5 {
		responses.AbortWithStatusJSONError(c, http.StatusBadRequest, fmt.Errorf("token is not expired yet"))
		return
	}

	expirationTime := time.Now().Add(10 * time.Minute)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var user models.SignUpUser

	if err := c.ShouldBindJSON(&user); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		log.Error(err)
		return
	}


	if err := validate.Struct(user); err != nil {
		responses.AbortWithStatusJSONValidationErrors(c, http.StatusBadRequest, err)
		return
	}

	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Error(err.Error())
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
		if utils.CheckPostgreError(err, pgerrcode.UniqueViolation) {
			responses.AbortWithStatusJSONError(c, http.StatusBadRequest, wrappers.NewErrAlreadyExists("email"))
			return
		}
		
		c.AbortWithStatus(http.StatusBadRequest)
		log.Error(fmt.Errorf("while registering: %w", err))
		return
	}
	c.Status(http.StatusCreated)
}
