package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/yven/backend/internal/config"
	"github.com/yven/backend/internal/db"
	"github.com/yven/backend/internal/routes"
)

func main() {
	cfg := config.Load()

	database := db.Connect(cfg.DatabaseURL)
	if err := db.AutoMigrate(database); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("invalid REDIS_URL: %v", err)
	}
	redisClient := redis.NewClient(redisOpts)
	_ = redisClient // wired for future use: rate limiting, session cache, job queues

	r := gin.Default()
	routes.Register(r, cfg, database)

	log.Printf("YVEN backend listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
