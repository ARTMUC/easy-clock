CREATE TABLE IF NOT EXISTS events
(
    id         CHAR(36)     NOT NULL PRIMARY KEY,
    child_id   CHAR(36)     NOT NULL,
    date       DATE         NOT NULL,
    from_time  TIME         NOT NULL,
    to_time    TIME         NOT NULL,
    label      VARCHAR(200) NOT NULL,
    emoji      VARCHAR(10),
    image_path VARCHAR(500),
    profile_id CHAR(36),
    version    INT UNSIGNED NOT NULL DEFAULT 0,
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME     NULL,
    FOREIGN KEY (child_id) REFERENCES children (id) ON DELETE CASCADE,
    FOREIGN KEY (profile_id) REFERENCES profiles (id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS event_activities
(
    id         CHAR(36)         NOT NULL PRIMARY KEY,
    event_id   CHAR(36)         NOT NULL,
    emoji      VARCHAR(10)      NOT NULL,
    label      VARCHAR(100)     NOT NULL,
    to_hour    TINYINT UNSIGNED NOT NULL,
    image_path VARCHAR(500),
    version    INT UNSIGNED     NOT NULL DEFAULT 0,
    created_at DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME         NULL,
    FOREIGN KEY (event_id) REFERENCES events (id) ON DELETE CASCADE
);