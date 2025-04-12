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
    if err := search.IndexImage(image.ID, image.Description, image.Tags); err != nil {
        log.Printf("Warning: Failed to index image in Elasticsearch: %v", err)
        // Don't return error to client as the image is already saved
    }

    c.JSON(http.StatusOK, image)
}

func GetImage(c *gin.Context) {
    idStr := c.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid image ID"})
        return
    }

    image, err := db.GetImageByID(id)
    if err != nil || image == nil {
        c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Image not found"})
        return
    }

    image.URL = storage.GetFileURL(image.StoragePath)
    c.HTML(http.StatusOK, "image.html", gin.H{"Image": image})
}

func SearchImages(c *gin.Context) {
    query := c.Query("q")
    var images []db.Image
    if query == "" {
        // Fetch the 9 most recent images from the database
        var err error
        images, err = db.GetRecentImages(9)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recent images"})
            return
        }
    } else {
        // Search in Elasticsearch
        searchResults, err := search.SearchImages(query)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search images"})
            return
        }

        // Get full image details from database
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
    }

    c.JSON(http.StatusOK, images)
}