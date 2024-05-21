package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	NewsPref  string `json:"newsPref"`
	MoviePref string `json:"moviePref"`
}

var usersCollection *mongo.Collection
var ctx = context.TODO()
var accessSecretKey = os.Getenv("ACCESS_SECRET_KEY")
var refreshSecretKey = os.Getenv("REFRESH_SECRET_KEY")

func init() {
	godotenv.Load()
	var mongoDatabaseName = os.Getenv("MONGO_DATABASE")
	var mongoCollectionName = os.Getenv("MONGO_COLLECTION")
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URL"))
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	usersCollection = client.Database(mongoDatabaseName).Collection(mongoCollectionName)
}

func getUsers(c *gin.Context) {
	filter := bson.D{{}}
	users, err := filterUsers(filter)
	if err != nil {
		panic(err)
	}
	c.IndentedJSON(http.StatusOK, users)
}

func registerUser(c *gin.Context) {
	var newUser User

	if err := c.BindJSON(&newUser); err != nil {
		if err.Error() == "json: cannot unmarshal number into Go struct field User.newsPref of type string" {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "newsPref and moviePref must be a string"})
		} else {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	// checks for bad request
	if newUser.Username == "" || newUser.Password == "" || newUser.NewsPref == "" || newUser.MoviePref == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "username, password, newsPref and moviePref are required"})
		return
	}

	// checks for duplicate username
	filter := bson.D{{Key: "username", Value: newUser.Username}}
	_, err := filterUsers(filter)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "username already exists"})
		return
	}

	var pwd = []byte(newUser.Password)
	hashedPassword, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	newUser.Password = string(hashedPassword)

	if err := createUser(&newUser); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error() + " Please Try Again!"})
	}

	c.IndentedJSON(http.StatusCreated, newUser)
}

func loginUser(c *gin.Context) {
	var user User

	if err := c.BindJSON(&user); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := bson.D{{Key: "username", Value: user.Username}}
	users, err := filterUsers(filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(users[0].Password), []byte(user.Password))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password"})
		return
	}

	// Generate JWT access token
	accessToken := generateAccessToken(c, user.Username)

	// Generate JWT refresh token
	refreshToken := generateRefreshToken(c, user.Username)

	if accessToken=="" || refreshToken=="" {
		return 
	}

	c.SetCookie("access-token", accessToken, 3600, "/", "", true, true)
	c.SetCookie("refresh-token", refreshToken, 3600, "/", "", true, true)

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Login Successful"})
}

func getUserInfo(c *gin.Context) {

	accessTokenCookie, err := c.Cookie("access-token")
	if err != nil {
		c.String(http.StatusNotFound, "Cookie not found")
		return
	}
	
	// Parse JWT token with claims
	accessToken, err := jwt.ParseWithClaims(accessTokenCookie, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(accessSecretKey), nil
	})
	
	// Handle token parsing errors
	if err != nil {
		c.String(http.StatusUnauthorized, "Unauthorized access. Please refresh the token.")
		return
	}

	// Extract claims from token
	claims, ok := accessToken.Claims.(*jwt.MapClaims)
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Error extracting claims"})
		return
	}

	var username, okay = (*claims)["sub"]
	if !okay {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Error extracting username"})
		return
	}
	filter := bson.D{{Key: "username", Value: username}}
	users, err := filterUsers(filter)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, users[0])
}

func refreshTokenPair(c *gin.Context) {
	refreshTokenCookie, err := c.Cookie("refresh-token")
	if err != nil {
		c.String(http.StatusNotFound, "Cookie not found")
		return
	}
	// Parse JWT token with claims
	refreshToken, err := jwt.ParseWithClaims(refreshTokenCookie, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(refreshSecretKey), nil
	})

	// Handle token parsing errors
	if err != nil {
		c.String(http.StatusUnauthorized, "Unauthorized access. Please login again.")
		return
	}

	// Extract claims from token
	claims, ok := refreshToken.Claims.(*jwt.MapClaims)
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Error extracting claims"})
		return
	}

	var username, okay = (*claims)["sub"]
	if !okay {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Error extracting username"})
		return
	}

	// Generate JWT access token
	newAccessToken := generateAccessToken(c, username.(string))

	// Generate JWT refresh token
	newRefreshToken := generateRefreshToken(c, username.(string))

	if newAccessToken=="" || newRefreshToken=="" {
		return 
	}

	c.SetCookie("access-token", newAccessToken, 3600, "/", "", true, true)
	c.SetCookie("refresh-token", newRefreshToken, 3600, "/", "", true, true)

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Token Refreshed"})

}

func main() {
	var port = os.Getenv("PORT")
	r := gin.Default()
	r.GET("/users", getUsers)
	r.POST("/register", registerUser)
	r.POST("/login", loginUser)
	r.GET("/userInfo", getUserInfo)
	r.POST("/refresh", refreshTokenPair)
	r.Run(":" + port)
	fmt.Println("Listening to http://localhost:" + port + "/")
}
