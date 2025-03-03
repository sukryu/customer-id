# tastesync-customer-id README

`tastesync-customer-id`는 TasteSync 플랫폼의 고객 식별 서비스로, 비콘 및 QRS 기반으로 고객을 실시간(1초 내) 식별하고 위치(입구/테이블)를 파악합니다。2027년 v1.0 출시 시 500~1,000 사용자(식당 중심)를 지원하며, 2040년 1,000만 사용자까지 확장 가능합니다。Go 1.23과 gRPC를 기반으로 초저지연을 구현하며, 외부 통신은 HTTPS로, 내부 통신은 gRPC로 처리합니다。DDD, 헥사고날, 이벤트 드리븐 아키텍처를 결합하여 유연성과 유지보수성을 보장합니다。

---

## 1. 서비스 개요

### 1.1 목표
- **초저지연**: 비콘/QRS 감지 후 1초 내 고객 식별 및 위치 반환.
- **확장성**: 초기 500 사용자에서 1,000만 사용자까지 지원.
- **소상공인 배려**: 정확한 고객 식별로 추천/운영 분석 제공.
- **통합성**: HTTPS(외부) ↔ gRPC(내부) 변환으로 서버와 통신.

### 1.2 역할
- 비콘 데이터(UUID, Major, Minor, RSSI) 파싱 후 고객 식별.
- 위치 기반 이벤트 발생(입구 입장, 테이블 감지).
- 다른 서비스(`recommendation`, `notification`)와 gRPC로 연동.

### 1.3 데이터 흐름
```
[외부 HTTPS] → [Envoy API Gateway] → [customer-id gRPC]
   ↓                ↓ (gRPC → HTTPS)         ↓
[Redis Cache]     [외부 응답]           [Kafka Event]
   ↓                                     ↓
[PostgreSQL]                        [추천/알림 서비스]
```

- **외부 → 서비스**: HTTPS 요청 → Envoy에서 gRPC로 변환 → `customer-id`.
- **서비스 → 외부**: gRPC 응답 → Envoy에서 HTTPS로 변환 → 클라이언트.
- **이벤트**: 식별 후 Kafka로 이벤트 발행 → 비동기 처리.

---

## 2. 아키텍처 설계

### 2.1 DDD (Domain-Driven Design)
- **도메인**: `Customer`, `Beacon`.
- **Bounded Context**: 고객 식별.
- **엔티티**: `Customer` (ID, 위치), `Beacon` (UUID, Major, Minor).
- **애그리게이트**: `CustomerIdentity`.
- **도메인 이벤트**: `CustomerIdentified`.

### 2.2 헥사고날 아키텍처 (Ports & Adapters)
- **코어**: 도메인 로직 (고객 식별, 위치 계산).
- **포트**: `IdentifyCustomerPort`, `EventPublisherPort`.
- **어댑터**: gRPC 입력, Kafka 출력, Redis/PostgreSQL 저장.

### 2.3 이벤트 드리븐 아키텍처
- **이벤트**: `CustomerIdentified` → Kafka 발행.
- **비동기**: 식별 후 즉시 응답, 후속 작업은 이벤트 처리.

---

## 3. 디렉토리 구조

