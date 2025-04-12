package api

import (
    "github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
    // Set maximum multipart form size for file uploads (default is 32MB)
    r.MaxMultipartMemory = 8 << 20 // 8 MB

    // Serve the frontend
    r.GET("/", func(c *gin.Context) {
        c.HTML(200, "index.html", nil)
    })

    // Image routes
    r.POST("/upload", UploadImage)
    r.GET("/search", SearchImages)
    r.GET("/image/:id", GetImage)
}