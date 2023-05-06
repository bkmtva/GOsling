package main

import (
	"log"
	"moduls/controllers"
	"moduls/database"
	"moduls/middleware"
	"moduls/routes"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("Port")
	if port == "" {
		port = "8000"
	}
	app := controllers.NewApplication(database.ProductData(database.Client, "Products"), database.UserData(database.UserData(database.Client, "Users")))
	router := gin.New()
	router.Use(gin.Logger())
	routes.UserRoutes(router)
	router.Use(middleware.Authentication())
	router.GET("/addtocart", app.AddToCart())
	router.GET("/removeitem", app.RemoveItem())
	router.GET("/cartcheckout", app.BuyFromCart())
	router.GET("/instantbuy", app.InstantBuy())

	log.Fatal(router.Run(":" + port))
}
