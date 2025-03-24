package main

import (
    "log"
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    "imagerr/src/api"
    "imagerr/src/db"
    "imagerr/src/storage"
    "imagerr/src/search"
)

func main() {
    if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }

    // Initialize services
    db.InitDB()
    storage.InitS3()
    search.InitElasticsearch()

    // Setup router
    r := gin.Default()
    
    // Serve static files
    r.Static("/static", "./src/frontend/static")
    r.LoadHTMLGlob("src/frontend/*.html")
    
    // Setup routes
    api.SetupRoutes(r)

    r.Run(":8080")
}