FROM mysql:8.0

# Slave 설정 파일
COPY my.cnf /etc/mysql/conf.d/slave.cnf

# slave-setup.sh: MySQL 초기화 후 복제 구성 수행
COPY init/slave-setup.sh /docker-entrypoint-initdb.d/slave-setup.sh
RUN chmod +x /docker-entrypoint-initdb.d/slave-setup.sh

EXPOSE 3306