```
tastesync-customer-id/
├── cmd/                  # 실행 가능한 애플리케이션 진입점
│   └── server/
│       └── main.go       # 서비스 진입점
├── internal/             # 내부 패키지 (K8s 스타일)
│   ├── config/           # 환경 변수 관리
│   │   ├── config.go     # 환경 변수 로드
│   │   ├── config.yaml   # 설정 파일
│   │   └── keys/         # JWT RSA 키
│   │       ├── dev/      # 개발용 .pem 파일
│   │       │   ├── private.pem
│   │       │   └── public.pem
│   │       └── prod/     # 운영용 .pem 파일
│   │           ├── private.pem
│   │           └── public.pem
│   ├── domain/           # DDD 도메인 로직
│   │   ├── entities/     # Customer, Beacon
│   │   ├── aggregates/   # CustomerIdentity
│   │   ├── events/       # CustomerIdentified
│   │   └── services/     # IdentificationService
│   ├── application/      # 애플리케이션 계층
│   │   ├── ports/        # IdentifyCustomerPort 등
│   │   └── use_cases/    # IdentifyCustomerUseCase
│   ├── infrastructure/   # 헥사고날 어댑터
│   │   ├── grpc/         # gRPC 서버/클라이언트
│   │   ├── kafka/        # Kafka 퍼블리셔
│   │   ├── redis/        # Redis 클라이언트
│   │   ├── db/           # PostgreSQL 클라이언트
│   │   └── logging/      # 로깅
│   └── auth/             # JWT 인증 (RSA)
├── pkg/                  # 재사용 가능 패키지 (K8s 스타일)
│   └── util/             # 공통 유틸리티
├── proto/                # gRPC 프로토콜 정의
│   └── customer_id.proto # IdentifyCustomer 정의
├── deploy/               # 배포 환경
│   ├── docker/           # Docker 설정
│   │   └── Dockerfile    # 컨테이너 설정
│   └── k8s/              # Kubernetes 설정
│       └── customer-id.yaml # 배포 매니페스트
├── tests/                # 테스트
│   ├── unit/
│   ├── integration/
│   └── mocks/
├── .github/              # CI/CD 워크플로우
│   └── workflows/
│       ├── build.yml     # 빌드/테스트 자동화
│       └── deploy.yml    # 배포 자동화
├── Makefile              # 자동화 스크립트
├── go.mod                # Go 의존성
├── .golangci.yml         # 코드 품질 설정
└── README.md             # 이 문서
```

---

## 4. 기술 스택
- **언어**: **Go 1.23**.
- **프레임워크**: **Fiber** (HTTP/gRPC 처리).
- **통신**: 
  - **gRPC**: 내부 통신.
  - **HTTPS**: 외부 통신 (Envoy 경유).
- **메시지 큐**: Kafka (이벤트 발행).
- **데이터베이스**: Redis (캐싱), PostgreSQL (영구 저장).
- **인증**: JWT (RSA 알고리즘).

---

## 5. 정책 및 보안

### 5.1 정책
- **코드 품질**: 
  - Golangci-lint 적용, PR 시 80% 커버리지 요구.
  - 경로: `.golangci.yml`.
- **버전 관리**: SemVer 준수 (예: `v1.23.0`).
- **의존성**: 
  - Go Modules 사용, 주 1회 취약점 스캔.
  - 경로: `go.mod`.
- **CI/CD**: 
  - GitHub Actions로 빌드/배포 자동화.
  - 경로: `.github/workflows/build.yml`, `deploy.yml`.

### 5.2 보안 수준
- **통신**: 
  - gRPC는 TLS 1.3 암호화.
  - HTTPS는 Envoy에서 처리.
  - 경로: `internal/infrastructure/grpc/tls_config.go`.
- **인증**: 
  - JWT 토큰 RSA 알고리즘 사용(2048비트 키).
  - .pem 파일: `internal/config/keys/dev/`, `prod/`.
  - 경로: `internal/auth/jwt.go`.
- **데이터**: 
  - 민감 데이터 AES-256 암호화.
  - Redis TTL 1시간.
- **취약점 관리**: 
  - OWASP ZAP 주 1회 점검.
  - 경로: `tests/security_test.sh`.

---

## 6. 설정 및 실행

### 6.1 필수 조건
- Go 1.23 설치.
- Docker, Docker Compose.
- gRPC 도구 (protoc, protoc-gen-go).

### 6.2 로컬 실행
1. **클론**:
   ```bash
   git clone https://github.com/yourusername/tastesync-customer-id.git
   cd tastesync-customer-id
   ```
2. **의존성 설치**:
   ```bash
   make deps
   ```
