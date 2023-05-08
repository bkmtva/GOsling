package database

import (
	"context"
	"errors"
	"log"
	"moduls/pkg/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrCantFindProduct    = errors.New("can't find the product")
	ErrCantDecodeProducts = errors.New("can't decode product")
	ErrUserIdIsNotVAlid   = errors.New("user is not valid")
	ErrCantUpdateUser     = errors.New("can't add this product to the cart")
	ErrCantRemoveItemCart = errors.New("can't remove this item from cart")
	ErrCantGetItem        = errors.New("can't get item")
	ErrCantBuyCartItem    = errors.New("can't buy cart item")
)

func AddProductToCart(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {

	searchfromdb, err := prodCollection.Find(ctx, bson.M{"id": productID})

	if err != nil {
		log.Println(err)
		return ErrCantFindProduct
	}

	var productCart []models.ProductUser
	err = searchfromdb.All(ctx, &productcart)
	if err != nil {
		log.Println(err)
		return ErrCantDecodeProducts
	}

	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "usercart", Value: bson.D{{Key: "$each", Value: productCart}}}}}}

	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return ErrCantUpdateUser
	}
	return nil
}

func RemoveCartItem() {

}

func BuyItemFromCart() {

}

func IstantBuyer() {

}
