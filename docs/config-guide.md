# TasteSync Configuration Guide

이 문서는 TasteSync 플랫폼의 환경 변수 및 설정 파일 관리 방법을 안내합니다. `tastesync-customer-id` 서비스를 중심으로, 2027년 v1.0 출시 시 500~1,000 사용자(식당 중심)를 지원하며, 초기 설정부터 2040년 1,000만 사용자 규모까지 확장 가능하도록 설계되었습니다. 설정은 `internal/config/config.yaml` 파일과 `config.go` 로직으로 통합 관리되며, 개발, 테스트, 운영 환경별로 구분됩니다.

---

## 1. 설정 개요

### 1.1 목표
- **일관성**: 모든 환경에서 설정 구조 통일.
- **보안성**: 민감 데이터(JWT 키, DB 비밀번호) 안전 관리.
- **유연성**: 환경별 설정 분리 및 동적 로드 지원.
- **간편성**: 설정 변경 시 코드 수정 최소화.

### 1.2 관리 방식
- **파일**: `internal/config/config.yaml`로 기본값 정의.
- **코드**: `internal/config/config.go`로 동적 로드 및 검증.
- **환경 변수**: 민감 데이터는 Kubernetes Secrets 또는 OS 환경 변수로 주입.

---

## 2. 설정 파일 구조

### 2.1 config.yaml
- **경로**: `internal/config/config.yaml`.
- **구조**: YAML 형식으로 계층적 설정 정의.
- **예시** (개발 환경):
  ```yaml
  server:
    http_port: 3000          # HTTP 서버 포트
    grpc_port: 50051         # gRPC 서버 포트
    timeout: 5s              # 요청 타임아웃
  redis:
    host: "localhost:6379"   # Redis 서버 주소
    password: ""             # Redis 비밀번호 (빈 문자열 가능)
    db: 0                    # Redis 데이터베이스 번호
  postgres:
    host: "localhost:5432"   # PostgreSQL 주소
    user: "tastesync"        # DB 사용자
    password: "secret"       # DB 비밀번호
    database: "tastesync"    # DB 이름
  jwt:
    private_key: "keys/dev/private.pem"  # JWT RSA 개인 키 경로
    public_key: "keys/dev/public.pem"    # JWT RSA 공개 키 경로
    expiration: 3600         # 토큰 만료 시간 (초, 1시간)
  kafka:
    broker: "localhost:9092" # Kafka 브로커 주소
    topic: "customer-events" # 이벤트 발행 토픽
    partition: 3             # 파티션 수
  logging:
    level: "info"            # 로그 레벨 (debug, info, warn, error)
    output: "stdout"         # 로그 출력 (stdout, file)
  ```

### 2.2 환경별 설정
- **개발 (Dev)**: 로컬 테스트용, 기본값 사용.
- **테스트 (Test)**: CI/CD 파이프라인용, 모의 데이터 포함.
- **운영 (Prod)**: 프로덕션용, 민감 데이터 Secret 주입.
- **구분**: 
  - 파일명: `config-<env>.yaml` (예: `config-prod.yaml`).
  - 환경 변수: `TASTESYNC_ENV`로 환경 지정 (예: `export TASTESYNC_ENV=prod`).

---

## 3. 설정 로드 로직

### 3.1 config.go
- **경로**: `internal/config/config.go`.
- **구현**: `viper` 라이브러리로 YAML 파일 로드 및 환경 변수 오버라이드.
- **예시**:
  ```go
  package config

  import (
      "fmt"
      "time"
      "github.com/spf13/viper"
  )

  type Config struct {
      Server   ServerConfig   `mapstructure:"server"`
      Redis    RedisConfig    `mapstructure:"redis"`
      Postgres PostgresConfig `mapstructure:"postgres"`
      JWT      JWTConfig      `mapstructure:"jwt"`
      Kafka    KafkaConfig    `mapstructure:"kafka"`
      Logging  LoggingConfig  `mapstructure:"logging"`
  }

  type ServerConfig struct {
      HTTPPort int           `mapstructure:"http_port"`
      GRPCPort int           `mapstructure:"grpc_port"`
      Timeout  time.Duration `mapstructure:"timeout"`
  }

  type RedisConfig struct {
      Host     string `mapstructure:"host"`
      Password string `mapstructure:"password"`
      DB       int    `mapstructure:"db"`
  }

  // PostgresConfig, JWTConfig, KafkaConfig, LoggingConfig 생략

  func Load() (*Config, error) {
      v := viper.New()
      env := viper.GetString("TASTESYNC_ENV")
      if env == "" {
          env = "dev"  // 기본값
      }
      v.SetConfigName(fmt.Sprintf("config-%s", env))
      v.SetConfigType("yaml")
      v.AddConfigPath("internal/config/")
      v.AutomaticEnv()  // 환경 변수 오버라이드

      if err := v.ReadInConfig(); err != nil {
          return nil, fmt.Errorf("failed to read config: %w", err)
      }

      var cfg Config
      if err := v.Unmarshal(&cfg); err != nil {
          return nil, fmt.Errorf("failed to unmarshal config: %w", err)
      }
      return &cfg, nil
  }
  ```
