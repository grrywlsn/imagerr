package db

type Image struct {
    ID          int64    `json:"id"`
    Filename    string   `json:"filename"`
    Description string   `json:"description"`
    Tags        []string `json:"tags"`
    StoragePath string   `json:"storage_path"`
    CreatedAt   string   `json:"created_at"`
    URL         string   `json:"url,omitempty"` // Added URL field
}