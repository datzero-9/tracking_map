package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	
	"tracking-map-backend/services" 
)

type TileController struct {
	Redis *redis.Client
}

func (ctrl *TileController) GetTile(c *gin.Context) {
	z, x, y := c.Param("z"), c.Param("x"), c.Param("y")
	ctx := context.Background()

	cacheKey := fmt.Sprintf("tile:%s:%s:%s", z, x, y)

	// 1. CHẠY RA TỦ LẠNH REDIS (RAM) TÌM TRƯỚC
	val, err := ctrl.Redis.Get(ctx, cacheKey).Bytes()
	if err == nil {
		c.Header("Cache-Control", "public, max-age=86400")
		c.Header("X-Cache", "HIT-REDIS")
		c.Data(200, "image/png", val)
		return
	}

	// 2. TỦ LẠNH KHÔNG CÓ -> GỌI SERVICE XUỐNG Ổ CỨNG LẤY
	imgData, err := services.GetTileFromDisk(z, x, y)
	if err != nil {
		c.JSON(404, gin.H{"error": "Tile not found"})
		return
	}

	// 3. LẤY ĐƯỢC RỒI THÌ BỎ VÀO TỦ LẠNH REDIS CHO LẦN SAU
	ctrl.Redis.Set(ctx, cacheKey, imgData, 24*time.Hour)

	c.Header("Cache-Control", "public, max-age=86400")
	c.Header("X-Cache", "MISS-DISK")
	c.Data(200, "image/png", imgData)
}