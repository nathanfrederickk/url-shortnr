package database

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/go-redis/redis/v8"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var (
    MongoClient *mongo.Client
    RedisClient *redis.Client
    Ctx         = context.Background()
)

// Connect initializes connections to both MongoDB and Redis
func Connect() error {
    if err := connectMongoDB(); err != nil {
        return err
    }
    if err := connectRedis(); err != nil {
        return err
    }
    return nil
}

// connectMongoDB handles the connection to the MongoDB database
func connectMongoDB() error {
    mongoURI := os.Getenv("MONGO_URI")
    if mongoURI == "" {
        mongoURI = "mongodb://localhost:27017"
    }

    clientOptions := options.Client().ApplyURI(mongoURI)
    client, err := mongo.Connect(Ctx, clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    err = client.Ping(Ctx, nil)
    if err != nil {
        log.Fatal(err)
    }

    MongoClient = client
    fmt.Println("Connected to MongoDB!")
    return nil
}

// connectRedis handles the connection to the Redis cache
func connectRedis() error {
    redisAddr := os.Getenv("REDIS_ADDR")
    if redisAddr == "" {
        redisAddr = "localhost:6379"
    }

    rdb := redis.NewClient(&redis.Options{
        Addr: redisAddr,
    })

    _, err := rdb.Ping(Ctx).Result()
    if err != nil {
        log.Fatal(err)
    }
    RedisClient = rdb
    fmt.Println("Connected to Redis!")
    return nil
}