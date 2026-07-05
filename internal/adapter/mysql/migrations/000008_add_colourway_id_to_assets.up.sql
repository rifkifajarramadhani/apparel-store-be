ALTER TABLE assets
  ADD COLUMN colourway_id BIGINT UNSIGNED NULL AFTER sku_id,
  ADD CONSTRAINT fk_assets_colourway_id FOREIGN KEY (colourway_id) REFERENCES colourways(id) ON DELETE CASCADE;
