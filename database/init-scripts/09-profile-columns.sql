\c taaza_users;
ALTER TABLE users ADD COLUMN IF NOT EXISTS profile_picture TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS unique_users_mobile ON users(mobile) WHERE mobile IS NOT NULL;
