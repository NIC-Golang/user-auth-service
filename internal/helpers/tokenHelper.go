package helpers

import (
	"context"
	"fmt"
	"go/auth-service/internal/config"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	UserType string `json:"user_type"`
	Uid      string `json:"uid"`
	jwt.RegisteredClaims
}

var userCollection *mongo.Collection = config.GetCollection(config.DB, "users")
var key = os.Getenv("KEY")

func CreateToken(email, name, userType, userId string) (tokenWithClaims, refreshTokenWithClaims string, err error) {
	claims := SignedDetails{
		Email:    email,
		Name:     name,
		UserType: userType,
		Uid:      userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	refreshClaims := SignedDetails{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(168 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenWithClaims, err = token.SignedString([]byte(key))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenWithClaims, err = refreshToken.SignedString([]byte(key))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return
}

func UpdateTokens(token, refreshToken, userId string) error {
	var ctx, cancel = context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()
	localzone := time.FixedZone("UTC+5", 5*60*60)
	var updateObj primitive.D
	updateObj = append(updateObj, bson.E{Key: "token", Value: token})
	updateObj = append(updateObj, bson.E{Key: "refresh_token", Value: refreshToken})
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: time.Now().In(localzone).Format(time.RFC3339)})

	upsert := true
	filter := bson.M{"user_id": userId}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := userCollection.UpdateOne(ctx, filter, bson.D{{Key: "$set", Value: updateObj}}, &opt)
	if err != nil {
		return fmt.Errorf("failed to update tokens for user %s: %v", userId, err)
	}

	return nil
}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, fmt.Sprintf("error loading .env file: %v", err)
	}

	token, err := jwt.ParseWithClaims(signedToken, &SignedDetails{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})

	if err != nil {
		return nil, fmt.Sprintf("error parsing token: %v", err)
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		return nil, "invalid token"
	}

	if claims.ExpiresAt.Unix() < time.Now().UTC().Unix() {
		return nil, "token is expired"
	}

	return claims, ""
}

func ExtractClaimsFromToken(userToken string) (*SignedDetails, error) {
	token, err := jwt.ParseWithClaims(userToken, &SignedDetails{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	if claims.ExpiresAt.Unix() < time.Now().UTC().Unix() {
		return nil, fmt.Errorf("token is expired")
	}
	if claims.Email == "" || claims.UserType == "" {
		return nil, fmt.Errorf("missing essential claim data")
	}
	return claims, nil
}
