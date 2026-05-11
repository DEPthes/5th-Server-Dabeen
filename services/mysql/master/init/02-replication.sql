-- Slave가 사용할 Replication 전용 계정 생성
CREATE USER IF NOT EXISTS 'replicator'@'%'
  IDENTIFIED WITH mysql_native_password BY 'repl_pass_123';

GRANT REPLICATION SLAVE ON *.* TO 'replicator'@'%';

FLUSH PRIVILEGES;

-- 확인
SHOW MASTER STATUS;
