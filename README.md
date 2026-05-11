# 5th-Server-Dabeen
## Docker Swarm 기반 TODO 애플리케이션 실습 환경

---

## 아키텍처

```
호스트 (Mac/Linux)
  │
  ├── localhost:80   ──────────────────────────────────┐
  ├── localhost:8404 (HAProxy Stats)                   │
  └── localhost:5000 (Private Registry)                │
                                                        │
  ┌─────────── dind-net (bridge) ─────────────────────┐│
  │                                                    ││
  │  ┌──────────────────────────────────────────────┐ ││
  │  │  manager (Docker Swarm Manager)              │ ││
  │  │                                              │ ││
  │  │  [todo-overlay network]                      │ ││
  │  │   HAProxy :80 ──→ nginx (×2 replicas) ──→ todoapi (×2) ──→ MySQL Master
  │  │                         └──────────────────→ frontend (×2)   ↑ replication
  │  │                                                               MySQL Slave
  │  │  [mysql-net - internal overlay]                               │
  │  │   mysql-master ←─────────────────────────────────────────────┘
  │  │   mysql-slave  (worker 노드에 배치)
  │  └──────────────────────────────────────────────┘
  │
  │  worker1 / worker2 / worker3 (DinD)
  └────────────────────────────────────────────────────┘
```

### 서비스 구성

| 서비스 | 이미지 | 역할 | 레플리카 | 배치 |
|--------|--------|------|----------|------|
| `haproxy` | `haproxy:3.0-alpine` | 외부 진입점, L7 LB | 1 | manager |
| `nginx` | `nginx:1.27-alpine` | Reverse Proxy (API/Frontend 라우팅) | 2 | any |
| `todoapi` | Go 1.23 | REST API (CRUD) | 2 | any |
| `frontend` | Nuxt.js 3 / Node 22 | SSR 프론트엔드 | 2 | any |
| `mysql-master` | MySQL 8.0 | DB 쓰기 (GTID 복제 소스) | 1 | manager |
| `mysql-slave` | MySQL 8.0 | DB 읽기 (GTID 복제 대상) | 1 | worker |

---

## 디렉토리 구조

```
5th-Server-Dabeen/
├── docker-compose.yml        # DinD 클러스터 (manager×1 + worker×3 + registry×1)
├── .env.example              # 환경변수 예시
├── scripts/
│   ├── init-swarm.sh         # Swarm 초기화 & worker join & overlay 네트워크 생성
│   ├── build-push.sh         # 이미지 빌드 & 로컬 레지스트리 push
│   └── deploy.sh             # 전체 배포 자동화
├── stack/
│   └── todo-stack.yml        # docker stack deploy 설정
└── services/
    ├── mysql/
    │   ├── master/
    │   │   ├── Dockerfile
    │   │   ├── my.cnf        # server-id=1, GTID, binlog 설정
    │   │   └── init/
    │   │       ├── 01-schema.sql      # DB/테이블 생성 + 샘플 데이터
    │   │       └── 02-replication.sql # replicator 계정 생성
    │   └── slave/
    │       ├── Dockerfile
    │       ├── my.cnf         # server-id=2, read_only, GTID 설정
    │       └── init/
    │           └── slave-setup.sh    # CHANGE REPLICATION SOURCE + START REPLICA
    ├── todoapi/
    │   ├── Dockerfile         # multi-stage, Go 1.23
    │   ├── go.mod
    │   └── main.go            # CRUD API (master=쓰기, slave=읽기)
    ├── frontend/
    │   ├── Dockerfile         # multi-stage, Node 22 + Nuxt 3
    │   ├── package.json
    │   ├── nuxt.config.ts
    │   └── pages/index.vue    # TODO UI (Vue 3 Composition API)
    ├── nginx/
    │   ├── Dockerfile
    │   └── nginx.conf         # /api/* → todoapi, /* → frontend
    └── haproxy/
        ├── Dockerfile
        └── haproxy.cfg        # roundrobin LB, stats, Docker DNS 리졸버
```

---

## 전제 조건

- Docker Desktop (Mac) 또는 Docker Engine (Linux) 설치
- `docker compose` v2 지원
- 포트 80, 5000, 8404 사용 가능

---

## 실행 방법

### 방법 A: 자동 배포 (권장)

```bash
cd 5th-Server-Dabeen
bash scripts/deploy.sh
```

### 방법 B: 단계별 수동 실행