3. **환경 변수**:
   ```bash
   cp internal/config/config.yaml.example internal/config/config.yaml
   # config.yaml 편집
   ```
4. **gRPC 프로토콜 생성**:
   ```bash
   make proto
   ```
5. **실행**:
   ```bash
   make run
   ```

### 6.3 배포
- **파일**: `deploy/docker/Dockerfile`, `deploy/k8s/customer-id.yaml`.
- **명령어**:
   ```bash
   make deploy
   ```

---

## 7. 테스트 전략
- **단위 테스트**: Go Test로 도메인/유스케이스 검증.
- **통합 테스트**: gRPC 클라이언트로 초저지연 확인.
- **실행**:
   ```bash
   make test-unit
   make test-integration
   ```

---

## 8. 데이터 흐름 예시
- **고객 식별**:
  1. HTTPS 요청 → Envoy → `IdentifyCustomer` gRPC.
  2. `customer-id` 처리 → Redis/PostgreSQL 조회 → `CustomerIdentified` 이벤트 발행(Kafka).
  3. gRPC 응답 → Envoy → HTTPS 반환(<1초).

---

## 9. 필요한 파일 및 경로
- **`proto/customer_id.proto`**: gRPC 서비스 정의.
- **`internal/config/config.yaml`**: 환경 변수 설정.
- **`internal/config/config.go`**: 설정 로드 로직.
- **`internal/config/keys/dev/private.pem`, `public.pem`**: 개발용 RSA 키.
- **`internal/config/keys/prod/private.pem`, `public.pem`**: 운영용 RSA 키.
- **`internal/infrastructure/grpc/tls_config.go`**: TLS 설정.
- **`internal/auth/jwt.go`**: JWT RSA 인증.
- **`tests/security_test.sh`**: 보안 테스트 스크립트.
- **`deploy/docker/Dockerfile`**: Docker 설정.
- **`deploy/k8s/customer-id.yaml`**: Kubernetes 매니페스트.
- **`.github/workflows/build.yml`**: 빌드/테스트 워크플로우.
- **`.github/workflows/deploy.yml`**: 배포 워크플로우.
- **`Makefile`**: 자동화 스크립트.
- **`.golangci.yml`**: 코드 품질 설정.

---

## 10. Makefile 예시
```
.PHONY: deps proto build run test-unit test-integration deploy

deps:
	go mod tidy

proto:
	protoc --go_out=. --go-grpc_out=. proto/customer_id.proto

build:
	go build -o customer-id cmd/server/main.go

run:
	./customer-id

test-unit:
	go test ./tests/unit/...

test-integration:
	go test ./tests/integration/...

deploy:
	kubectl apply -f deploy/k8s/customer-id.yaml
```

---

## 11. CI/CD 워크플로우 예시

### .github/workflows/build.yml
```yaml
name: Build and Test

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Set up Go
      uses: actions/setup-go@v3
      with: { go-version: 1.23 }
    - name: Install Dependencies
      run: make deps
    - name: Generate Protobuf
      run: make proto
    - name: Build
      run: make build
    - name: Test
      run: make test-unit && make test-integration
```

### .github/workflows/deploy.yml
```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Set up Docker
      uses: docker/setup-buildx-action@v2
    - name: Build and Push Docker Image
      run: |
        docker build -t yourusername/customer-id:latest -f deploy/docker/Dockerfile .
        docker push yourusername/customer-id:latest
    - name: Deploy to Kubernetes
      run: make deploy
      env:
        KUBECONFIG: ${{ secrets.KUBECONFIG }}
```

---

## 12. 결론
`tastesync-customer-id`는 Go 1.23(Fiber)와 gRPC로 초저지연을 구현하며, HTTPS 외부 통신과 DDD/헥사고날/이벤트 드리븐 아키텍처를 결합합니다。RSA 기반 JWT 인증과 자동화된 배포를 통해 안정성과 확장성을 보장합니다。