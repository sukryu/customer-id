# TasteSync Deployment Guide

이 문서는 TasteSync 플랫폼의 배포 프로세스를 설명합니다. `tastesync-customer-id` 서비스를 중심으로, 2027년 v1.0 출시 시 500~1,000 사용자(식당 중심)를 지원하며, 2040년 1,000만 사용자까지 확장 가능하도록 설계되었습니다. Docker 컨테이너와 Kubernetes를 사용하며, Go 1.23 기반의 초저지연(1초 내) 고객 식별 기능을 AWS ECS/EKS에 배포합니다. CI/CD 파이프라인(GitHub Actions)을 통해 빌드, 테스트, 배포를 자동화합니다.

---

## 1. 배포 개요

### 1.1 목표
- **초저지연**: 1초 내 응답을 보장하는 배포 환경.
- **확장성**: 자동 스케일링으로 1,000만 사용자 지원.
- **안정성**: 99.9% 가동률, 장애 복구 5분 내(RTO).
- **자동화**: CI/CD 워크플로우로 배포 효율성 극대화.

### 1.2 배포 환경
- **로컬**: Docker Compose로 개발/테스트 환경.
- **프로덕션**: AWS ECS/EKS, Kubernetes로 관리.
- **의존성**: Redis, PostgreSQL, Kafka (Docker로 실행).

---

## 2. 배포 준비

