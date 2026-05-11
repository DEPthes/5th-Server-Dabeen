FROM haproxy:3.0-alpine

# 503 에러 페이지 디렉토리 생성
RUN mkdir -p /usr/local/etc/haproxy/errors

# 커스텀 503 페이지
RUN printf 'HTTP/1.0 503 Service Unavailable\r\nContent-Type: text/plain\r\n\r\n서비스가 일시적으로 불가합니다. 잠시 후 다시 시도해주세요.\n' \
    > /usr/local/etc/haproxy/errors/503.http

COPY haproxy.cfg /usr/local/etc/haproxy/haproxy.cfg

EXPOSE 80 8404
HEALTHCHECK --interval=10s --timeout=5s --retries=3 \
  CMD wget -qO- http://localhost:8404/stats || exit 1
