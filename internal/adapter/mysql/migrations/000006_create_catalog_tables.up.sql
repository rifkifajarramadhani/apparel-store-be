CREATE TABLE categories (
  id VARCHAR(64) PRIMARY KEY,
  slug VARCHAR(64) NOT NULL,
  name VARCHAR(80) NOT NULL,
  parent_id VARCHAR(64) NULL,
  gender VARCHAR(16) NOT NULL,
  level INT NOT NULL DEFAULT 0,
  INDEX idx_categories_parent (parent_id)
);

CREATE TABLE collections (
  id VARCHAR(64) PRIMARY KEY,
  slug VARCHAR(64) NOT NULL,
  name VARCHAR(80) NOT NULL
);

CREATE TABLE size_scales (
  id VARCHAR(64) PRIMARY KEY,
  sizes JSON NULL
);

CREATE TABLE products (
  id VARCHAR(64) PRIMARY KEY,
  slug VARCHAR(191) NOT NULL,
  name VARCHAR(191) NOT NULL,
  subtitle VARCHAR(191) NOT NULL,
  brand VARCHAR(80) NOT NULL,
  gender VARCHAR(16) NOT NULL,
  `type` VARCHAR(80) NOT NULL,
  category_id VARCHAR(64) NOT NULL,
  category_slug VARCHAR(64) NOT NULL,
  collection_ids JSON NULL,
  size_scale VARCHAR(64) NOT NULL,
  base_price INT NOT NULL DEFAULT 0,
  min_price INT NOT NULL DEFAULT 0,
  max_price INT NOT NULL DEFAULT 0,
  badges JSON NULL,
  colorway_count INT NOT NULL DEFAULT 0,
  color_families JSON NULL,
  swatches JSON NULL,
  thumbnail_url VARCHAR(512) NOT NULL DEFAULT '',
  hover_image_url VARCHAR(512) NOT NULL DEFAULT '',
  default_colorway_id VARCHAR(64) NOT NULL DEFAULT '',
  sizes JSON NULL,
  description TEXT NULL,
  published_at VARCHAR(32) NOT NULL DEFAULT '',
  INDEX idx_products_slug (slug),
  INDEX idx_products_gender (gender),
  INDEX idx_products_category (category_id),
  INDEX idx_products_category_slug (category_slug),
  INDEX idx_products_min_price (min_price),
  INDEX idx_products_published (published_at)
);

CREATE TABLE colorways (
  id VARCHAR(64) PRIMARY KEY,
  product_id VARCHAR(64) NOT NULL,
  style_color VARCHAR(64) NOT NULL,
  name VARCHAR(191) NOT NULL,
  color_family VARCHAR(40) NOT NULL,
  swatch_hex VARCHAR(16) NOT NULL,
  price INT NOT NULL DEFAULT 0,
  is_default BOOLEAN NOT NULL DEFAULT FALSE,
  on_sale BOOLEAN NOT NULL DEFAULT FALSE,
  images JSON NULL,
  INDEX idx_colorways_product (product_id)
);

CREATE TABLE skus (
  id VARCHAR(80) PRIMARY KEY,
  colorway_id VARCHAR(64) NOT NULL,
  product_id VARCHAR(64) NOT NULL,
  size VARCHAR(16) NOT NULL,
  size_label VARCHAR(32) NOT NULL,
  size_scale VARCHAR(64) NOT NULL,
  in_stock BOOLEAN NOT NULL DEFAULT TRUE,
  stock_qty INT NOT NULL DEFAULT 0,
  price INT NOT NULL DEFAULT 0,
  INDEX idx_skus_colorway (colorway_id),
  INDEX idx_skus_product (product_id)
);
