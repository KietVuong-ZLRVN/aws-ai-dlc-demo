-- combo_portfolio database schema

CREATE DATABASE IF NOT EXISTS combo_portfolio CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE combo_portfolio;

CREATE TABLE IF NOT EXISTS combos (
    id           CHAR(36)                     NOT NULL PRIMARY KEY,
    shopper_id   VARCHAR(255)                 NOT NULL,
    name         VARCHAR(100)                 NOT NULL,
    visibility   ENUM('public','private')     NOT NULL DEFAULT 'private',
    share_token  CHAR(36)                     NULL,
    created_at   DATETIME(3)                  NOT NULL,
    updated_at   DATETIME(3)                  NOT NULL,
    INDEX idx_shopper_id (shopper_id),
    UNIQUE INDEX idx_share_token (share_token)
);

CREATE TABLE IF NOT EXISTS combo_items (
    id           BIGINT                       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    combo_id     CHAR(36)                     NOT NULL,
    config_sku   VARCHAR(255)                 NOT NULL,
    simple_sku   VARCHAR(255)                 NOT NULL,
    name         VARCHAR(500)                 NOT NULL,
    image_url    TEXT                         NOT NULL,
    price        DECIMAL(10,2)                NOT NULL,
    sort_order   TINYINT                      NOT NULL DEFAULT 0,
    FOREIGN KEY (combo_id) REFERENCES combos(id) ON DELETE CASCADE,
    INDEX idx_combo_id (combo_id)
);
