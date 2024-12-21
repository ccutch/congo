ALTER TABLE servers ADD COLUMN ip_address TEXT DEFAULT '';
ALTER TABLE servers ADD COLUMN error      TEXT DEFAULT '';

UPDATE servers SET ip_address = '' WHERE ip_address IS NULL;
UPDATE servers SET error = '' WHERE error IS NULL;
