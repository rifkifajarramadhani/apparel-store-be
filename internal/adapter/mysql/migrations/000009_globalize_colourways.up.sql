-- Globalize colourways: merge per-product duplicate colourway rows that share
-- the same colour family into a single canonical row, then drop the
-- per-product slug/colour_family columns so colourways become a flat,
-- reusable lookup keyed only by a plain unique name.

CREATE TEMPORARY TABLE colourway_merge_map (
  old_id BIGINT UNSIGNED PRIMARY KEY,
  canonical_id BIGINT UNSIGNED NOT NULL
);

INSERT INTO colourway_merge_map (old_id, canonical_id)
SELECT c.id,
  (SELECT MIN(c2.id) FROM colourways c2
     WHERE LOWER(TRIM(COALESCE(c2.colour_family, c2.name))) = LOWER(TRIM(COALESCE(c.colour_family, c.name))))
FROM colourways c;

UPDATE skus s
JOIN colourway_merge_map m ON m.old_id = s.colourway_id
SET s.colourway_id = m.canonical_id
WHERE m.old_id <> m.canonical_id;

UPDATE assets a
JOIN colourway_merge_map m ON m.old_id = a.colourway_id
SET a.colourway_id = m.canonical_id
WHERE m.old_id <> m.canonical_id;

UPDATE colourways
SET name = TRIM(colour_family)
WHERE id IN (SELECT DISTINCT canonical_id FROM colourway_merge_map)
  AND colour_family IS NOT NULL AND TRIM(colour_family) <> '';

DELETE c FROM colourways c
JOIN colourway_merge_map m ON m.old_id = c.id
WHERE m.old_id <> m.canonical_id;

DROP TEMPORARY TABLE colourway_merge_map;

ALTER TABLE colourways
  DROP INDEX slug,
  DROP COLUMN slug,
  DROP COLUMN colour_family,
  ADD UNIQUE INDEX uq_colourways_name (name);
