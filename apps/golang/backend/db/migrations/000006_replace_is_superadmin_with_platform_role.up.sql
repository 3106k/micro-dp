ALTER TABLE users ADD COLUMN platform_role TEXT NOT NULL DEFAULT 'user';

UPDATE users SET platform_role = 'superadmin' WHERE is_superadmin = 1;

ALTER TABLE users DROP COLUMN is_superadmin;
