FROM mysql:8.0

# Master 설정 파일
COPY my.cnf /etc/mysql/conf.d/master.cnf

# 초기화 SQL (컨테이너 최초 기동 시 실행)
COPY init/ /docker-entrypoint-initdb.d/

EXPOSE 3306
