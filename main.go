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
	// "go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/gin-gonic/gin"

	"golang.org/x/crypto/bcrypt"

	"crypto/rand"
	"encoding/hex"
	"time"
)

type User struct {
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	NewsPref    string    `json:"newsPref"`
	MoviePref   string    `json:"moviePref"`
	AccessToken string    `json:"accessToken"`
	ExpiryTime  time.Time `json:"expiryTime"`
}

var usersCollection *mongo.Collection
var ctx = context.TODO()

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

func filterTasks(filter interface{}) ([]*User, error) {
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

func randomHex() (string, error) {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func getUsers(c *gin.Context) {
	filter := bson.D{{}}
	users, err := filterTasks(filter)
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
	_, err := filterTasks(filter)
	log.Println(err)
	if err != nil {
		log.Println("yha hai")
		if err != mongo.ErrNoDocuments {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "username already exists"})
		return
	}

	// hash password
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

	// checks for bad request
	if user.Username == "" || user.Password == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	// checks for username
	filter := bson.D{{Key: "username", Value: user.Username}}
	users, err := filterTasks(filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "username not found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// compare password
	err = bcrypt.CompareHashAndPassword([]byte(users[0].Password), []byte(user.Password))
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}

	// generate access token
	accessToken, err := randomHex()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// update access token and expiry time
	filter = bson.D{{Key: "username", Value: user.Username}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "accesstoken", Value: accessToken}, {Key: "expirytime", Value: time.Now().Add(time.Hour * 2)}}}}
	_, err = usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"accessToken": accessToken})
}

func getUserInfo(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "access token is required"})
		return
	}

	// verifies access token
	accessToken = accessToken[7:] // authorization header starts with "Bearer "
	filter := bson.D{{Key: "accesstoken", Value: accessToken}}
	users, err := filterTasks(filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// check for expiry time
	var user = users[0]
	if time.Now().After(user.ExpiryTime) {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "access token has expired! Please login again!"})
		return
	}

	c.IndentedJSON(http.StatusOK, user)
}

func main() {
	var port = os.Getenv("PORT")
	r := gin.Default()
	r.GET("/users", getUsers)
	r.POST("/register", registerUser)
	r.POST("/login", loginUser)
	r.GET("/userInfo", getUserInfo)
	r.Run(":" + port)
	fmt.Println("Listening to http://localhost:" + port + "/")
}
