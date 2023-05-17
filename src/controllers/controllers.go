package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"moduls/database"
	"moduls/pkg/models"
	generate "moduls/tokens"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var UserCollection *mongo.Collection = database.UserData(database.Client, "Users")
var ProductCollection *mongo.Collection = database.ProductData(database.Client, "Products")
var CommentCollection *mongo.Collection = database.CommentData(database.Client, "Comments")
var RatingCollection *mongo.Collection = database.RatingData(database.Client, "Rating")
var Validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userpassword string, givenpassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(givenpassword), []byte(userpassword))
	valid := true
	msg := ""
	if err != nil {
		msg = "Login Or Passowrd is Incorerct"
		valid = false
	}
	return valid, msg
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validationErr := Validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr})
			return
		}

		count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
		}
		count, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Phone is already in use"})
			return
		}
		password := HashPassword(*user.Password)
		user.Password = &password

		user.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_ID = user.ID.Hex()
		token, refreshtoken, _ := generate.TokenGenerator(*user.Email, *user.First_Name, *user.Last_Name, user.User_ID)
		user.Token = &token
		user.Refresh_Token = &refreshtoken
		user.UserCart = make([]models.ProductUser, 0)
		user.Address_Details = make([]models.Address, 0)
		user.Order_Status = make([]models.Order, 0)
		_, inserterr := UserCollection.InsertOne(ctx, user)
		if inserterr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "not created"})
			return
		}
		defer cancel()
		c.JSON(http.StatusCreated, "Successfully Signed Up!!")
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User
		var founduser models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		err := UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&founduser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login or password incorrect"})
			return
		}
		PasswordIsValid, msg := VerifyPassword(*user.Password, *founduser.Password)
		defer cancel()
		if !PasswordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			fmt.Println(msg)
			return
		}
		token, refreshToken, _ := generate.TokenGenerator(*founduser.Email, *founduser.First_Name, *founduser.Last_Name, founduser.User_ID)
		defer cancel()
		generate.UpdateAllTokens(token, refreshToken, founduser.User_ID)
		c.JSON(http.StatusFound, founduser)

	}
}

func ProductViewerAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var products models.Product
		defer cancel()
		if err := c.BindJSON(&products); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		products.Product_ID = primitive.NewObjectID()
		_, anyerr := ProductCollection.InsertOne(ctx, products)
		if anyerr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Not Created"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, "Successfully added our Product Admin!!")
	}
}

func SearchProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var productlist []models.Product
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		cursor, err := ProductCollection.Find(ctx, bson.D{{}})
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "Someting Went Wrong Please Try After Some Time")
			return
		}
		err = cursor.All(ctx, &productlist)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		defer cursor.Close(ctx)
		if err := cursor.Err(); err != nil {
			log.Println(err)
			c.IndentedJSON(400, "invalid")
			return
		}
		defer cancel()
		c.IndentedJSON(200, productlist)

	}
}

func SearchProductByQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		var searchproducts []models.Product
		queryParam := c.Query("name")
		if queryParam == "" {
			log.Println("query is empty")
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "Invalid Search Index"})
			c.Abort()
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		searchquerydb, err := ProductCollection.Find(ctx, bson.M{"product_name": bson.M{"$regex": queryParam}})
		if err != nil {
			c.IndentedJSON(404, "something went wrong in fetching the dbquery")
			return
		}
		err = searchquerydb.All(ctx, &searchproducts)
		if err != nil {
			log.Println(err)
			c.IndentedJSON(400, "invalid")
			return
		}
		defer searchquerydb.Close(ctx)
		if err := searchquerydb.Err(); err != nil {
			log.Println(err)
			c.IndentedJSON(400, "invalid request")
			return
		}
		defer cancel()
		c.IndentedJSON(200, searchproducts)
	}
}

