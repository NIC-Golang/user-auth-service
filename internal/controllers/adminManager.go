package controllers

import (
	"context"
	"go/auth-service/internal/helpers"
	"go/auth-service/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func PromoteAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		authHeader := c.Request.Header.Get("Authorization")

		err := helpers.CheckAdmin(authHeader)
		if err != nil {
			c.JSON(403, gin.H{"error": "Access denied: only admins can promote users"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
		defer cancel()

		var user models.User
		err = userCollection.FindOne(ctx, bson.M{"user_id": id}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(404, gin.H{"error": "User not found"})
			} else if err == context.DeadlineExceeded {
				c.JSON(504, gin.H{"error": "Database timeout"})
			} else {
				c.JSON(500, gin.H{"error": "Database error"})
			}
			return
		}

		updateUser, err := helpers.UpdateMongo(id, "ADMIN")
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"message": "User promoted to admin",
			"user_id": updateUser.ID,
			"email":   updateUser.Email,
			"role":    updateUser.Type,
		})
	}
}

func DeleteAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		authHeader := c.Request.Header.Get("Authorization")

		err := helpers.CheckAdmin(authHeader)
		if err != nil {
			c.JSON(403, gin.H{"error": "Access denied: only admins can demote users"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
		defer cancel()

		var user models.User
		err = userCollection.FindOne(ctx, bson.M{"user_id": id}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(404, gin.H{"error": "User not found"})
			} else if err == context.DeadlineExceeded {
				c.JSON(504, gin.H{"error": "Database timeout"})
			} else {
				c.JSON(500, gin.H{"error": "Database error"})
			}
			return
		}

		updateUser, err := helpers.UpdateMongo(id, "USER")
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"message": "Admin demoted to user",
			"user_id": updateUser.ID,
			"email":   updateUser.Email,
			"role":    updateUser.Type,
		})
	}
}
