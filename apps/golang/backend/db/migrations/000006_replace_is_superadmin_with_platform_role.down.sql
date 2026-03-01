ALTER TABLE users ADD COLUMN is_superadmin INTEGER NOT NULL DEFAULT 0;

UPDATE users SET is_superadmin = 1 WHERE platform_role = 'superadmin';

ALTER TABLE users DROP COLUMN platform_role;
