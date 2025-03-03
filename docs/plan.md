# TasteSync Development Plan

이 문서는 TasteSync 플랫폼의 개발 계획을 정의합니다. `tastesync-customer-id` 서비스를 중심으로, 2027년 v1.0 출시를 목표로 하며, 초저지연(1초 내), 확장성(1,000만 사용자), 소상공인 배려를 구현합니다. 개발 일정은 포함하지 않고, 우선순위에 따라 개발해야 할 항목을 순서대로 나열하며, 각 단계의 이유와 세부 작업을 상세히 설명합니다. 이 계획은 DDD, 헥사고날, 이벤트 드리븐 아키텍처를 기반으로 진행됩니다.

---

## 1. 개발 계획 개요

### 1.1 목표
- **초저지연**: 비콘/QRS 감지 후 1초 내 고객 식별.
- **확장성**: 초기 500~1,000 사용자에서 1,000만 사용자까지 지원.
- **소상공인 배려**: 정확한 고객 식별로 추천/운영 분석 제공.
- **안정성**: 모니터링/로깅으로 품질 유지.

### 1.2 우선순위 원칙
- **핵심 기능 먼저**: 고객 식별 로직 구축.
- **의존성 최소화**: 독립적 개발 가능한 모듈 우선.
- **확장성 대비**: 기반 구조(설정, 인증, 이벤트) 선행.
- **품질 보장**: 테스트, 모니터링 병행.

---

## 2. 개발 우선순위 및 세부 작업

### 2.1 설정 및 환경 구축
- **우선순위**: 1
- **이유**: 모든 개발의 기반이 되는 환경과 설정을 먼저 마련해야 이후 작업이 원활히 진행됨.
- **세부 작업**:
  1. **프로젝트 구조 초기화**:
     - 디렉토리 생성: `cmd/`, `internal/`, `pkg/`, `proto/`, `deploy/`, `tests/`, `docs/`.
     - Git 초기화 및 `go.mod` 설정 (`go mod init github.com/yourusername/tastesync-customer-id`).
  2. **환경 변수 설정**:
     - `internal/config/config.yaml` 정의 (HTTP/gRPC 포트, Redis/PostgreSQL 연결 등).
     - `internal/config/config.go` 작성 (`viper`로 로드 로직 구현).
  3. **JWT 키 생성**:
     - RSA 키 쌍 생성 (`make gen-keys ENV=dev`).
     - 경로: `internal/config/keys/dev/`.
  4. **로컬 환경 구성**:
     - `deploy/docker/docker-compose.yml` 작성 (Redis, PostgreSQL, Kafka 포함).
     - `make docker-up`으로 로컬 서비스 실행 확인.
- **출력물**: 실행 가능한 기본 프로젝트 환경.

### 2.2 도메인 모델 설계 및 구현
- **우선순위**: 2
- **이유**: 고객 식별의 핵심 비즈니스 로직을 정의하며, 이후 모든 기능의 기반이 됨.
- **세부 작업**:
  1. **엔티티 정의**:
     - `internal/domain/entities/customer.go`: `Customer` 구조체.
     - `internal/domain/entities/beacon.go`: `Beacon` 구조체.
  2. **애그리게이트 정의**:
     - `internal/domain/aggregates/customer_identity.go`: `CustomerIdentity` 구조체 및 규칙 (`Confidence` 검증).
  3. **값 객체 정의**:
     - `internal/domain/entities/beacon_data.go`: `BeaconData`.
     - `internal/domain/entities/location.go`: `Location`.
  4. **도메인 서비스 구현**:
     - `internal/domain/services/identification.go`: `IdentifyCustomer` 로직 (비콘 데이터 → 고객 식별).
- **출력물**: 도메인 모델 초기 구현 (`Customer`, `Beacon`, `CustomerIdentity`).

### 2.3 데이터 저장소 및 캐싱 구현
- **우선순위**: 3
- **이유**: 도메인 모델을 영구 저장하고 초저지연 조회를 위해 캐싱이 필요하며, 핵심 기능의 데이터 기반 구축.
- **세부 작업**:
  1. **PostgreSQL 스키마 생성**:
     - `internal/infrastructure/db/schema.sql` 작성 (`customers`, `beacons`, `customer_identities` 테이블).
     - 파티션 및 인덱스 적용 (예: `idx_detected_at`, `idx_customer_id`).
     - `scripts/migrate.sh`로 초기화.
  2. **Redis 캐싱 구현**:
     - `internal/infrastructure/redis/cache.go`: `SetCustomerIdentity`, `GetCustomerIdentity` 함수.
     - 키 구조 정의 (예: `customer:<customer_id>`).
  3. **저장소 어댑터 구현**:
     - `internal/infrastructure/db/postgres.go`: PostgreSQL 연결 및 CRUD (`pgx` 라이브러리 사용).
     - 헥사고날 포트 인터페이스 정의 (`internal/application/ports/storage.go`).
- **출력물**: 데이터 저장 및 캐싱 초기 구현.

### 2.4 gRPC 서비스 구현
- **우선순위**: 4
- **이유**: 외부 요청(HTTPS → gRPC)과 내부 통신의 핵심 인터페이스이며, 도메인 로직을 클라이언트에 노출.
- **세부 작업**:
  1. **gRPC 프로토콜 정의**:
     - `proto/customer_id.proto` 작성 (`IdentifyCustomer` 메서드).
     - `make proto`로 Go 코드 생성.
  2. **gRPC 서버 구현**:
     - `internal/infrastructure/grpc/server.go`: Fiber와 gRPC 서버 통합.
     - `IdentifyCustomer` 핸들러 추가 (도메인 서비스 호출).
  3. **인증 추가**:
     - `internal/auth/jwt.go`: JWT RSA 검증 미들웨어 구현.
     - gRPC 요청에 토큰 검증 로직 통합.
