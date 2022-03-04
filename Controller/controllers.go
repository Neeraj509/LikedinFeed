package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	M "example.com/m/Model"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

const connectionString = "mongodb://localhost:27017"
const dbName = "FeedSystem"
const colName = "User"

var SECRET_KEY = []byte("gosecretkey")

var client *mongo.Client
var collection *mongo.Collection

func init() {
	//client option
	clientOption := options.Client().ApplyURI(connectionString)

	//connect to mongodb
	client, _ = mongo.Connect(context.TODO(), clientOption)

	fmt.Println("MongoDB connection success")

	collection := client.Database(dbName).Collection(colName)

	//collection instance
	fmt.Println("Collection instance is ready", collection)
}
func getHash(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}

func GenerateJWT() (string, error) {
	var user M.User
	expirationTime := time.Now().Add(time.Minute * 5)
	claims := &M.Claims{
		UserName: user.FirstName + user.LastName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(SECRET_KEY)
	if err != nil {
		log.Println("Error in JWT token generation")
		return claims.UserName, err
	}
	return tokenString, nil
}

func UserSignup(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var user M.User
	var dbUser M.User
	json.NewDecoder(request.Body).Decode(&user)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := client.Database(dbName).Collection(colName)

	err := collection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&dbUser)

	if err != nil {
		user.Password = getHash([]byte(user.Password))
		collection := client.Database(dbName).Collection(colName)
		result, _ := collection.InsertOne(ctx, user)

		json.NewEncoder(response).Encode(result)

	} else {

		fmt.Println("already registered")

	}

}

func UserLogin(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var user M.User
	var dbUser M.User
	json.NewDecoder(request.Body).Decode(&user)
	collection := client.Database(dbName).Collection(colName)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := collection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&dbUser)

	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	userPass := []byte(user.Password)
	dbPass := []byte(dbUser.Password)

	passErr := bcrypt.CompareHashAndPassword(dbPass, userPass)

	if passErr != nil {
		log.Println(passErr)
		response.Write([]byte(`{"response":"Wrong Password!"}`))
		return
	}

	jwtToken, err := GenerateJWT()
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	response.Write([]byte(`{"token":"` + jwtToken + `"}`))
	fmt.Println("you have succesfully logged in")

}

func ViewDetails(response http.ResponseWriter, r *http.Request) {

	response.Header().Set("Content-Type", "application/json")

	var email M.Email
	json.NewDecoder(r.Body).Decode(&email)

	collection := client.Database(dbName).Collection(colName)

	opts := options.Find().SetProjection(bson.M{"password": 0})

	cur, err := collection.Find(context.TODO(), bson.M{"email": email.Email}, opts)

	if err != nil {
		log.Fatal(err)
	}
	var episodes []bson.M
	if err = cur.All(context.Background(), &episodes); err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(response).Encode(episodes)

}

func getAllGroups(email string) []bson.M {

	episodesCollection := client.Database(dbName).Collection("Groups")

	opts := options.Find().SetProjection(bson.M{"grpname": 0, "_id": 0, "members": 0, "post": 0})

	Cursor, err := episodesCollection.Find(context.Background(), bson.M{"members": email}, opts)
	if err != nil {
		panic(err)
	}
	var Grps []bson.M
	if err = Cursor.All(context.Background(), &Grps); err != nil {
		panic(err)
	}

	return Grps

}
func getAllPosts(email string) []bson.M {
	collection := client.Database(dbName).Collection("Groups")

	matchStage := primitive.D{{"$match", primitive.D{{"members", email}}}}

	project := primitive.D{{"$project", primitive.D{{"_id", 0}, {"gid", 0}}}}

	sorting := primitive.D{{"$sort", primitive.D{{"createdOn", -1}}}}

	lookupStage := primitive.D{{"$lookup", primitive.D{{"from", "Feeds"}, {"localField", "gid"}, {"foreignField", "gid"}, {"as", "result"}}}}

	unwindStage := primitive.D{{"$unwind", primitive.D{{"path", "$result"}, {"preserveNullAndEmptyArrays", false}}}}

	replaceRoot := primitive.D{{"$replaceRoot", primitive.D{{"newRoot", "$result"}}}}

	cur, err := collection.Aggregate(context.Background(), mongo.Pipeline{matchStage, lookupStage, unwindStage, replaceRoot, sorting, project})
	if err != nil {
		log.Fatal(err)
	}
	var episodes []bson.M
	if err = cur.All(context.Background(), &episodes); err != nil {
		log.Fatal(err)
	}

	return episodes

}
func insertPost(feed M.Feeds) {
	fmt.Println(feed.Post)
	collection := client.Database(dbName).Collection("Feeds")

	inserted, err := collection.InsertOne(context.Background(), feed)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted 1 post in db with id: ", inserted.InsertedID)
	updateOnePost(feed.Post)

}
func updateOneGroup(feed M.Feeds) {
	collection := client.Database(dbName).Collection("Groups")

	filter := bson.M{"gid": feed.Gid}

	update := bson.M{"$set": bson.M{"post": feed.Post}}

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("modified count: ", result.ModifiedCount)
}

func updateOnePost(feed string) {
	collection := client.Database(dbName).Collection("Feeds")

	filter := bson.M{"post": feed}
	update := bson.M{"$set": bson.M{"createdOn": time.Now()}}

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("modified count: ", result.ModifiedCount)
}

func CreatePost(response http.ResponseWriter, r *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	var feed M.Feeds
	json.NewDecoder(r.Body).Decode(&feed)
	insertPost(feed)
	updateOneGroup(feed)
	json.NewEncoder(response).Encode(feed)

}
func GetMyAllPosts(response http.ResponseWriter, r *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var email M.Email
	json.NewDecoder(r.Body).Decode(&email)
	allpost := getAllPosts(email.Email)
	json.NewEncoder(response).Encode(allpost)
}
func UpdatePost(response http.ResponseWriter, r *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	var email M.Email
	json.NewDecoder(r.Body).Decode(&email)
	fmt.Println(email)

	updateOnePost(email.Email)
	json.NewEncoder(response).Encode(email)
}

func CreateOneGroup(group M.Group) {
	collection2 := client.Database(dbName).Collection("Groups")
	inserted, err := collection2.InsertOne(context.Background(), group)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("created 1 group in db with id: ", inserted.InsertedID)
}

func CreateGroup(response http.ResponseWriter, r *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	var group M.Group
	json.NewDecoder(r.Body).Decode(&group)
	CreateOneGroup(group)
	json.NewEncoder(response).Encode(group)
}

func GetMyAllGroups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var email M.Email

	json.NewDecoder(r.Body).Decode(&email)

	allGroups := getAllGroups(email.Email)
	json.NewEncoder(w).Encode(allGroups)
}
