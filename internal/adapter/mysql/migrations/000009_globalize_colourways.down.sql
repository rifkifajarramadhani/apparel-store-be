-- NOTE: this rollback restores the `slug`/`colour_family` columns and
-- best-effort backfills `slug` from `name`, but it CANNOT restore the
-- duplicate per-product colourway rows deleted by the up migration, nor
-- their original compound names/per-product associations. Rolling back
-- recreates the old schema shape, not the old data.

ALTER TABLE colourways
  DROP INDEX uq_colourways_name,
  ADD COLUMN colour_family VARCHAR(80) NULL AFTER name,
  ADD COLUMN slug VARCHAR(191) NULL AFTER public_id;

UPDATE colourways SET slug = LOWER(REPLACE(name, ' ', '-'));

ALTER TABLE colourways
  MODIFY COLUMN slug VARCHAR(191) NOT NULL,
  ADD UNIQUE (slug);
