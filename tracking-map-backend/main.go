package main

import (
	"fmt"

	"tracking-map-backend/controllers"
	"tracking-map-backend/database"
	"tracking-map-backend/middlewares"
	"tracking-map-backend/repositories"
	"tracking-map-backend/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Kết nối Database & Redis
	dbPool := database.InitPostgres()
	defer dbPool.Close()

	redisClient := database.InitRedis()
	defer redisClient.Close()
	fmt.Println("Hệ thống tracking map sẵn sàng!")

	// Khởi tạo các lớp Repository & Service
	repo := &repositories.LocationRepo{DB: dbPool, Redis: redisClient}
	service := &services.LocationService{Repo: repo}
	
	locController := &controllers.LocationController{Repo: repo, Service: service}
	
	tileController := &controllers.TileController{Redis: redisClient} 

	r := gin.Default()
	r.Use(cors.Default())

	// 1. Nạp file biên giới vào RAM để Crawler sử dụng
	err := services.LoadVnBoundary()
	if err != nil {
		panic("Không thể khởi động vì thiếu vn.json: " + err.Error())
	}
	
	// 2. Chạy ngầm con Robot cào bản đồ
	go services.StartVietnamCrawler()

	r.GET("/tiles/:z/:x/:y", tileController.GetTile)
	
	
	r.StaticFile("/api/boundary", "./vn.json") 

	// 4. CÁC API CỐT LÕI (Bảo vệ bằng Rate Limiter Redis)
	api := r.Group("/api/v1")
	api.Use(middlewares.RateLimiter(redisClient))
	{
		api.GET("/devices", locController.GetDeviceList)       
		api.POST("/locations", locController.PostLocation) 
		api.GET("/devices/:device_id/latest", locController.GetLatest)
		api.GET("/devices/:device_id/history", locController.GetHistory)
		api.GET("/route", locController.GetRoute)
	}

	// Khởi động server
	r.Run(":8080")
}