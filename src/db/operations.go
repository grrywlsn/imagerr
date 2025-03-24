package db

import (
    "database/sql"
    "encoding/json"
)

func CreateImage(filename, description, storagePath string, tags []string) (*Image, error) {
    var img Image
    err := DB.QueryRow(`
        INSERT INTO images (filename, description, tags, storage_path)
        VALUES ($1, $2, $3, $4)
        RETURNING id, filename, description, tags, storage_path, created_at
    `, filename, description, tags, storagePath).Scan(
        &img.ID,
        &img.Filename,
        &img.Description,
        &img.Tags,
        &img.StoragePath,
        &img.CreatedAt,
    )
    if err != nil {
        return nil, err
    }
    return &img, nil
}

func GetImageByID(id int64) (*Image, error) {
    var img Image
    err := DB.QueryRow(`
        SELECT id, filename, description, tags, storage_path, created_at
        FROM images WHERE id = $1
    `, id).Scan(
        &img.ID,
        &img.Filename,
        &img.Description,
        &img.Tags,
        &img.StoragePath,
        &img.CreatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    return &img, nil
}

func SearchImages(query string) ([]Image, error) {
    rows, err := DB.Query(`
        SELECT id, filename, description, tags, storage_path, created_at
        FROM images
        WHERE description ILIKE $1 OR $1 = ANY(tags)
    `, "%"+query+"%")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var images []Image
    for rows.Next() {
        var img Image
        err := rows.Scan(
            &img.ID,
            &img.Filename,
            &img.Description,
            &img.Tags,
            &img.StoragePath,
            &img.CreatedAt,
        )
        if err != nil {
            return nil, err
        }
        images = append(images, img)
    }
    return images, nil
}