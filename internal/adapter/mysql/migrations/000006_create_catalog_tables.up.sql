CREATE TABLE brands (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, public_id CHAR(26) NOT NULL, slug VARCHAR(191) NOT NULL, name VARCHAR(191) NOT NULL,
  archived_at DATETIME(6) NULL, created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  UNIQUE KEY uq_brands_public_id(public_id), UNIQUE KEY uq_brands_slug(slug)
);
CREATE TABLE categories (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, public_id CHAR(26) NOT NULL, parent_id BIGINT UNSIGNED NULL, slug VARCHAR(191) NOT NULL, name VARCHAR(191) NOT NULL, gender VARCHAR(16) NULL,
  archived_at DATETIME(6) NULL, created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  UNIQUE KEY uq_categories_public_id(public_id), UNIQUE KEY uq_categories_slug(slug), CONSTRAINT fk_categories_parent FOREIGN KEY(parent_id) REFERENCES categories(id) ON DELETE RESTRICT
);
CREATE TABLE collections (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, public_id CHAR(26) NOT NULL, slug VARCHAR(191) NOT NULL, name VARCHAR(191) NOT NULL,
  archived_at DATETIME(6) NULL, created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  UNIQUE KEY uq_collections_public_id(public_id), UNIQUE KEY uq_collections_slug(slug)
);
CREATE TABLE size_scales (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, public_id CHAR(26) NOT NULL, code VARCHAR(64) NOT NULL, name VARCHAR(191) NOT NULL,
  archived_at DATETIME(6) NULL, created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  UNIQUE KEY uq_size_scales_public_id(public_id), UNIQUE KEY uq_size_scales_code(code)
);
CREATE TABLE sizes (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, public_id CHAR(26) NOT NULL, size_scale_id BIGINT UNSIGNED NOT NULL, code VARCHAR(32) NOT NULL, name VARCHAR(80) NOT NULL, sort_order SMALLINT UNSIGNED NOT NULL DEFAULT 0,
  archived_at DATETIME(6) NULL, created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  UNIQUE KEY uq_sizes_public_id(public_id), UNIQUE KEY uq_sizes_scale_code(size_scale_id,code), CONSTRAINT fk_sizes_scale FOREIGN KEY(size_scale_id) REFERENCES size_scales(id) ON DELETE RESTRICT
);
CREATE TABLE colourways (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, public_id CHAR(26) NOT NULL, slug VARCHAR(191) NOT NULL, name VARCHAR(191) NOT NULL, colour_family VARCHAR(80) NULL, hex_code CHAR(7) NOT NULL,
  archived_at DATETIME(6) NULL, created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  UNIQUE KEY uq_colourways_public_id(public_id), UNIQUE KEY uq_colourways_slug(slug), CONSTRAINT chk_colourways_hex CHECK(hex_code REGEXP '^#[0-9A-Fa-f]{6}$')
);
CREATE TABLE products (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, public_id CHAR(26) NOT NULL, style_code VARCHAR(64) NOT NULL, slug VARCHAR(191) NOT NULL, brand_id BIGINT UNSIGNED NOT NULL,
  name VARCHAR(191) NOT NULL, subtitle VARCHAR(191) NOT NULL DEFAULT '', gender VARCHAR(16) NULL, product_type VARCHAR(80) NULL, description TEXT NULL, published_at DATETIME(6) NULL,
  archived_at DATETIME(6) NULL, created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  UNIQUE KEY uq_products_public_id(public_id), UNIQUE KEY uq_products_style_code(style_code), UNIQUE KEY uq_products_slug(slug),
  KEY idx_products_browse(archived_at,published_at,id), CONSTRAINT fk_products_brand FOREIGN KEY(brand_id) REFERENCES brands(id) ON DELETE RESTRICT
);
CREATE TABLE product_categories (
  product_id BIGINT UNSIGNED NOT NULL, category_id BIGINT UNSIGNED NOT NULL, is_primary BOOLEAN NOT NULL DEFAULT FALSE, created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY(product_id,category_id), CONSTRAINT fk_product_categories_product FOREIGN KEY(product_id) REFERENCES products(id) ON DELETE CASCADE,
  CONSTRAINT fk_product_categories_category FOREIGN KEY(category_id) REFERENCES categories(id) ON DELETE RESTRICT
);
CREATE TABLE product_collections (
  product_id BIGINT UNSIGNED NOT NULL, collection_id BIGINT UNSIGNED NOT NULL, created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY(product_id,collection_id), CONSTRAINT fk_product_collections_product FOREIGN KEY(product_id) REFERENCES products(id) ON DELETE CASCADE,
  CONSTRAINT fk_product_collections_collection FOREIGN KEY(collection_id) REFERENCES collections(id) ON DELETE RESTRICT
);
CREATE TABLE skus (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, public_id CHAR(26) NOT NULL, sku_code VARCHAR(80) NOT NULL, barcode VARCHAR(80) NULL,
  product_id BIGINT UNSIGNED NOT NULL, colourway_id BIGINT UNSIGNED NOT NULL, size_id BIGINT UNSIGNED NOT NULL,
  archived_at DATETIME(6) NULL, created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  UNIQUE KEY uq_skus_public_id(public_id), UNIQUE KEY uq_skus_code(sku_code), UNIQUE KEY uq_skus_variant(product_id,colourway_id,size_id),
  CONSTRAINT fk_skus_product FOREIGN KEY(product_id) REFERENCES products(id) ON DELETE RESTRICT, CONSTRAINT fk_skus_colourway FOREIGN KEY(colourway_id) REFERENCES colourways(id) ON DELETE RESTRICT,
  CONSTRAINT fk_skus_size FOREIGN KEY(size_id) REFERENCES sizes(id) ON DELETE RESTRICT
);
CREATE TABLE prices (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, public_id CHAR(26) NOT NULL, product_id BIGINT UNSIGNED NULL, sku_id BIGINT UNSIGNED NULL, currency CHAR(3) NOT NULL,
  amount BIGINT UNSIGNED NOT NULL, compare_at_amount BIGINT UNSIGNED NULL, valid_from DATETIME(6) NOT NULL, valid_to DATETIME(6) NULL,
  archived_at DATETIME(6) NULL, created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  UNIQUE KEY uq_prices_public_id(public_id), CONSTRAINT fk_prices_product FOREIGN KEY(product_id) REFERENCES products(id) ON DELETE RESTRICT,
  CONSTRAINT fk_prices_sku FOREIGN KEY(sku_id) REFERENCES skus(id) ON DELETE RESTRICT, CONSTRAINT chk_prices_owner CHECK((product_id IS NULL)<>(sku_id IS NULL)),
  CONSTRAINT chk_prices_currency CHECK(currency=UPPER(currency)), CONSTRAINT chk_prices_compare CHECK(compare_at_amount IS NULL OR compare_at_amount>=amount),
  CONSTRAINT chk_prices_interval CHECK(valid_to IS NULL OR valid_to>valid_from)
);
CREATE TABLE inventory_locations (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, public_id CHAR(26) NOT NULL, code VARCHAR(64) NOT NULL, name VARCHAR(191) NOT NULL, archived_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  UNIQUE KEY uq_inventory_locations_public_id(public_id), UNIQUE KEY uq_inventory_locations_code(code)
);
CREATE TABLE inventory_balances (
  sku_id BIGINT UNSIGNED NOT NULL, location_id BIGINT UNSIGNED NOT NULL, on_hand INT UNSIGNED NOT NULL DEFAULT 0, reserved INT UNSIGNED NOT NULL DEFAULT 0,
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6), PRIMARY KEY(sku_id,location_id),
  CONSTRAINT fk_inventory_sku FOREIGN KEY(sku_id) REFERENCES skus(id) ON DELETE RESTRICT, CONSTRAINT fk_inventory_location FOREIGN KEY(location_id) REFERENCES inventory_locations(id) ON DELETE RESTRICT,
  CONSTRAINT chk_inventory_reserved CHECK(reserved<=on_hand)
);
CREATE TABLE assets (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, public_id CHAR(26) NOT NULL, media_type ENUM('image','video','document') NOT NULL, storage_key VARCHAR(512) NULL,
  url VARCHAR(700) NOT NULL, mime_type VARCHAR(127) NULL, width INT UNSIGNED NULL, height INT UNSIGNED NULL, alt_text VARCHAR(512) NULL, archived_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  UNIQUE KEY uq_assets_public_id(public_id), UNIQUE KEY uq_assets_url(url)
);
CREATE TABLE product_assets (
  product_id BIGINT UNSIGNED NOT NULL, asset_id BIGINT UNSIGNED NOT NULL, role VARCHAR(64) NOT NULL, sort_order SMALLINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY(product_id,asset_id,role), CONSTRAINT fk_product_assets_product FOREIGN KEY(product_id) REFERENCES products(id) ON DELETE CASCADE,
  CONSTRAINT fk_product_assets_asset FOREIGN KEY(asset_id) REFERENCES assets(id) ON DELETE RESTRICT
);
CREATE TABLE sku_assets (
  sku_id BIGINT UNSIGNED NOT NULL, asset_id BIGINT UNSIGNED NOT NULL, role VARCHAR(64) NOT NULL, sort_order SMALLINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY(sku_id,asset_id,role), CONSTRAINT fk_sku_assets_sku FOREIGN KEY(sku_id) REFERENCES skus(id) ON DELETE CASCADE,
  CONSTRAINT fk_sku_assets_asset FOREIGN KEY(asset_id) REFERENCES assets(id) ON DELETE RESTRICT
);
