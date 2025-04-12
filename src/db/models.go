package db

type Image struct {
    ID              int64    `json:"id"`
    OriginalFilename string   `json:"original_filename"`
    UUIDFilename    string   `json:"uuid_filename"`
    Description     string   `json:"description"`
    Tags            []string `json:"tags"`
    StoragePath     string   `json:"storage_path"`
    CreatedAt       string   `json:"created_at"`
    URL             string   `json:"url,omitempty"`
}