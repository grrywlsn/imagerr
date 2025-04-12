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

    // API routes group
    api := r.Group("/api")
    {
        // Image routes
        api.POST("/images", UploadImage)
        api.GET("/images/search", SearchImages)
        api.GET("/images/:id", GetImage)
        api.POST("/test-s3", TestS3Upload)
        
        // Optional: Add more routes as needed
        // api.DELETE("/images/:id", DeleteImage)
        // api.PUT("/images/:id", UpdateImage)
    }
}