ALTER TABLE images DROP COLUMN view_count;
DROP INDEX IF EXISTS idx_images_view_count;