package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"coordinator/app/api/app"
	"coordinator/app/api/provider"
	"coordinator/app/client"
	"coordinator/app/ws"
	"coordinator/settings"

	dotenv "github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var port = flag.Int("port", 8080, "port")

func initEnv() {
    err := dotenv.Load(".env.development")
    if err != nil {
        log.Fatalln("Failed to load env from .env.development")
    }
}

func initDB() *gorm.DB {
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("PG_HOST"),
		os.Getenv("PG_USERNAME"),
		os.Getenv("PG_PASSWORD"),
		os.Getenv("PG_DB"),
		os.Getenv("PG_PORT"),
		os.Getenv("PG_SSL_MODE"))
    
    schemaName := "public"


    db, err := gorm.Open(postgres.New(postgres.Config{
        DSN:            dsn,
        PreferSimpleProtocol: true,
    }), &gorm.Config{
        NamingStrategy: schema.NamingStrategy{
            TablePrefix: fmt.Sprintf("%s.", schemaName),
        },
    })
    
    if err != nil {
        log.Fatalf("Failed to connect to postgres: %s", err)
    }
    log.Println("connected to db");
    return db;
}
func initRedis() *redis.Client {
    client := redis.NewClient(&redis.Options{
        Addr: os.Getenv("REDIS_URL"),
        Password: "",
    })
    fmt.Println("Connected to redis")
    return client;
}
var db *gorm.DB
var redisClient *redis.Client
func main() {
    initEnv()
    db = initDB()
    redisClient = initRedis()
    flag.Parse()
	hub := client.NewHub()

	mux := http.NewServeMux()
	
    mux.HandleFunc("/apps", func(w http.ResponseWriter, r *http.Request){
        app.GetAppList(db, w, r)
    })

	mux.HandleFunc("/providers", func(w http.ResponseWriter, r *http.Request) {
		provider.GetProviderList(context.Background(), redisClient, hub, w, r)
	})
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})

	c := cors.New(cors.Options{
		AllowedOrigins: settings.AllowedOrigins,
	})
	handler := c.Handler(mux)

	log.Println("Start listening on port", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), handler))
}
