package helpers

import (
	"context"
	"fmt"
	"go/auth-service/internal/models"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TokenTaking(authHeader string) (string, error) {
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", fmt.Errorf("your token is empty")
	}
	return token, nil
}

func CheckAdmin(authHeader string) error {
	token, err := TokenTaking(authHeader)
	if err != nil {
		return err
	}
	claims, msg := ValidateToken(token)
	if msg != "" {
		return fmt.Errorf("error: %v", msg)
	}
	if claims.UserType != "ADMIN" {
		return fmt.Errorf("access is allowed only to administrators")
	}
	return nil
}

func UpdateMongo(id, usertype string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	update := bson.M{"$set": bson.M{"type": usertype}}
	res, err := userCollection.UpdateOne(ctx, bson.M{"user_id": id}, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update user type: %v", err)
	}
	if res.MatchedCount == 0 {
		return nil, fmt.Errorf("user not found")
	}

	var updatedUser models.User
	err = userCollection.FindOne(ctx, bson.M{"user_id": id}).Decode(&updatedUser)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated user data")
	}

	token, refreshToken, err := CreateToken(*updatedUser.Email, *updatedUser.Name, usertype, id)
	if err != nil {
		return nil, err
	}

	err = UpdateTokens(token, refreshToken, id)
	if err != nil {
		return nil, err
	}

	return &updatedUser, nil
}
