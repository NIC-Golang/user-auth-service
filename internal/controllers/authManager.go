package controllers

import (
	"context"
	"fmt"
	"go/auth-service/internal/config"
	"go/auth-service/internal/helpers"
	"go/auth-service/internal/models"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = config.GetCollection(config.DB, "users")

const timeoutDuration = 10 * time.Second

var validate = validator.New()

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
		defer cancel()
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if user.Type == nil || *user.Type == "" {
			defaultType := "USER"
			user.Type = &defaultType
		}
		validationErr := validate.Struct(&user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		userType := "USER"
		localzone, _ := time.LoadLocation("Asia/Almaty")
		userId := primitive.NewObjectID().Hex()
		token, refreshToken, err := helpers.CreateToken(*user.Email, *user.Name, *user.Type, userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token: " + err.Error()})
			return
		}
		hashedPass := helpers.HashPassword(*user.Password)
		newUser := models.User{
			ID:           primitive.NewObjectID(),
			User_id:      userId,
			Name:         user.Name,
			Email:        user.Email,
			Phone:        user.Phone,
			Password:     &hashedPass,
			Type:         &userType,
			Token:        token,
			RefreshToken: refreshToken,
			Created_at:   time.Now().In(localzone),
			Updated_at:   time.Now().In(localzone),
		}

		resultInsertionNumber, err := userCollection.InsertOne(ctx, newUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
			return
		}

		// TODO: reassign `http://notifier-service:8082` part in config/conf.yml file and use
		resp, err := http.Post("http://notifier-service:8082/auth/signup", "application/json", strings.NewReader(fmt.Sprintf(`{"name": "%s", "email": "%s"}`, *user.Name, *user.Email)))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request to notifier-service"})
			return
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Notifier-service returned an error",
				"details": string(body),
			})
			return
		}
		c.JSON(http.StatusCreated, resultInsertionNumber)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		var foundUser models.User
		ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
		defer cancel()

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "There was an error with scanning data..."})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		passwordIsValid, msg := helpers.VerifyingOfPassword(*user.Password, *foundUser.Password)
		if !passwordIsValid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials", "message": msg})
			return
		}

		token, refreshToken, err := helpers.CreateToken(*foundUser.Email, *foundUser.Name, *foundUser.Type, foundUser.User_id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating JWT"})
			return
		}

		if err := helpers.UpdateTokens(token, refreshToken, foundUser.User_id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tokens"})
			return
		}

		resp, err := http.Post("http://notifier-service:8082/auth/login", "application/json", strings.NewReader(fmt.Sprintf(`{"name": "%s", "email": "%s"}`, *foundUser.Name, *foundUser.Email)))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request to notifier-service"})
			return
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong with notifier-service", "details": string(body)})
			return
		}

		c.JSON(http.StatusOK, foundUser)
	}
}

func AdminCreatingWithContext(ctx context.Context, name, email, password, phone *string) error {
	typeAdmin := "ADMIN"
	hashedPass := helpers.HashPassword(*password)
	userId := primitive.NewObjectID().Hex()
	token, refreshToken, err := helpers.CreateToken(*email, *name, "ADMIN", userId)
	if err != nil {
		return err
	}
	localzone, _ := time.LoadLocation("Asia/Almaty")
	admin := &models.User{
		ID:           primitive.NewObjectID(),
		User_id:      userId,
		Name:         name,
		Email:        email,
		Password:     &hashedPass,
		Phone:        phone,
		Token:        token,
		RefreshToken: refreshToken,
		Type:         &typeAdmin,
		Created_at:   time.Now().In(localzone),
		Updated_at:   time.Now().In(localzone),
	}
	_, err = userCollection.InsertOne(ctx, admin)
	if err != nil {
		return err
	}
	return nil
}
