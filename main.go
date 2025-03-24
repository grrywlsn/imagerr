package main

import (
    "log"
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    "github.com/grrywlsn/imagerr/src/api"
    "github.com/grrywlsn/imagerr/src/db"
    "github.com/grrywlsn/imagerr/src/storage"
    "github.com/grrywlsn/imagerr/src/search"
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