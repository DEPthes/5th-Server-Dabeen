#!/bin/bash
# docker-entrypoint-initdb.d 에 의해 MySQL 초기화 완료 후 실행됨
set -e

MASTER_HOST="${MYSQL_MASTER_HOST:-mysql-master}"
REPL_USER="replicator"
REPL_PASS="${REPLICATION_PASSWORD:-repl_pass_123}"
MYSQL_ROOT="${MYSQL_ROOT_PASSWORD:-root_pass_123}"

echo "[slave-setup] Master($MASTER_HOST) 연결 대기 중..."

# Master MySQL이 준비될 때까지 대기
MAX_RETRY=60
COUNT=0
until mysql -h "$MASTER_HOST" -u "$REPL_USER" -p"$REPL_PASS" \
      -e "SELECT 1" >/dev/null 2>&1; do
  COUNT=$((COUNT + 1))
  if [ "$COUNT" -ge "$MAX_RETRY" ]; then
    echo "[slave-setup] ERROR: Master에 연결할 수 없습니다. 타임아웃."
    exit 1
  fi
  echo "[slave-setup] 재시도 $COUNT/$MAX_RETRY..."
  sleep 5
done

echo "[slave-setup] Master 연결 성공. 복제 설정 중..."

mysql -u root -p"$MYSQL_ROOT" <<SQL
-- 기존 복제 설정 초기화
STOP REPLICA;
RESET REPLICA ALL;

-- GTID 기반 복제 설정 (binlog position 불필요)
CHANGE REPLICATION SOURCE TO
  SOURCE_HOST      = '${MASTER_HOST}',
  SOURCE_PORT      = 3306,
  SOURCE_USER      = '${REPL_USER}',
  SOURCE_PASSWORD  = '${REPL_PASS}',
  SOURCE_AUTO_POSITION = 1,
  GET_SOURCE_PUBLIC_KEY = 1;

START REPLICA;
SQL

echo "[slave-setup] 복제 시작 완료!"
echo "[slave-setup] 복제 상태:"
mysql -u root -p"$MYSQL_ROOT" -e "SHOW REPLICA STATUS\G" 2>/dev/null | \
  grep -E "Replica_IO_Running|Replica_SQL_Running|Seconds_Behind|Last_Error" || true
