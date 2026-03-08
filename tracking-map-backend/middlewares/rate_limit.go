package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter giới hạn mỗi IP chỉ được gọi API tối đa 60 lần / 1 phút
func RateLimiter(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := "rate_limit:" + clientIP

		ctx := context.Background()

		// Tăng biến đếm số lần gọi của IP này lên 1
		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		// Nếu đây là lần gọi đầu tiên, đặt đồng hồ đếm ngược 1 phút (60 giây)
		if count == 1 {
			redisClient.Expire(ctx, key, time.Minute)
		}

		// Nếu vượt quá giới hạn 60 lần / phút -> Chặn ngay lập tức!
		if count > 60 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Quá nhiều yêu cầu (Spam)! Vui lòng thử lại sau 1 phút.",
			})
			return
		}

		// Cho phép đi tiếp vào Controller
		c.Next()
	}
}