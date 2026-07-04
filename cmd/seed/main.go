// Command seed loads seed/catalog.json into the catalog tables. Idempotent.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	mysqladapter "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/mysql"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/catalog"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/config"
)

type catalogSeed struct {
	Products    []catalog.SeedProduct    `json:"products"`
	Colorways   []catalog.SeedColourway  `json:"colorways"`
	Skus        []catalog.SeedSKU        `json:"skus"`
	Categories  []catalog.SeedCategory   `json:"categories"`
	Collections []catalog.SeedCollection `json:"collections"`
	SizeScales  []catalog.SeedSizeScale  `json:"sizeScales"`
}

func main() {
	if err := run(); err != nil {
		log.Printf("seed failed: %v", err)
		os.Exit(1)
	}
}

func run() error {
	path := "seed/catalog.json"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	var data catalogSeed
	if err := json.Unmarshal(raw, &data); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	ctx := context.Background()
	db, err := mysqladapter.Open(ctx, cfg.Database.DSN, nil)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer func() { _ = mysqladapter.Close(db) }()

	repo := mysqladapter.NewCatalogRepository(db)
	if err := repo.SeedCatalog(ctx, data.Products, data.Colorways, data.Skus,
		data.Categories, data.Collections, data.SizeScales); err != nil {
		return fmt.Errorf("seed catalog: %w", err)
	}
	log.Printf("seeded %d products, %d colorways, %d skus, %d categories, %d collections, %d size scales",
		len(data.Products), len(data.Colorways), len(data.Skus),
		len(data.Categories), len(data.Collections), len(data.SizeScales))
	return nil
}
