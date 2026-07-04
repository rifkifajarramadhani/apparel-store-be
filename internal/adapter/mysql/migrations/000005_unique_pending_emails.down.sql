ALTER TABLE users
  DROP INDEX pending_email_unique,
  DROP COLUMN pending_email_unique;
