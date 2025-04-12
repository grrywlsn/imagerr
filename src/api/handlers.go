package api

import (
    "log"
    "fmt"
    "net/http"
    "strings"
    "path/filepath"
    "github.com/gin-gonic/gin"
    "github.com/grrywlsn/imagerr/src/db"
    "github.com/grrywlsn/imagerr/src/storage"
    "github.com/grrywlsn/imagerr/src/search"
    "strconv"
    "os"
    "time"
)

func UploadImage(c *gin.Context) {
    // Get the file from form
    file, header, err := c.Request.FormFile("image")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
        return
    }
    defer file.Close()

    // Get other form data
    description := c.PostForm("description")
    tagsStr := c.PostForm("tags")
    tags := strings.Split(tagsStr, ",")
    for i := range tags {
        tags[i] = strings.TrimSpace(tags[i])
    }

    // Upload to S3
    filename := filepath.Base(header.Filename)
    storagePath, err := storage.UploadFile(file, filename)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
        return
    }

    // Save to database
    image, err := db.CreateImage(filename, description, storagePath, tags)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save to database"})
        return
    }

    // Index in Elasticsearch
    if err := search.IndexImage(image.ID, image.Description, image.Tags); err != nil {
        log.Printf("Warning: Failed to index image in Elasticsearch: %v", err)
        // Don't return error to client as the image is already saved
    }

    c.JSON(http.StatusOK, image)
}

func SearchImages(c *gin.Context) {
    query := c.Query("q")
    if query == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter is required"})
        return
    }

    // Search in Elasticsearch
    searchResults, err := search.SearchImages(query)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search images"})
        return
    }

    // Get full image details from database
    var images []db.Image
    for _, result := range searchResults {
        image, err := db.GetImageByID(result.ID)
        if err != nil {
            continue
        }
        if image != nil {
            image.URL = storage.GetFileURL(image.StoragePath)
            images = append(images, *image)
        }
    }

    c.JSON(http.StatusOK, images)
}

func GetImage(c *gin.Context) {
    id := c.Param("id")
    imageID, err := strconv.ParseInt(id, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
        return
    }

    image, err := db.GetImageByID(imageID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get image"})
        return
    }

    if image == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
        return
    }

    image.URL = storage.GetFileURL(image.StoragePath)
    c.JSON(http.StatusOK, image)
}

func TestS3Upload(c *gin.Context) {
    log.Printf("Starting S3 test upload with bucket: %s", os.Getenv("S3_BUCKET"))
    
    testContent := strings.NewReader("This is a test file to verify S3 connectivity")
    storagePath, err := storage.UploadFile(testContent, fmt.Sprintf("test-%d.txt", time.Now().Unix()))
    if err != nil {
        log.Printf("S3 upload error: %v", err)
        c.HTML(http.StatusInternalServerError, "index.html", gin.H{
            "error": fmt.Sprintf("S3 test failed: %v", err),
        })
        return
    }
    
    url := storage.GetFileURL(storagePath)
    log.Printf("S3 upload successful. URL: %s", url)
    c.HTML(http.StatusOK, "index.html", gin.H{
        "success": "S3 test successful",
        "url": url,
    })
}