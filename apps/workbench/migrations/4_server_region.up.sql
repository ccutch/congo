ALTER TABLE servers ADD COLUMN region TEXT DEFAULT '';

UPDATE servers SET region = 'sfo2' WHERE region IS NULL;
