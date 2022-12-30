package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/ronenniv/Go-Gin-Auth/handlers"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var recipesHandler *handlers.RecipesHandler
var authHandler *handlers.AuthHandler

func init() {
	ctx := context.Background()

	// mongodb
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal("Errorr: cannot conenct ot MongoDB", err)
	}
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal("Error: cannot ping MongoDB", err)
	}
	log.Printf("Connected to MongoDB at %s", os.Getenv("MONGO_URI"))

	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	usersCollection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("users")

	// redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})
	status := redisClient.Ping(ctx)
	if status.Err() != nil {
		log.Fatal("Error: ping Redis", status.Err())
	}
	log.Printf("redisClient at %s with status %v\n", os.Getenv("REDIS_ADDR"), status)

	recipesHandler = handlers.NewRecipesHandler(ctx, collection, redisClient)
	authHandler = handlers.NewAuthHAndler(usersCollection, ctx)
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	// router.Use(authHandler.CORSMiddleware())
	router.Use(cors.Default())
	{
		router.POST("/login", authHandler.SignInHandlerJWT) // JWT
		router.POST("/refresh", authHandler.RefreshHandler)
		router.POST("/adduser", authHandler.AddUser) // for testing only - to create users
	}

	authorized := router.Group("/v1")
	authorized.Use(authHandler.AuthMiddlewareJWT()) // JWT
	{
		authorized.POST("/recipes", recipesHandler.NewRecipeHandler)
		authorized.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)
		authorized.DELETE("/recipes/:id", recipesHandler.DelRecipeHandler)
		authorized.GET("/recipes", recipesHandler.ListRecipesHandler)
		authorized.GET("/recipes/search", recipesHandler.SearchRecipesHandler)
		authorized.GET("/recipes/:id", recipesHandler.GetRecipeHandler)
	}
	router.Run()
}
