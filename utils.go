package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/mongo"
)

func filterUsers(filter interface{}) ([]*User, error) {
	// A slice of users for storing the decoded documents
	var users []*User

	cur, err := usersCollection.Find(ctx, filter)
	if err != nil {
		return users, err
	}

	for cur.Next(ctx) {
		var u User
		err := cur.Decode(&u)
		if err != nil {
			return users, err
		}

		users = append(users, &u)
	}

	if err := cur.Err(); err != nil {
		return users, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)
	if len(users) == 0 {
		return users, mongo.ErrNoDocuments
	}

	return users, nil
}

func createUser(user *User) error {
	_, err := usersCollection.InsertOne(ctx, user)
	return err
}

func generateAccessToken(c *gin.Context, username string) (newAccessToken string) {
	// Generate JWT access token
	newAccessTokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(time.Hour * 2).Unix(), // Expires in 2 hours
	})
	newAccessToken, err := newAccessTokenClaims.SignedString([]byte(accessSecretKey))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}
	return newAccessToken
}

func generateRefreshToken(c *gin.Context, username string) (newRefreshToken string) {
	// Generate JWT refresh token
	newRefreshTokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(time.Hour * 24).Unix(), // Expires in 24 hours
	})
	newRefreshToken, err := newRefreshTokenClaims.SignedString([]byte(accessSecretKey))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return ""
	}
	return newRefreshToken
}

func contains(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
