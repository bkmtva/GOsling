package token

import (
	"log"
	"moduls/database"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/mongo"
)

var UserData *mongo.Collection = database.userData(database.Client, "Users")
var SECRET_KEY = os.Getenv("SECRET_KEY")

type SignedDetails struct {
	Email      string
	Firts_Name string
	Last_Name  string
	Uid        string
	jwt.StandardClaims
}

func TokenGenerator(email string, firstname string, lastname string, uid string) (signedtoken string, signedrefreashtoken string, err error) {

	claims := &SignedDetails{
		Email:      email,
		Firts_Name: firstname,
		Last_Name:  lastname,
		Uid:        uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	refreshclaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		return "", "", err
	}

	refreshtoken, err := jwt.NewWithClaims(jwt.SigningMethodHS384, refreshclaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshtoken, err
}

func ValidateToken() {

}

func UpdateAllTokens() {

}
