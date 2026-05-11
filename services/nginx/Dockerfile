FROM nginx:1.27-alpine

# 기본 설정 제거 후 커스텀 설정 적용
RUN rm /etc/nginx/conf.d/default.conf
COPY nginx.conf /etc/nginx/conf.d/todo.conf

EXPOSE 80
HEALTHCHECK --interval=10s --timeout=5s --retries=3 \
  CMD wget -qO- http://localhost/nginx-status || exit 1
