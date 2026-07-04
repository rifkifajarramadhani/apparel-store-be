CREATE TABLE orders (
  id INT PRIMARY KEY AUTO_INCREMENT,
  user_id INT NOT NULL,
  status VARCHAR(20) NOT NULL,
  total INT NOT NULL DEFAULT 0,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  INDEX idx_orders_user (user_id)
);

CREATE TABLE order_items (
  id INT PRIMARY KEY AUTO_INCREMENT,
  order_id INT NOT NULL,
  sku_id VARCHAR(80) NOT NULL,
  product_id VARCHAR(64) NOT NULL,
	 sku_ref_id BIGINT UNSIGNED NULL,
	 product_ref_id BIGINT UNSIGNED NULL,
  name VARCHAR(191) NOT NULL,
  size VARCHAR(16) NOT NULL,
  unit_price INT NOT NULL DEFAULT 0,
  qty INT NOT NULL DEFAULT 0,
	 INDEX idx_order_items_order (order_id),
	 INDEX idx_order_items_sku_ref (sku_ref_id),
	 INDEX idx_order_items_product_ref (product_ref_id),
	 CONSTRAINT fk_order_items_sku_ref FOREIGN KEY (sku_ref_id) REFERENCES skus(id) ON DELETE SET NULL,
	 CONSTRAINT fk_order_items_product_ref FOREIGN KEY (product_ref_id) REFERENCES products(id) ON DELETE SET NULL
);