- **사용**:
  ```go
  cfg, err := config.Load()
  if err != nil {
      log.Fatal(err)
  }
  fmt.Println(cfg.Server.HTTPPort)  // 3000 출력
  ```

### 3.2 환경 변수 오버라이드
- **형식**: 대문자, 밑줄로 구분 (예: `POSTGRES_PASSWORD`).
- **우선순위**: 환경 변수 > `config.yaml`.

---

## 4. 환경별 설정 관리

### 4.1 개발 환경 (Dev)
- **파일**: `internal/config/config-dev.yaml`.
- **특징**: 로컬 Docker 서비스에 연결, 디버깅 용이.
- **실행**:
  ```bash
  export TASTESYNC_ENV=dev
  make run
  ```

### 4.2 테스트 환경 (Test)
- **파일**: `internal/config/config-test.yaml`.
- **특징**: CI/CD에서 모의 데이터 사용, 외부 의존성 최소화.
- **예시**:
  ```yaml
  redis:
    host: "mock-redis:6379"
  postgres:
    host: "mock-postgres:5432"
  ```

### 4.3 운영 환경 (Prod)
- **파일**: `internal/config/config-prod.yaml`.
- **특징**: Kubernetes Secrets로 민감 데이터 주입.
- **Secret 설정**:
  ```yaml
  apiVersion: v1
  kind: Secret
  metadata:
    name: customer-id-secrets
  type: Opaque
  data:
    postgres-password: c2VjcmV0  # base64 인코딩된 "secret"
  ```
- **적용**: `deploy/k8s/customer-id.yaml`의 `env` 섹션 참조.

---

## 5. JWT 키 관리

### 5.1 키 생성
- **경로**: `internal/config/keys/<env>/`.
- **명령어**:
  ```bash
  make gen-keys ENV=dev  # 개발용
  make gen-keys ENV=prod # 운영용
  ```
- **Makefile 예시**:
  ```makefile
  gen-keys:
    mkdir -p internal/config/keys/$(ENV)
    openssl genrsa -out internal/config/keys/$(ENV)/private.pem 2048
    openssl rsa -in internal/config/keys/$(ENV)/private.pem -pubout -out internal/config/keys/$(ENV)/public.pem
  ```

### 5.2 키 로드
- **`config.yaml`에서 지정**:
  ```yaml
  jwt:
    private_key: "keys/prod/private.pem"
    public_key: "keys/prod/public.pem"
  ```
- **구현**: `internal/auth/jwt.go`에서 키 로드 및 검증.

---

## 6. 설정 검증 및 디버깅

### 6.1 검증
- **필수 필드**: `server.http_port`, `redis.host`, `postgres.host` 등 확인.
- **구현**: `config.go`에서 로드 후 유효성 검사 추가.
  ```go
  if cfg.Server.HTTPPort == 0 {
      return nil, fmt.Errorf("http_port is required")
  }
  ```

### 6.2 디버깅
- **로그**: `internal/infrastructure/logging/`에서 설정값 출력.
- **명령어**: `make run` 시 디버그 모드로 실행 (`logging.level: debug`).

---

## 7. 필요한 파일 및 경로
- **`internal/config/config.yaml`**: 기본 설정 파일.
- **`internal/config/config-<env>.yaml`**: 환경별 설정 (dev, test, prod).
- **`internal/config/config.go`**: 설정 로드 로직.
- **`internal/config/keys/<env>/private.pem`, `public.pem`**: JWT RSA 키.
- **`deploy/k8s/customer-id.yaml`**: Secret 및 환경 변수 주입 설정.
- **`Makefile`**: 키 생성 및 실행 명령어.

---

## 8. 결론
TasteSync의 설정 관리는 `config.yaml`과 `config.go`로 통합되어 환경별 유연성과 보안성을 보장합니다. 개발자는 이 가이드를 따라 설정을 관리하고, 운영 환경에서 Secret과 키를 안전하게 적용하세요.