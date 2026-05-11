-- tododb 스키마 생성
CREATE DATABASE IF NOT EXISTS tododb
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE tododb;

CREATE TABLE IF NOT EXISTS todos (
  id         BIGINT       NOT NULL AUTO_INCREMENT,
  title      VARCHAR(255) NOT NULL,
  done       TINYINT(1)   NOT NULL DEFAULT 0,
  created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX idx_created_at (created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 샘플 데이터
INSERT INTO todos (title) VALUES
  ('Docker Swarm 실습 환경 구성하기'),
  ('MySQL Master-Slave 복제 확인'),
  ('Go REST API 동작 테스트'),
  ('Nuxt.js 프론트엔드 접속 확인'),
  ('HAProxy 통계 페이지 확인');
