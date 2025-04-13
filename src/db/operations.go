package db

import (
    "database/sql"
    "log"
    "github.com/lib/pq"
)

func CreateImage(originalFilename, uuidFilename, description, storagePath string, tags []string) (*Image, error) {
    var img Image
    err := DB.QueryRow(`
        INSERT INTO images (original_filename, uuid_filename, description, tags, storage_path)
        VALUES ($1, $2, $3, $4::text[], $5)
        RETURNING id, original_filename, uuid_filename, description, tags, storage_path, created_at
    `, originalFilename, uuidFilename, description, pq.Array(tags), storagePath).Scan(
        &img.ID,
        &img.OriginalFilename,
        &img.UUIDFilename,
        &img.Description,
        pq.Array(&img.Tags),
        &img.StoragePath,
        &img.CreatedAt,
    )
    if err != nil {
        log.Printf("Error creating image record: %v\nParams: filename=%s, uuid=%s, path=%s, tags=%v", 
            err, originalFilename, uuidFilename, storagePath, tags)
        return nil, err
    }
    return &img, nil
}

func GetImageByID(id int64) (*Image, error) {
    var img Image
    err := DB.QueryRow(`
        SELECT id, original_filename, uuid_filename, description, tags, storage_path, created_at
        FROM images WHERE id = $1
    `, id).Scan(
        &img.ID,
        &img.OriginalFilename,
        &img.UUIDFilename,
        &img.Description,
        pq.Array(&img.Tags),
        &img.StoragePath,
        &img.CreatedAt,
    )
    if err == sql.ErrNoRows {
        log.Printf("No image found with ID: %d", id)
        return nil, nil
    }
    if err != nil {
        log.Printf("Error retrieving image with ID %d: %v", id, err)
        return nil, err
    }
    return &img, nil
}

func GetRecentImages(limit int) ([]Image, error) {
    rows, err := DB.Query(`
        SELECT id, original_filename, uuid_filename, description, tags, storage_path, created_at
        FROM images
        ORDER BY created_at DESC
        LIMIT $1
    `, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var images []Image
    for rows.Next() {
        var img Image
        err := rows.Scan(
            &img.ID,
            &img.OriginalFilename,
            &img.UUIDFilename,
            &img.Description,
            pq.Array(&img.Tags),
            &img.StoragePath,
            &img.CreatedAt,
        )
        if err != nil {
            return nil, err
        }
        images = append(images, img)
    }
    if err = rows.Err(); err != nil {
        return nil, err
    }
    return images, nil
}

func SearchImages(query string) ([]Image, error) {
    // Add logging to debug the search query
    log.Printf("Searching for: %s", query)
    
    rows, err := DB.Query(`
        SELECT id, original_filename, uuid_filename, description, tags, storage_path, created_at
        FROM images
        WHERE description ILIKE $1 OR EXISTS (
            SELECT 1 FROM unnest(tags) tag WHERE tag ILIKE $1
        )
    `, "%"+query+"%")
    
    if err != nil {
        log.Printf("Search query error: %v", err)
        return nil, err
    }
    defer rows.Close()

    var images []Image
    for rows.Next() {
        var img Image
        err := rows.Scan(
            &img.ID,
            &img.OriginalFilename,
            &img.UUIDFilename,
            &img.Description,
            pq.Array(&img.Tags),
            &img.StoragePath,
            &img.CreatedAt,
        )
        if err != nil {
            return nil, err
        }
        images = append(images, img)
    }
    if err = rows.Err(); err != nil {
        return nil, err
    }
    return images, nil
}

func GetAllImages() ([]Image, error) {
    rows, err := DB.Query(`
        SELECT id, original_filename, uuid_filename, description, tags, storage_path, created_at
        FROM images
        ORDER BY id ASC
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var images []Image
    for rows.Next() {
        var img Image
        err := rows.Scan(
            &img.ID,
            &img.OriginalFilename,
            &img.UUIDFilename,
            &img.Description,
            pq.Array(&img.Tags),
            &img.StoragePath,
            &img.CreatedAt,
        )
        if err != nil {
            return nil, err
        }
        images = append(images, img)
    }
    if err = rows.Err(); err != nil {
        return nil, err
    }
    return images, nil
}