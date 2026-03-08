package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// khởi tạo kết nối tới Redis Cache
func InitRedis() *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		panic("Kết nối Redis thất bại: " + err.Error())
	}

	fmt.Println("✅ Đã kết nối Redis Cache thành công!")
	return redisClient
}