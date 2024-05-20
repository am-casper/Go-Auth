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
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	NewsPref string `json:"newsPref"`
	MoviePref string `json:"moviePref"`
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

func createTask(user *User) error {
	_, err := usersCollection.InsertOne(ctx, user)
	return err
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

	if err := createTask(&newUser); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()+" Please Try Again!"})
	}

    c.IndentedJSON(http.StatusCreated, newUser)
}

func main() {
	var port = os.Getenv("PORT")
	r := gin.Default()
	r.GET("/users", getUsers)
	r.POST("/users", registerUser)
	r.Run(":"+port)
	fmt.Println("Listening to http://localhost:"+port+"/")
}
