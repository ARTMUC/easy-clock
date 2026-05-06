ALTER TABLE users
  ADD COLUMN name               VARCHAR(100)  NOT NULL DEFAULT '' AFTER id,
  ADD COLUMN active             TINYINT(1)    NOT NULL DEFAULT 0,
  ADD COLUMN verification_token VARCHAR(64)   NOT NULL DEFAULT '';
