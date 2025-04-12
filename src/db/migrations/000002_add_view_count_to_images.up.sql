ALTER TABLE images ADD COLUMN view_count INTEGER DEFAULT 0;
CREATE INDEX idx_images_view_count ON images (view_count);