func FilterProducts() gin.HandlerFunc {
	return func(c *gin.Context) {

		minPriceStr := c.Query("min_price")
		maxPriceStr := c.Query("max_price")

		minPrice, err := strconv.ParseFloat(minPriceStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid min_price parameter"})
			return
		}

		maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid max_price parameter"})
			return
		}

		filter := bson.M{
			"price": bson.M{
				"$gte": minPrice,
				"$lte": maxPrice,
			},
		}

		cursor, err := ProductCollection.Find(context.Background(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products from MongoDB"})
			return
		}
		defer cursor.Close(context.Background())

		var filteredProducts []models.Product
		for cursor.Next(context.Background()) {
			var product models.Product
			if err := cursor.Decode(&product); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode product"})
				return
			}
			filteredProducts = append(filteredProducts, product)
		}

		c.JSON(http.StatusOK, filteredProducts)
	}
}

func CommentProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var comment models.Comment
		if err := c.ShouldBindJSON(&comment); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		comment.Comment_ID = primitive.NewObjectID()
		result, err := CommentCollection.InsertOne(ctx, comment)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the comment"})
			return
		}

		commentID := result.InsertedID.(primitive.ObjectID).Hex()

		c.JSON(http.StatusOK, gin.H{"message": "Comment saved successfully", "comment_id": commentID})

	}
}

// RATING

func RateProduct() gin.HandlerFunc {
	return func(c *gin.Context) {

		var rating models.Rating
		if err := c.ShouldBindJSON(&rating); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if rating.Rating < 1 || rating.Rating > 5 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rating value. Rating must be between 1 and 5."})
			return
		}

		existingRating := getExistingRating(rating.UserID, rating.ProductID)
		if existingRating != nil {

			existingRating.Rating = rating.Rating
			existingRating.LastRated = time.Now()
			updateRating(existingRating)
			UpdateProductRating()
			c.JSON(http.StatusOK, gin.H{"message": "Rating updated successfully"})
			return
		}

		newRating := &models.Rating{
			UserID:    rating.UserID,
			ProductID: rating.ProductID,
			Rating:    rating.Rating,
			LastRated: time.Now(),
		}
		saveRating(newRating)
		UpdateProductRating()

		c.JSON(http.StatusOK, gin.H{"message": "Rating saved successfully"})
	}
}

func getExistingRating(userID string, productID primitive.ObjectID) *models.Rating {

	fmt.Println("getExistingRating")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var founduser models.User
	err := UserCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&founduser)
	fmt.Println(founduser)
	filter := bson.M{"user_id": userID, "productid": productID}
	var existingRating models.Rating
	err = RatingCollection.FindOne(ctx, filter).Decode(&existingRating)

	if err != nil {

		// No existing rating found
		return nil
	}

	return &existingRating
}

func saveRating(rating *models.Rating) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := RatingCollection.InsertOne(ctx, rating)
	if err != nil {
		fmt.Println("Failed to save the rating")
	}
}

func updateRating(rating *models.Rating) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"user_id": rating.UserID, "productid": rating.ProductID}
	update := bson.M{"$set": bson.M{"rating": rating.Rating, "last_rated": rating.LastRated}}
	_, err := RatingCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		fmt.Println("Failed to update the rating")
	}
}

func UpdateProductRating() {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := bson.A{
		bson.M{
			"$group": bson.M{
				"_id":           "$productid",
				"averageRating": bson.M{"$avg": "$rating"},
			},
		},
	}

	cursor, err := RatingCollection.Aggregate(ctx, pipeline)
	if err != nil {
		fmt.Println("Failed to aggregate ratings")
		return
	}

	for cursor.Next(ctx) {
		var result struct {
			ProductID     primitive.ObjectID `bson:"_id"`
			AverageRating float64            `bson:"averageRating"`
		}
		if err := cursor.Decode(&result); err != nil {
			fmt.Println("Failed to decode aggregate result")
			return
		}

		filter := bson.M{"_id": result.ProductID}
		update := bson.M{"$set": bson.M{"rating": result.AverageRating}}

		_, err := ProductCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			fmt.Println("Failed to update product rating")
		}
	}

	if err := cursor.Err(); err != nil {
		fmt.Println("Error occurred during cursor iteration")
	}

	fmt.Println("Product ratings updated successfully")
}
