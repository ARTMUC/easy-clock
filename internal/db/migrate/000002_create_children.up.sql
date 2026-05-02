CREATE TABLE IF NOT EXISTS children (
  id                 CHAR(36)     NOT NULL PRIMARY KEY,
  user_id            CHAR(36)     NOT NULL,
  name               VARCHAR(100) NOT NULL,
  timezone           VARCHAR(100) NOT NULL DEFAULT 'Europe/Warsaw',
  avatar_path        VARCHAR(500),
  default_profile_id CHAR(36),
  clock_token        CHAR(64)     NOT NULL UNIQUE,
  version            INT UNSIGNED NOT NULL DEFAULT 0,
  created_at         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
