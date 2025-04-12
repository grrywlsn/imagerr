CREATE TABLE IF NOT EXISTS images (
    id SERIAL PRIMARY KEY,
    original_filename VARCHAR(255) NOT NULL,
    uuid_filename VARCHAR(255) NOT NULL,
    description TEXT,
    tags TEXT[],
    storage_path VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_images_tags ON images USING GIN (tags);