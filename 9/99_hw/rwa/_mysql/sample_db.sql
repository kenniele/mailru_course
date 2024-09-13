-- sample_db.sql

CREATE DATABASE IF NOT EXISTS photolist;

USE photolist;

DROP TABLE IF EXISTS `profiles`;
CREATE TABLE `profiles`
(
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    email      varchar(255) NOT NULL,
    username   varchar(255) NOT NULL,
    password   varchar(255) NOT NULL,
    bio        text,
    image      text,
    created_at varchar(255) NOT NULL,
    updated_at varchar(255) NOT NULL,
    token      varchar(255),
    following  boolean DEFAULT FALSE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

DROP TABLE IF EXISTS `articles`;
CREATE TABLE `articles`
(
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    author_id       BIGINT,
    body            TEXT                NOT NULL,
    created_at      VARCHAR(255)            NOT NULL,
    updated_at      VARCHAR(255)            NOT NULL,
    description     TEXT,
    favorited       BOOLEAN DEFAULT FALSE,
    favorites_count INT     DEFAULT 0,
    slug            VARCHAR(255) UNIQUE NOT NULL,
    title           VARCHAR(255)        NOT NULL,
    FOREIGN KEY (author_id) REFERENCES profiles (id) ON DELETE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

DROP TABLE IF EXISTS `tags`;
CREATE TABLE tags
(
    id   BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL
);

DROP TABLE IF EXISTS `article_tags`;
CREATE TABLE article_tags
(
    article_id BIGINT,
    tag_id     BIGINT,
    PRIMARY KEY (article_id, tag_id),
    FOREIGN KEY (article_id) REFERENCES articles (id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags (id) ON DELETE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

DROP TABLE IF EXISTS `sessions`;
CREATE TABLE `sessions`
(
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    profile_id BIGINT NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    FOREIGN KEY (profile_id) REFERENCES profiles (id) ON DELETE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;
