package helpers

import (
	"context"
	"go/auth-service/internal/config"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func CheckType(c *gin.Context, userId string) string {
	uid := c.GetString("uid")
	userType := c.GetString("user_type")
	if userType == "ADMIN" && uid != userId {
		return ""
	} else {
		return "Unauthorized access to the server"
	}
}

func HashPassword(password string) string {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(hashed)
}

func VerifyingOfPassword(userPassword, foundUserPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(foundUserPassword), []byte(userPassword))
	check := true
	msg := ""
	if err != nil {
		check = false
		msg = "Email or password is incorrect"
	}
	return check, msg
}

type User struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

var userCollect *mongo.Collection = config.GetCollection(config.DB, "users")

func TakeName() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user User
		err := c.ShouldBindJSON(&user)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}
		var foundUser User
		err = userCollect.FindOne(context.Background(), bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"name": foundUser.Name})
	}
}