### 2.1 필수 도구
- **Docker**: 컨테이너 빌드/실행 (`docker`, `docker-compose`).
- **Kubernetes CLI**: `kubectl` (https://kubernetes.io/docs/tasks/tools/).
- **AWS CLI**: AWS ECS/EKS 관리 (https://aws.amazon.com/cli/).
- **Go 1.23**: 서비스 빌드용.

### 2.2 환경 변수 설정
- **파일**: `internal/config/config.yaml`.
- **예시** (프로덕션 환경):
  ```yaml
  server:
    http_port: 3000
    grpc_port: 50051
  redis:
    host: "redis.tastesync.svc.cluster.local:6379"
  postgres:
    host: "postgres.tastesync.svc.cluster.local:5432"
    user: "tastesync"
    password: "${POSTGRES_PASSWORD}"  # Secret 관리
  jwt:
    private_key: "keys/prod/private.pem"
    public_key: "keys/prod/public.pem"
  kafka:
    broker: "kafka.tastesync.svc.cluster.local:9092"
    topic: "customer-events"
  ```
- **Secret 관리**: Kubernetes Secrets로 민감 데이터 주입 (예: `${POSTGRES_PASSWORD}`).

### 2.3 JWT 키 준비
- **경로**: `internal/config/keys/prod/`.
- **생성**:
  ```bash
  openssl genrsa -out internal/config/keys/prod/private.pem 2048
  openssl rsa -in internal/config/keys/prod/private.pem -pubout -out internal/config/keys/prod/public.pem
  ```

---

## 3. 로컬 배포

### 3.1 Docker Compose 설정
- **파일**: `deploy/docker/docker-compose.yml`.
- **예시**:
  ```yaml
  version: '3.8'
  services:
    customer-id:
      build: .
      ports:
        - "3000:3000"  # HTTP
        - "50051:50051"  # gRPC
      volumes:
        - ./internal/config:/app/internal/config
      environment:
        - POSTGRES_PASSWORD=secret
    redis:
      image: redis:7
      ports: [ "6379:6379" ]
    postgres:
      image: postgres:15
      environment:
        - POSTGRES_USER=tastesync
        - POSTGRES_PASSWORD=secret
    kafka:
      image: confluentinc/cp-kafka:latest
      ports: [ "9092:9092" ]
  ```

### 3.2 로컬 실행
1. **빌드 및 실행**:
   ```bash
   make docker-build
   make docker-up
   ```
2. **확인**:
   - HTTP: `curl http://localhost:3000/health`.
   - gRPC: `grpcurl -plaintext localhost:50051 customerid.CustomerID/IdentifyCustomer`.

---

## 4. 프로덕션 배포

### 4.1 Docker 이미지 빌드
- **파일**: `deploy/docker/Dockerfile`.
- **예시**:
  ```dockerfile
  FROM golang:1.23 AS builder
  WORKDIR /app
  COPY . .
  RUN make build

  FROM alpine:latest
  WORKDIR /app
  COPY --from=builder /app/customer-id .
  COPY internal/config/ internal/config/
  EXPOSE 3000 50051
  CMD ["./customer-id"]
  ```
- **빌드**:
  ```bash
  docker build -t yourusername/customer-id:latest -f deploy/docker/Dockerfile .
  docker push yourusername/customer-id:latest
  ```

### 4.2 Kubernetes 배포
- **파일**: `deploy/k8s/customer-id.yaml`.
- **예시**:
  ```yaml
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: customer-id
  spec:
    replicas: 3
    selector:
      matchLabels:
        app: customer-id
    template:
      metadata:
        labels: { app: customer-id }
      spec:
        containers:
        - name: customer-id
          image: yourusername/customer-id:latest
          ports:
          - containerPort: 3000
          - containerPort: 50051
          env:
          - name: POSTGRES_PASSWORD
            valueFrom:
              secretKeyRef:
                name: customer-id-secrets
                key: postgres-password
          volumeMounts:
          - name: config-volume
            mountPath: /app/internal/config
        volumes:
        - name: config-volume
          configMap:
            name: customer-id-config
  ---
  apiVersion: v1
  kind: Service
  metadata:
    name: customer-id
  spec:
    ports:
    - port: 3000
      targetPort: 3000
      name: http
    - port: 50051
      targetPort: 50051
      name: grpc
    selector: { app: customer-id }
  ```
- **적용**:
  ```bash
  make deploy
  ```

### 4.3 Envoy 설정
- **파일**: `tastesync-backend-common/envoy.yaml`.
- **예시** (간략):
  ```yaml
  static_resources:
    listeners:
    - address: { socket_address: { address: "0.0.0.0", port_value: 443 } }
      filter_chains:
      - filters:
        - name: envoy.filters.network.http_connection_manager
          config:
            route_config:
              virtual_hosts:
              - name: customer-id
                domains: ["api.tastesync.com"]
                routes:
                - match: { prefix: "/customer-id" }
                  route: { cluster: "customer-id-grpc" }
    clusters:
    - name: customer-id-grpc
      connect_timeout: 0.25s
      type: LOGICAL_DNS
      lb_policy: ROUND_ROBIN
      http2_protocol_options: {}
      load_assignment:
        cluster_name: customer-id-grpc
        endpoints:
        - lb_endpoints:
          - endpoint: { address: { socket_address: { address: "customer-id", port_value: 50051 } } }
  ```

---

## 5. CI/CD 파이프라인

### 5.1 빌드 및 테스트
- **파일**: `.github/workflows/build.yml`.
- **내용**: [기존 예시 유지].

### 5.2 배포
- **파일**: `.github/workflows/deploy.yml`.
- **내용**: [기존 예시 유지, AWS 자격 증명 추가 필요].

### 5.3 실행
- **트리거**: `main` 브랜치 푸시 시 자동 실행.
- **환경 변수**: GitHub Secrets에 `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `KUBECONFIG` 설정.

---

## 6. 모니터링 및 유지보수

### 6.1 모니터링 설정
- **파일**: `deploy/k8s/prometheus-config.yaml`.
- **예시** (간략):
  ```yaml
  apiVersion: monitoring.coreos.com/v1
  kind: ServiceMonitor
  metadata:
    name: customer-id-monitor
  spec:
    selector: { matchLabels: { app: customer-id } }
    endpoints:
    - port: http
      path: /metrics
  ```

### 6.2 로그 관리
- **경로**: S3 버킷 (`s3://tastesync-logs/customer-id/`).
- **설정**: `internal/infrastructure/logging/`에서 Fluentd로 전송.

---

## 7. 배포 팁

### 7.1 스케일링
- **수평 확장**: `replicas` 조정 (`kubectl scale deployment customer-id --replicas=5`).
- **자동 확장**: HPA 설정 추천.

### 7.2 롤백
- **명령어**: `kubectl rollout undo deployment/customer-id`.

### 7.3 문제 해결
- **로그 확인**: `kubectl logs -l app=customer-id`.
- **서비스 상태**: `kubectl get pods -l app=customer-id`.

---

## 8. 필요한 파일 및 경로
- **`deploy/docker/Dockerfile`**: Docker 이미지 빌드.
- **`deploy/docker/docker-compose.yml`**: 로컬 환경 설정.
- **`deploy/k8s/customer-id.yaml`**: Kubernetes 배포.
- **`deploy/k8s/prometheus-config.yaml`**: 모니터링 설정.
- **`tastesync-backend-common/envoy.yaml`**: Envoy 설정.
- **`.github/workflows/build.yml`, `deploy.yml`**: CI/CD 워크플로우.

---

## 9. 결론
`tastesync-customer-id`는 Docker와 Kubernetes를 통해 초저지연과 확장성을 구현하며, CI/CD로 배포를 자동화합니다. 이 가이드를 따라 로컬에서 프로덕션까지 안정적으로 배포하세요.