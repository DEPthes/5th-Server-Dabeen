#!/bin/bash
# 호스트에서 실행: 서비스 이미지를 빌드하고 로컬 레지스트리에 push
set -e

REGISTRY="${REGISTRY:-localhost:5000}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

echo "========================================"
echo " 이미지 빌드 & Push → $REGISTRY"
echo "========================================"

build_push() {
  local name=$1
  local path=$2
  echo ""
  echo "── $name ──────────────────────────────"
  docker build -t "$REGISTRY/$name:latest" "$path"
  docker push "$REGISTRY/$name:latest"
  echo "  $name 완료"
}

build_push "mysql-master"  "$ROOT_DIR/services/mysql/master"
build_push "mysql-slave"   "$ROOT_DIR/services/mysql/slave"
build_push "todoapi"       "$ROOT_DIR/services/todoapi"
build_push "frontend"      "$ROOT_DIR/services/frontend"
build_push "nginx"         "$ROOT_DIR/services/nginx"
build_push "haproxy"       "$ROOT_DIR/services/haproxy"

echo ""
echo "========================================"
echo " 레지스트리 이미지 목록"
echo "========================================"
curl -s "http://${REGISTRY}/v2/_catalog" | python3 -m json.tool 2>/dev/null || \
  curl -s "http://${REGISTRY}/v2/_catalog"
echo ""
echo "빌드 & Push 완료!"