- **출력물**: `IdentifyCustomer` gRPC 엔드포인트 초기 구현.

### 2.5 이벤트 발행 구현
- **우선순위**: 5
- **이유**: 고객 식별 후 비동기 처리를 위해 이벤트 드리븐 아키텍처의 핵심인 이벤트 발행이 필요.
- **세부 작업**:
  1. **이벤트 정의**:
     - `internal/domain/events/customer_identified.go`: `CustomerIdentified` 구조체.
  2. **Kafka 퍼블리셔 구현**:
     - `internal/infrastructure/kafka/publisher.go`: `PublishCustomerIdentified` 함수.
     - `customer-events` 토픽으로 발행 로직 추가.
  3. **도메인 서비스 연계**:
     - `IdentificationService`에서 이벤트 발행 호출 추가.
- **출력물**: `CustomerIdentified` 이벤트 발행 기능.

### 2.6 테스트 작성
- **우선순위**: 6
- **이유**: 핵심 기능의 안정성을 보장하고, 이후 확장 시 리그레션 방지.
- **세부 작업**:
  1. **단위 테스트**:
     - `tests/unit/services_test.go`: `IdentificationService` 테스트.
     - `make test-unit`으로 실행 확인.
  2. **통합 테스트**:
     - `tests/integration/grpc_test.go`: gRPC 엔드포인트 테스트.
     - Redis, PostgreSQL 모킹 추가.
  3. **성능 테스트**:
     - `tests/performance/load_test.go`: 1,000 요청 동시 처리 확인.
- **출력물**: 테스트 코드 초기 세트.

### 2.7 모니터링 및 로깅 구현
- **우선순위**: 7
- **이유**: 서비스 품질 유지와 문제 진단을 위해 필수이며, 초기부터 모니터링 기반 구축 필요.
- **세부 작업**:
  1. **로깅 설정**:
     - `internal/infrastructure/logging/logger.go`: Zap 로거 초기화.
     - `IdentifyCustomer` 호출 시 구조화 로그 추가.
  2. **메트릭 설정**:
     - `internal/infrastructure/grpc/server.go`: Prometheus 메트릭 추가 (`http_request_duration_seconds` 등).
     - `/metrics` 엔드포인트 활성화.
  3. **Fluentd 통합**:
     - `deploy/k8s/fluentd-config.yaml` 작성 (S3 로그 전송).
- **출력물**: 모니터링/로깅 초기 구현.

### 2.8 배포 환경 설정
- **우선순위**: 8
- **이유**: 로컬 및 프로덕션 배포를 통해 실제 동작 확인 필요.
- **세부 작업**:
  1. **Docker 설정**:
     - `deploy/docker/Dockerfile` 작성 및 이미지 빌드 테스트.
     - `deploy/docker/docker-compose.yml`으로 통합 환경 실행.
  2. **Kubernetes 설정**:
     - `deploy/k8s/customer-id.yaml` 작성 (Deployment, Service 정의).
     - `make deploy`로 로컬 클러스터 테스트.
  3. **CI/CD 설정**:
     - `.github/workflows/build.yml`, `deploy.yml` 작성 및 GitHub Actions 테스트.
- **출력물**: 배포 가능한 컨테이너 및 Kubernetes 매니페스트.

### 2.9 추가 최적화 및 확장
- **우선순위**: 9
- **이유**: 초기 기능을 안정화한 후 성능과 확장성을 강화.
- **세부 작업**:
  1. **데이터베이스 최적화**:
     - PostgreSQL 파티션 추가 (`customer_identities_2026` 등).
     - DynamoDB 이벤트 로그 테이블 초기화.
  2. **NATS 통합**:
     - `internal/infrastructure/nats/publisher.go`로 실시간 알림 추가.
  3. **Envoy 설정**:
     - `tastesync-backend-common/envoy.yaml`으로 HTTPS ↔ gRPC 변환 구현.
- **출력물**: 최적화된 데이터 저장소 및 추가 메시지 큐.

---

## 3. 개발 순서 이유

1. **설정 및 환경 구축**: 모든 작업의 기반, 초기 설정 없이는 개발 불가.
2. **도메인 모델**: 고객 식별의 핵심 로직, 이후 의존성 최소화하며 구현 가능.
3. **데이터 저장소**: 도메인 데이터를 저장/조회하는 필수 인프라.
4. **gRPC 서비스**: 외부 통신 인터페이스, 클라이언트 연동의 첫 단계.
5. **이벤트 발행**: 비동기 처리로 추천/알림 연결, 핵심 기능 완성.
6. **테스트**: 기능 안정성 보장, 초기 버그 방지.
7. **모니터링/로깅**: 품질 유지와 문제 진단, 운영 준비.
8. **배포 환경**: 실제 환경에서 동작 확인, 프로덕션 준비.
9. **최적화/확장**: 초기 기능 완성 후 성능 개선 및 추가 기능.

---

## 4. 결론
이 계획은 `customer-id`의 개발을 초저지연과 확장성을 고려하여 단계별로 진행하며, 설정부터 최적화까지 체계적으로 접근합니다. 각 단계는 독립적이면서도 후속 작업의 기반을 제공하므로, 우선순위를 따라 개발을 시작하세요.