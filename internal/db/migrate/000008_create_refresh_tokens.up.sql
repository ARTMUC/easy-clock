CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         CHAR(36)  NOT NULL,
    user_id    CHAR(36)  NOT NULL,
    token_hash CHAR(64)  NOT NULL,
    expires_at DATETIME  NOT NULL,
    created_at DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_rt_hash (token_hash),
    INDEX      idx_rt_user_id (user_id),
    CONSTRAINT fk_rt_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
