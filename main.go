package main

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/marcelluseasley/recipes-api/handlers"
)

func main() {
	var recipesHandler *handlers.RecipesHandler

	ctx := context.Background()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal(err)
	}
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	status := redisClient.Ping(ctx)
	log.Println(status)

	collection := client.Database("demo").Collection("recipes")
	recipesHandler = handlers.NewRecipesHandler(ctx, collection, redisClient)

	router := gin.Default()

	router.POST("/recipes", recipesHandler.CreateRecipesHandler)
	router.GET("/recipes", recipesHandler.ListRecipesHandler)
	router.PUT("/recipes/:id", recipesHandler.UpdateRecipesHandler)
	router.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
	router.GET("/recipes/search", recipesHandler.SearchRecipesHandler)

	router.Run()
}
