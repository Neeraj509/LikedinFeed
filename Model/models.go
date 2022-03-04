package model

import (
	"github.com/dgrijalva/jwt-go"
)

type User struct {
	FirstName string `json:"firstname" bson:"firstname"`
	LastName  string `json:"lastname" bson:"lastname"`
	Email     string `json:"email" bson:"email"`
	Password  string `json:"password" bson:"password"`
}

type Claims struct {
	UserName string `json:"username"`
	jwt.StandardClaims
}

type Feeds struct {
	Post string `json:"post" bson:"post"`
	Gid  int    `json:"gid" bson:"gid"`
}

type GroupId struct {
	Gid int `json:"gid" bson:"gid"`
}

type Email struct {
	Email string `json:"email" bson:"email"`
}

type Group struct {
	GrpName string   `json:"grpname" bson:"grpname"`
	Gid     int      `json:"gid" bson:"gid"`
	Members []string `json:"members" bson:"members"`
}

type FetchGroup struct {
	Members []string `json:"members" bson:"members"`
}
