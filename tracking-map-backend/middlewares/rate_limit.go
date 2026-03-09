package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// limit gọi API tối đa 60 lần/1 phút
func RateLimiter(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := "rate_limit:" + clientIP

		ctx := context.Background()
		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}
		if count == 1 {
			redisClient.Expire(ctx, key, time.Minute)
		}

		// Nếu vượt quá giới hạn Chặn ngay lập tức!
		if count > 60 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Quá nhiều yêu cầu (Spam)! Vui lòng thử lại sau 1 phút.",
			})
			return
		}
		// cho phép đi tiếp
		c.Next()
	}
}
