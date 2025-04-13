package api

import (
    "log"
    "net/http"
    "strings"
    "path/filepath"
    "github.com/gin-gonic/gin"
    "github.com/grrywlsn/imagerr/src/db"
    "github.com/grrywlsn/imagerr/src/storage"
    "github.com/grrywlsn/imagerr/src/search"
    "strconv"
    "github.com/google/uuid"
)

func SuggestTags(c *gin.Context) {
    query := c.Query("q")
    if query == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
        return
    }

    tags, err := search.SuggestTags(query)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tag suggestions"})
        return
    }

    c.JSON(http.StatusOK, tags)
}

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

    // Generate UUID for filename
    originalFilename := filepath.Base(header.Filename)
    fileExt := filepath.Ext(originalFilename)
    uuidFilename := uuid.New().String() + fileExt

    // Upload to S3
    storagePath, err := storage.UploadFile(file, uuidFilename)
    if err != nil {
        log.Printf("Error uploading file to S3: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
        return
    }

    // Save to database with both original and UUID filenames
    image, err := db.CreateImage(originalFilename, uuidFilename, description, storagePath, tags)
    if err != nil {
        log.Printf("Error saving to database: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save to database"})
        return
    }

    // Index in Elasticsearch
    if err := search.IndexImage(image); err != nil {
        log.Printf("Warning: Failed to index image in Elasticsearch: %v", err)
        // Don't return error to client as the image is already saved
    }

    c.JSON(http.StatusOK, image)
}

func GetImage(c *gin.Context) {
    idStr := c.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid image ID", "message": err.Error()})
        return
    }

    image, err := db.GetImageByID(id)
    if err != nil || image == nil {
        errorMessage := "Image not found"
        if err != nil {
            errorMessage = err.Error()
        }
        c.HTML(http.StatusNotFound, "error.html", gin.H{"error": errorMessage})
        return
    }

    image.URL = storage.GetFileURL(image.StoragePath)
    c.HTML(http.StatusOK, "image.html", gin.H{"Image": image})
}

func SearchImages(c *gin.Context) {
    query := c.Query("q")
    tags := c.Query("tags")
    var images []db.Image
    if query == "" && tags == "" {
        // Fetch the 9 most recent images from the database
        var err error
        images, err = db.GetRecentImages(9)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recent images"})
            return
        }
    } else {
        // Search in Elasticsearch
        searchResults, err := search.SearchImages(query, tags)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search images"})
            return
        }

        // Convert search results to images and construct URLs
        for _, result := range searchResults {
            image := db.Image{
                ID:               result.ID,
                OriginalFilename: result.OriginalFilename,
                UUIDFilename:     result.UUIDFilename,
                Description:      result.Description,
                Tags:             result.Tags,
                StoragePath:      result.StoragePath,
                CreatedAt:        result.CreatedAt,
                ViewCount:        result.ViewCount,
            }
            image.URL = storage.GetFileURL(result.StoragePath)
            images = append(images, image)
        }
    }

    c.JSON(http.StatusOK, images)
}

func ReindexImages(c *gin.Context) {
    // Get all images from database
    images, err := db.GetAllImages()
    if err != nil {
        log.Printf("Error fetching images from database: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch images"})
        return
    }

    // Reindex all images in Elasticsearch
    if err := search.ReindexAll(images); err != nil {
        log.Printf("Error reindexing images: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reindex images"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Successfully reindexed all images"})
}