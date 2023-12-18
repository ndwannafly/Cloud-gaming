package session

import (
	"sync"
	"time"
	// "encoding/json"
	"log"
	// "net/http"
	"fmt"
	"os"
    "context"

	dotenv "github.com/joho/godotenv"
	// "golang.org/x/net/idna"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
    "github.com/redis/go-redis/v9"
)


type Hub struct {
	sessions map[string]*Session
	rwMutex  sync.RWMutex
}

var db *gorm.DB
var redisClient *redis.Client
func initEnv() {
    err := dotenv.Load(".env.development")
    if err != nil {
        log.Fatalln("Failed to load env from .env.development")
    }
    log.Println("imported env from .env.development")
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

func init() {
    initEnv()
    db = initDB()
    redisClient = initRedis()
    
}

type PlayerSession struct {
    ID          int
    PlayerID    string
    TimeStart   time.Time
    TimeEnd     *time.Time
    AppID       string
    OwnerID     string
}

func NewHub() *Hub {
	return &Hub{
		sessions: make(map[string]*Session),
		rwMutex:  sync.RWMutex{},
	}
}

func (h *Hub) AddSession(s *Session) {
	h.rwMutex.Lock()
	defer h.rwMutex.Unlock()

	h.sessions[s.playerID] = s
    newPlayerSession := &PlayerSession{
        PlayerID: s.playerID,
        TimeStart: s.timeStart,
        TimeEnd: nil,
        AppID: s.appID,
        OwnerID: s.ownerID,
    }
    db.Model(&PlayerSession{}).Create(newPlayerSession)
    err := redisClient.Incr(context.Background(), s.ownerID).Err()
    if err != nil {
        fmt.Println("Can not get number of session from Redis");
        return;
    }


}

func (h *Hub) RemoveSession(playerID string, ownerID string) {
	h.rwMutex.Lock()
	defer h.rwMutex.Unlock()

	if _, ok := h.sessions[playerID]; ok {
		delete(h.sessions, playerID)
	}
    lastSession := &PlayerSession{}
    db.Model(&PlayerSession{}).
        Where("player_id = ?", playerID).
        Where("time_end IS NULL").
        Where("owner_id = ?", ownerID).
        Order("id DESC").
        First(&lastSession)
    fmt.Println(lastSession)
    if lastSession.ID != 0 {
        now := time.Now()
        db.Model(&PlayerSession{}).
            Where("id = ?", lastSession.ID).
            UpdateColumn("time_end", &now)
    }
    err := redisClient.Decr(context.Background(), ownerID).Err()
    if err != nil {
        fmt.Println("Can not get number of session from Redis");
        return;
    }
}

func (h *Hub) GetSession(playerID string) *Session {
	h.rwMutex.RLock()
	defer h.rwMutex.RUnlock()

	if s, ok := h.sessions[playerID]; ok {
		return s
	}

	return nil
}
