CREATE TABLE IF NOT EXISTS schedule_days (
  id          CHAR(36)      NOT NULL PRIMARY KEY,
  child_id    CHAR(36)      NOT NULL,
  day_of_week TINYINT UNSIGNED NOT NULL,
  profile_id  CHAR(36)      NOT NULL,
  version     INT UNSIGNED  NOT NULL DEFAULT 0,
  created_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uq_child_day (child_id, day_of_week),
  FOREIGN KEY (child_id)   REFERENCES children(id) ON DELETE CASCADE,
  FOREIGN KEY (profile_id) REFERENCES profiles(id)
);
