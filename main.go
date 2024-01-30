package main

import (
	"context"
	"log"
	"net/http"
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

	authorized := router.Group("")
	authorized.Use(AuthMiddleware())
	authorized.GET("/recipes", recipesHandler.ListRecipesHandler)
	authorized.PUT("/recipes/:id", recipesHandler.UpdateRecipesHandler)
	authorized.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
	authorized.POST("/recipes", recipesHandler.CreateRecipesHandler)
	authorized.GET("/recipes/search", recipesHandler.SearchRecipesHandler)

	if err = router.Run(); err != nil {
		log.Fatal(err)
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-API-KEY") != os.Getenv("X_API_KEY") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.AbortWithStatus(401)
			return
		}
		c.Next() 
	}
}
