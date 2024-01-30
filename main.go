package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/marcelluseasley/recipes-api/handlers"
)

func main() {
	var recipesHandler *handlers.RecipesHandler
	var authHandler *handlers.AuthHandler

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

	collectionUsers := client.Database("demo").Collection("users")
	authHandler = handlers.NewAuthHandler(ctx, collectionUsers)

	router := gin.Default()

	s := &http.Server{
		Addr: ":8080",
		Handler: router,
		ReadTimeout: 2 *time.Second,
		WriteTimeout: 2*time.Second,
	}

	authorized := router.Group("")
	authorized.Use(handlers.AuthMiddleware())
	authorized.GET("/recipes", recipesHandler.ListRecipesHandler)
	authorized.PUT("/recipes/:id", recipesHandler.UpdateRecipesHandler)
	authorized.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
	authorized.POST("/recipes", recipesHandler.CreateRecipesHandler)
	authorized.GET("/recipes/search", recipesHandler.SearchRecipesHandler)
	router.POST("/signin", authHandler.SignInHandler)
	router.POST("/refresh", authHandler.RefreshHandler)

	//router.Run()
	log.Fatal(s.ListenAndServe())
}
