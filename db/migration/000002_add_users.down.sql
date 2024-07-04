ALTER TABLE IF EXISTS "acccounts" DROP CONSTRAINT IF EXISTS "owner_cuurency_key";

ALTER TABLE IF EXISTS "acccounts" DROP CONSTRAINT IF EXISTS "accounts_owner_fkey";

DROP TABLE IF EXISTS "users"
