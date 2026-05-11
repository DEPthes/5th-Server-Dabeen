#!/bin/bash
# 전체 배포 자동화 스크립트
# 실행: bash scripts/deploy.sh
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

step() { echo ""; echo "══════════════════════════════════════════"; echo " $*"; echo "══════════════════════════════════════════"; }
ok()   { echo "  ✓ $*"; }
info() { echo "  → $*"; }

cd "$ROOT_DIR"

step "STEP 1/5  DinD 클러스터 시작"
docker compose up -d
ok "컨테이너 시작됨"

step "STEP 2/5  이미지 빌드 & Push"
bash scripts/build-push.sh

step "STEP 3/5  Swarm 초기화 & Worker join"
info "DinD 노드 준비 대기 중 (최대 60s)..."
sleep 10
docker exec manager sh /scripts/init-swarm.sh
ok "Swarm 구성 완료"

step "STEP 4/5  Stack 배포"
docker exec manager docker stack deploy \
  --with-registry-auth \
  -c /stack/todo-stack.yml todo
ok "Stack 배포 요청 완료"

step "STEP 5/5  서비스 상태 확인"
info "서비스 시작 대기 (30s)..."
sleep 30
docker exec manager docker service ls

echo ""
echo "══════════════════════════════════════════"
echo " 배포 완료!"
echo "══════════════════════════════════════════"
echo ""
echo "  앱          : http://localhost"
echo "  API         : http://localhost/api/todos"
echo "  HAProxy 통계 : http://localhost:8404/stats"
echo "  Registry    : http://localhost:5000/v2/_catalog"
echo ""
echo " Swarm 노드 확인:"
docker exec manager docker node ls
echo ""
echo " 서비스 목록:"
docker exec manager docker service ls