#### 1단계 — 이미지 빌드 & 레지스트리 Push

```bash
# 로컬 레지스트리가 없으면 먼저 시작
docker run -d -p 5000:5000 --name registry registry:2

bash scripts/build-push.sh
```

#### 2단계 — DinD 클러스터 시작

```bash
docker compose up -d

# 모든 노드 헬스 확인
docker compose ps
```

#### 3단계 — Swarm 초기화 & Worker join

```bash
# manager 컨테이너 안에서 실행
docker exec manager sh /scripts/init-swarm.sh

# 결과 확인
docker exec manager docker node ls
```

#### 4단계 — Overlay 네트워크는 init-swarm.sh 에서 자동 생성됨

```bash
# 확인
docker exec manager docker network ls | grep overlay
```

#### 5단계 — Stack 배포

```bash
docker exec manager docker stack deploy \
  --with-registry-auth \
  -c /stack/todo-stack.yml todo

# 서비스 상태 확인
docker exec manager docker service ls
docker exec manager docker stack ps todo
```

---

## 접속 정보

| URL | 설명 |
|-----|------|
| `http://localhost` | TODO 앱 (Nuxt.js 프론트엔드) |
| `http://localhost/api/todos` | REST API |
| `http://localhost/health` | API 헬스체크 |
| `http://localhost:8404/stats` | HAProxy 통계 (admin / admin123) |
| `http://localhost:5000/v2/_catalog` | Private Registry 이미지 목록 |

---

## 주요 명령어

### Swarm 관리

```bash
# 노드 목록
docker exec manager docker node ls

# 서비스 목록 & 레플리카 상태
docker exec manager docker service ls

# 특정 서비스 로그
docker exec manager docker service logs todo_todoapi -f
docker exec manager docker service logs todo_nginx -f

# 서비스 스케일 조정
docker exec manager docker service scale todo_todoapi=3
docker exec manager docker service scale todo_nginx=3

# 서비스 롤링 업데이트
docker exec manager docker service update \
  --image registry:5000/todoapi:latest \
  --update-parallelism 1 \
  --update-delay 10s \
  todo_todoapi

# 스택 전체 제거
docker exec manager docker stack rm todo
```

### MySQL 복제 확인

```bash
# Slave 복제 상태 확인
docker exec manager docker exec \
  $(docker exec manager docker ps -qf "name=todo_mysql-slave") \
  mysql -u root -proot_pass_123 -e "SHOW REPLICA STATUS\G"

# Master 바이너리 로그 확인
docker exec manager docker exec \
  $(docker exec manager docker ps -qf "name=todo_mysql-master") \
  mysql -u root -proot_pass_123 -e "SHOW MASTER STATUS; SHOW BINARY LOGS;"
```

### 클러스터 정리

```bash
# Stack & 클러스터 완전 제거
docker exec manager docker stack rm todo
docker compose down -v
```

---

## REST API 명세

| Method | Path | 설명 |
|--------|------|------|
| `GET` | `/api/todos` | 전체 목록 조회 (slave에서 읽기) |
| `POST` | `/api/todos` | 할 일 생성 `{ "title": "..." }` |
| `GET` | `/api/todos/:id` | 단건 조회 |
| `PUT` | `/api/todos/:id` | 수정 `{ "title": "...", "done": true }` |
| `DELETE` | `/api/todos/:id` | 삭제 |
| `GET` | `/health` | DB 연결 상태 확인 |

---

## MySQL Master-Slave 복제 흐름

```
todoapi (쓰기)          todoapi (읽기)
     │                       ↑
     ▼                       │
mysql-master            mysql-slave
  [bin log]   ──GTID→  [relay log]
  server-id=1           server-id=2
                        read_only=ON
```

- **GTID 기반** 복제 → binlog position 관리 불필요
- `todoapi`는 INSERT/UPDATE/DELETE → `mysql-master`, SELECT → `mysql-slave`
- slave 연결 실패 시 자동으로 master fallback

---

## 트러블슈팅

```bash
# DinD 노드 Docker 데몬 직접 접근
docker exec -it manager docker info
docker exec -it worker1 docker info

# 서비스 배치 실패 시 상세 확인
docker exec manager docker service ps todo_todoapi --no-trunc

# 네트워크 확인
docker exec manager docker network inspect todo-overlay

# 이미지가 pull 안 될 때
docker exec manager docker pull registry:5000/todoapi:latest
```
