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


    repo := &repositories.LocationRepo{DB: dbPool, Redis: redisClient}
    service := &services.LocationService{Repo: repo}
    controller := &controllers.LocationController{Repo: repo, Service: service}

    r := gin.Default()
    r.Use(cors.Default())

    // bản đồ khu vực FPT/Ngũ Hành Sơn
    go services.DownloadMapArea(108.254083, 15.983176)

    // tránh limit nên để ngoài api
    r.Static("/tiles", "./tiles")


    api := r.Group("/api/v1")
    api.Use(middlewares.RateLimiter(redisClient)) 
    {
        api.GET("/devices", controller.GetDeviceList)
        api.POST("/locations", controller.PostLocation)
        api.GET("/devices/:device_id/latest", controller.GetLatest)
        api.GET("/devices/:device_id/history", controller.GetHistory)
        api.GET("/route", controller.GetRoute)
    }

    r.Run(":8080")
}