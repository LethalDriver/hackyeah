package main

import (
	"context"
	"log"
	"payments-service/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Set up MongoDB client options
	clientOptions := options.Client().ApplyURI("mongodb+srv://fastapi:123fastapi@hackyeah-db.3xvq7.mongodb.net/?retryWrites=true&w=majority&appName=hackyeah-db")
	dbName := "hackyeahdb"

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	benefitRepo, err := repository.NewBenefitRepository(client.Database(dbName).Collection("benefits"), client.Database(dbName).Collection("owned_benefits"))
	if err != nil {
		log.Fatal(err)
	}
	walletRepo, err := repository.NewWalletRepository(client.Database(dbName).Collection("wallets"))
	if err != nil {
		log.Fatal(err)
	}

	// Create a new Gin router
	r := gin.Default()

	// Define your routes here
	r.GET("/benefits", func(c *gin.Context) {
		getAllBenefits(c, benefitRepo)
	})
	r.POST("/benefits", func(c *gin.Context) {
		addBenefit(c, benefitRepo)
	})

	r.GET("/benefits/:id", func(c *gin.Context) {
		getBenefit(c, benefitRepo)
	})

	r.POST("/benefits/:benefit_id/buy", func(c *gin.Context) {
		buyBenefit(c, benefitRepo, walletRepo)
	})

	r.GET("/wallets", func(c *gin.Context) {
		getAllWallets(c, walletRepo)
	})

	log.Println("Server running on port 8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}

}
