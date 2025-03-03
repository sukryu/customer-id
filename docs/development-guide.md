# TasteSync Development Guide

이 문서는 TasteSync 프로젝트의 개발자를 위한 가이드로, 개발 환경 설정, 코드 작성, 테스트, 기여 프로세스를 설명합니다. `tastesync-customer-id` 서비스를 중심으로 작성되었으며, Go 1.23 기반의 초저지연(1초 내) 고객 식별 기능을 개발하기 위한 실질적인 지침을 제공합니다. DDD, 헥사고날, 이벤트 드리븐 아키텍처를 따르며, HTTPS ↔ gRPC 통신을 구현합니다.

---

## 1. 개발 환경 설정

### 1.1 필수 도구
- **Go 1.23**: https://go.dev/dl/
- **Docker**: 컨테이너 실행 (`docker`, `docker-compose`).
- **gRPC 도구**: 
  - `protoc`: Protocol Buffers 컴파일러 (`brew install protobuf` 또는 OS별 설치).
  - `protoc-gen-go`, `protoc-gen-go-grpc`: Go 플러그인 (`go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`).
- **Kubernetes CLI**: `kubectl` (배포용, https://kubernetes.io/docs/tasks/tools/).
- **IDE**: VS Code 추천 (Go 확장 설치).

### 1.2 로컬 환경 설정
1. **리포지토리 클론**:
   ```bash
   git clone https://github.com/yourusername/tastesync-customer-id.git
   cd tastesync-customer-id
   ```
2. **의존성 설치**:
   ```bash
   make deps
   ```
3. **환경 변수 설정**:
   - `internal/config/config.yaml.example`을 복사하여 `config.yaml` 생성.
   - 예시:
     ```yaml
     server:
       http_port: 3000
       grpc_port: 50051
     redis:
       host: "localhost:6379"
     postgres:
       host: "localhost:5432"
       user: "tastesync"
       password: "secret"
     jwt:
       private_key: "keys/dev/private.pem"
       public_key: "keys/dev/public.pem"
     kafka:
       broker: "localhost:9092"
       topic: "customer-events"
     ```
4. **Docker 서비스 실행** (Redis, PostgreSQL, Kafka):
   ```bash
   docker-compose -f deploy/docker/docker-compose.yml up -d
   ```

---

## 2. 코드 작성 가이드

### 2.1 디렉토리 구조 이해
- **`cmd/server/main.go`**: 서비스 진입점.
- **`internal/`**: 내부 패키지 (도메인, 애플리케이션, 인프라).
- **`pkg/`**: 공통 유틸리티.
- **`proto/`**: gRPC 프로토콜 정의.

### 2.2 코드 스타일
Kubernetes 스타일을 따라 코드의 일관성과 가독성을 유지합니다. 아래 규칙을 준수하세요.

- **명명 규칙**: 
  - **변수**: `lowerCamelCase` (예: `customerID`, `beaconData`).
  - **함수**: `UpperCamelCase` (예: `IdentifyCustomer`, `ParseBeaconPacket`).
  - **패키지**: 소문자 단일 단어, 약어 피하기 (예: `grpc`, `domain`).
  - **상수**: `UPPER_SNAKE_CASE` (예: `MAX_RSSI_VALUE`).
- **포매팅**: 
  - `gofmt`로 자동 포매팅 필수. 탭(indent)은 2칸 사용.
  - 줄 길이는 최대 120자 권장.
- **주석**: 
  - 모든 공용 함수/구조체는 GoDoc 스타일 주석 필수.
  - 예:
    ```go
    // IdentifyCustomer identifies a customer based on beacon data.
    // Returns customer ID, location, and error if any.
    func IdentifyCustomer(uuid, major, minor string, rssi int) (string, string, error) {
        // Implementation
    }
    ```
  - 내부 로직은 간단한 inline 주석 사용 (예: `// Parse RSSI value`).
- **구조체**: 
  - 필드는 `UpperCamelCase`, JSON 태그는 `lowerCamelCase`.
  - 예:
    ```go
    type Customer struct {
        ID       string `json:"customerId"`
        Location string `json:"location"`
    }
    ```
- **에러 처리**: 
  - 에러 변수는 `err`로 통일.
  - `errors.New` 또는 `fmt.Errorf` 사용.
- **린팅**: 
  - `golangci-lint` 실행 (`make lint`)으로 코드 품질 점검.
  - 설정 파일: `.golangci.yml`.

### 2.3 gRPC 프로토콜 작성
- **파일**: `proto/customer_id.proto`.
- **예시**:
  ```proto
  syntax = "proto3";
  package customerid;
  option go_package = "./proto";

  service CustomerID {
    rpc IdentifyCustomer (IdentifyRequest) returns (IdentifyResponse) {}
  }

  message IdentifyRequest {
    string uuid = 1;
    int32 major = 2;
    int32 minor = 3;
    int32 rssi = 4;
  }

  message IdentifyResponse {
    string customer_id = 1;
    string location = 2;
  }
  ```
- **생성**:
  ```bash
  make proto
  ```

### 2.4 도메인 로직 작성
- **경로**: `internal/domain/services/`.
- **예시**: `IdentificationService` 구현.
  ```go
  package services

  type IdentificationService struct{}

  func (s *IdentificationService) Identify(uuid, major, minor string, rssi int) (string, string, error) {
      // 고객 식별 로직
      customerID := "123"
      location := "Table 3"
      return customerID, location, nil
  }
  ```

---

## 3. 테스트 방법

### 3.1 단위 테스트
- **경로**: `tests/unit/`.
- **예시**: `services_test.go`.
  ```go
  package services_test

  import (
      "testing"
      "github.com/yourusername/tastesync-customer-id/internal/domain/services"
  )

  func TestIdentifyCustomer(t *testing.T) {
      svc := services.IdentificationService{}
      id, loc, err := svc.Identify("uuid", "major", "minor", 50)
      if err != nil || id != "123" || loc != "Table 3" {
          t.Errorf("Expected id: 123, loc: Table 3, got %s, %s", id, loc)
      }
  }
  ```
- **실행**:
  ```bash
  make test-unit
  ```

### 3.2 통합 테스트
- **경로**: `tests/integration/`.
- **목적**: gRPC 호출 및 Redis/PostgreSQL 통합 확인.
- **실행**:
  ```bash
  make test-integration
  ```

---

## 4. 개발 프로세스

### 4.1 브랜치 전략
- **main**: 안정 버전.
- **feature/**: 새 기능 개발 (예: `feature/add-auth`).
- **bugfix/**: 버그 수정.

### 4.2 커밋 메시지 규칙
- **형식**: `<type>(<scope>): <description>` (예: `feat(grpc): add IdentifyCustomer endpoint`).
- **타입**: `feat`, `fix`, `docs`, `test`, `chore`.

### 4.3 PR 프로세스
1. 브랜치 생성 및 작업.
2. `make lint`와 `make test-unit` 실행.
3. PR 작성, 최소 1명 리뷰 요청.
4. 리뷰 통과 후 `main`에 병합.

---

## 5. 개발 팁

### 5.1 환경 변수 관리
- **`internal/config/config.go`**에서 `config.yaml` 로드.
- **예시**:
  ```go
  package config

  type Config struct {
      Server struct { HTTPPort, GRPCPort int }
      Redis  struct { Host string }
      JWT    struct { PrivateKey, PublicKey string }
  }

  func Load() (*Config, error) {
      // viper로 config.yaml 로드
  }
  ```

### 5.2 JWT 인증 (RSA)
- **경로**: `internal/auth/jwt.go`.
- **예시**:
  ```go
  package auth

  import (
      "crypto/rsa"
      "github.com/golang-jwt/jwt/v5"
  )

  func GenerateToken(customerID string, privateKey *rsa.PrivateKey) (string, error) {
      token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
          "customer_id": customerID,
      })
      return token.SignedString(privateKey)
  }
  ```

### 5.3 로깅
- **경로**: `internal/infrastructure/logging/`.
- **사용**: `zap` 로거 추천.

---

## 6. 문제 해결

### 6.1 일반적인 오류
- **gRPC 연결 실패**: `GRPC_PORT` 확인, Envoy 실행 여부 점검.
- **Redis 연결 오류**: `docker-compose.yml`에서 Redis 상태 확인.

### 6.2 디버깅
- **로그**: `internal/infrastructure/logging/` 확인.
- **프로파일링**: `go tool pprof` 사용.

---

## 7. 필요한 파일 및 경로
- **`internal/config/config.yaml`**: 환경 변수 설정.
- **`internal/config/config.go`**: 설정 로드.
- **`internal/auth/jwt.go`**: JWT RSA 인증.
- **`proto/customer_id.proto`**: gRPC 정의.
- **`deploy/docker/docker-compose.yml`**: 로컬 서비스 설정.
- **`Makefile`**: 자동화 명령어.

---

## 8. 결론
이 가이드는 `customer-id` 개발을 위한 환경 설정, 코드 작성, 테스트 방법을 제공하며, 팀이 빠르게 온보딩하고 일관된 품질을 유지하도록 돕습니다. 추가 질문은 PR 또는 팀 채널에서 논의하세요.