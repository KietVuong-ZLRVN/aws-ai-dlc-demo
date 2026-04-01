-- cart_handoff database schema

CREATE DATABASE IF NOT EXISTS cart_handoff CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE cart_handoff;

CREATE TABLE IF NOT EXISTS cart_handoff_records (
    id               CHAR(36)                              NOT NULL PRIMARY KEY,
    shopper_id       VARCHAR(255)                          NOT NULL,
    source_type      ENUM('saved_combo','inline_items')    NOT NULL,
    source_combo_id  CHAR(36)                              NULL,
    status           ENUM('ok','partial','failed')         NOT NULL,
    recorded_at      DATETIME(3)                           NOT NULL,
    INDEX idx_shopper_id (shopper_id),
    INDEX idx_recorded_at (recorded_at)
);

CREATE TABLE IF NOT EXISTS handoff_record_items (
    id           BIGINT                              NOT NULL AUTO_INCREMENT PRIMARY KEY,
    record_id    CHAR(36)                            NOT NULL,
    simple_sku   VARCHAR(255)                        NOT NULL,
    outcome      ENUM('added','skipped')             NOT NULL,
    skip_reason  ENUM('out_of_stock','platform_error') NULL,
    FOREIGN KEY (record_id) REFERENCES cart_handoff_records(id) ON DELETE CASCADE,
    INDEX idx_record_id (record_id)
);
