#!/bin/sh
# manager 컨테이너 내부에서 실행됨
# DinD 노드들이 모두 준비될 때까지 대기한 뒤 Swarm 구성
set -e

NODES="manager worker1 worker2 worker3"

echo "======================================"
echo " [1/4] Docker 데몬 준비 대기"
echo "======================================"
for node in $NODES; do
  printf "  %-10s 대기 중..." "$node"
  until docker -H tcp://${node}:2375 info >/dev/null 2>&1; do
    printf "."
    sleep 2
  done
  echo " OK"
done

echo ""
echo "======================================"
echo " [2/4] Swarm 초기화 (manager)"
echo "======================================"
if docker info 2>/dev/null | grep -q "Swarm: active"; then
  echo "  이미 Swarm 초기화됨"
else
  docker swarm init --advertise-addr manager
  echo "  Swarm 초기화 완료"
fi

echo ""
echo "======================================"
echo " [3/4] Worker 노드 join"
echo "======================================"
JOIN_TOKEN=$(docker swarm join-token worker -q)
echo "  Join token: $JOIN_TOKEN"

for worker in worker1 worker2 worker3; do
  printf "  %-10s join 시도..." "$worker"
  if docker -H tcp://${worker}:2375 info 2>/dev/null | grep -q "Swarm: active"; then
    echo " 이미 참가됨"
  else
    docker -H tcp://${worker}:2375 swarm join \
      --token "$JOIN_TOKEN" \
      manager:2377
    echo " 완료"
  fi
done

echo ""
echo "======================================"
echo " [4/4] Overlay 네트워크 생성"
echo "======================================"
docker network create \
  --driver overlay \
  --attachable \
  todo-overlay 2>/dev/null && echo "  todo-overlay 생성 완료" || echo "  todo-overlay 이미 존재"

docker network create \
  --driver overlay \
  --internal \
  mysql-net 2>/dev/null && echo "  mysql-net 생성 완료" || echo "  mysql-net 이미 존재"

echo ""
echo "======================================"
echo " Swarm 클러스터 상태"
echo "======================================"
docker node ls
echo ""
echo "완료!"
