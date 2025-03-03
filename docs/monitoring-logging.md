# TasteSync Monitoring and Logging

이 문서는 TasteSync 플랫폼의 모니터링 및 로깅 전략을 정의합니다. `tastesync-customer-id` 서비스를 중심으로, 2027년 v1.0 출시 시 500~1,000 사용자(식당 중심)를 지원하며, 2040년 1,000만 사용자까지 확장 가능하도록 설계되었습니다. 초저지연(1초 내 응답), 안정성(99.9% 가동률), 서비스 품질 유지를 위해 Prometheus/Grafana 기반의 메트릭 수집과 구조화된 로깅 시스템을 구현합니다. 모든 모니터링과 로깅 데이터는 실시간 분석 및 장기 보존을 지원합니다.

---

## 1. 모니터링 및 로깅 개요

### 1.1 목표
- **초저지연 추적**: 1초 내 응답 시간 보장 여부 실시간 확인.
- **안정성 보장**: 99.9% 가동률 유지, 장애 복구 5분 내(RTO).
- **품질 유지**: 서비스 성능, 오류, 사용자 행동 모니터링.
- **문제 진단**: 로깅으로 원인 분석 및 신속 대응.
- **확장성**: 대규모 트래픽(1,000만 사용자) 처리 가능.

### 1.2 적용 범위
- **`customer-id`**: 고객 식별 서비스 (HTTP/gRPC 엔드포인트).
- **의존성**: Redis, PostgreSQL, Kafka.
- **클라이언트**: 모바일/패드 앱, POS 앱, 홈페이지.

---

## 2. 모니터링 전략

### 2.1 모니터링 도구
- **Prometheus**: 메트릭 수집 및 저장.
- **Grafana**: 시각화 대시보드.
- **Alertmanager**: 경고 알림.

### 2.2 설정 파일
- **Prometheus**: `deploy/k8s/prometheus-config.yaml`.
- **예시**:
  ```yaml
  global:
    scrape_interval: 15s
  scrape_configs:
  - job_name: 'customer-id'
    static_configs:
    - targets: ['customer-id:3000']
      labels: { service: 'customer-id' }
  ```
- **Grafana**: `deploy/k8s/grafana-dashboard.json` (대시보드 정의).

### 2.3 수집 메트릭

#### 2.3.1 HTTP/gRPC 엔드포인트
- **메트릭**:
  - `http_request_duration_seconds`: 요청 처리 시간 (Histogram).
    - Label: `method`, `path`, `status`.
    - 예: `http_request_duration_seconds{method="POST", path="/identify", status="200"}`.
  - `grpc_request_total`: gRPC 요청 수 (Counter).
    - Label: `method`, `status`.
  - `request_errors_total`: 오류 요청 수 (Counter).
- **구현**: 
  - 경로: `internal/infrastructure/grpc/server.go`.
  - 예시:
    ```go
    import "github.com/prometheus/client_golang/prometheus"

    var (
        httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Buckets: []float64{0.1, 0.5, 1.0, 2.0},
        }, []string{"method", "path", "status"})
    )

    func init() {
        prometheus.MustRegister(httpDuration)
    }
    ```

#### 2.3.2 서비스 상태
- **메트릭**:
  - `up`: 서비스 가동 여부 (Gauge, 1=정상, 0=다운).
  - `memory_usage_bytes`: 메모리 사용량 (Gauge).
  - `cpu_usage_percentage`: CPU 사용률 (Gauge).
- **구현**: `/metrics` 엔드포인트에서 제공.

#### 2.3.3 의존성
- **Redis**:
  - `redis_commands_total`: 명령 실행 수 (Counter).
  - `redis_latency_seconds`: 명령 지연 시간 (Histogram).
- **PostgreSQL**:
  - `pg_queries_total`: 쿼리 실행 수 (Counter).
  - `pg_query_duration_seconds`: 쿼리 실행 시간 (Histogram).
- **Kafka**:
  - `kafka_messages_produced_total`: 발행 메시지 수 (Counter).
  - `kafka_produce_latency_seconds`: 발행 지연 시간 (Histogram).
- **구현**: `internal/infrastructure/<dependency>/metrics.go`.

### 2.4 대시보드 (Grafana)
- **대시보드**: 
  - 이름: `CustomerID Monitoring`.
  - 패널:
    - 요청 지연 시간 (HTTP/gRPC).
    - 오류율 그래프.
    - 서비스 상태 (Up/Down).
    - 의존성 성능 (Redis, PostgreSQL, Kafka).
- **파일**: `deploy/k8s/grafana-dashboard.json`.

### 2.5 경고 설정 (Alertmanager)
- **파일**: `deploy/k8s/alert-rules.yaml`.
- **예시**:
  ```yaml
  groups:
  - name: customer-id-alerts
    rules:
    - alert: HighLatency
      expr: rate(http_request_duration_seconds[5m]) > 1
      for: 1m
      labels: { severity: "critical" }
      annotations:
        summary: "High request latency detected"
        description: "Average latency exceeds 1s for 1 minute."
    - alert: ServiceDown
      expr: up{service="customer-id"} == 0
      for: 1m
      labels: { severity: "critical" }
      annotations:
        summary: "CustomerID service is down"
  ```
- **알림**: Slack/이메일로 전송 (Alertmanager 설정 필요).

---

## 3. 로깅 전략

### 3.1 로깅 도구
- **Zap**: Go용 고성능 구조화 로깅 라이브러리.
- **Fluentd**: 로그 수집 및 S3 전송.

### 3.2 설정 파일
- **Zap**: `internal/infrastructure/logging/logger.go`.
- **예시**:
  ```go
  package logging

  import (
      "go.uber.org/zap"
      "github.com/yourusername/tastesync-customer-id/internal/config"
  )

  func NewLogger(cfg *config.Config) (*zap.Logger, error) {
      zapCfg := zap.NewProductionConfig()
      zapCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
      if cfg.Logging.Level == "debug" {
          zapCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
      }
      zapCfg.OutputPaths = []string{cfg.Logging.Output}
      return zapCfg.Build()
  }
  ```
- **Fluentd**: `deploy/k8s/fluentd-config.yaml`.

### 3.3 로그 레벨
- **Debug**: 개발/디버깅용 상세 로그 (예: "Beacon data received").
- **Info**: 정상 동작 확인 (예: "Customer identified: cust123").
- **Warn**: 경미한 문제 (예: "Redis cache miss").
- **Error**: 오류 발생 (예: "Failed to identify customer").

### 3.4 로그 형식
- **구조화**: JSON 형식으로 필드 정의.
- **예시**:
  ```json
  {
    "timestamp": "2025-03-02T12:00:00.123Z",
    "level": "info",
    "service": "customer-id",
    "event_id": "evt-uuid-1234-5678-9012",
    "message": "Customer identified",
    "data": {
      "customer_id": "cust123",
      "location": "Table 3",
      "confidence": 0.95
    }
  }
  ```
- **필드**:
  - `timestamp`: UTC, millisecond 단위.
  - `level`: 로그 레벨.
  - `service`: 서비스 이름.
  - `event_id`: 이벤트 식별자 (선택).
  - `message`: 로그 메시지.
  - `data`: 추가 데이터.

### 3.5 로그 수집 및 저장
- **수집**: Fluentd로 stdout 로그 수집.
- **저장**: 
  - S3: `s3://tastesync-logs/customer-id/YYYY/MM/DD/`.
  - 보존 기간: 90일 (GDPR 준수).
- **설정**: 
  - Fluentd: `deploy/k8s/fluentd-config.yaml`.
  - 예시:
    ```yaml
    <match customer-id.*>
      @type s3
      s3_bucket tastesync-logs
      path customer-id/
      buffer_path /var/log/fluentd/buffer
    </match>
    ```

### 3.6 로그 분석
- **도구**: AWS Athena로 S3 로그 쿼리.
- **예시 쿼리**:
  ```sql
  SELECT timestamp, level, message
  FROM "tastesync_logs"."customer_id"
  WHERE level = 'error' AND timestamp > '2025-03-01'
  LIMIT 100;
  ```

---

## 4. 모니터링 및 로깅 연계

### 4.1 실시간 모니터링
- **경고**: Prometheus → Alertmanager → Slack.
- **로그 확인**: 오류 발생 시 S3 로그 즉시 조회.

### 4.2 문제 진단 워크플로우
1. **경고 수신**: "HighLatency" 알림.
2. **메트릭 확인**: Grafana에서 지연 시간 그래프 분석.
3. **로그 조회**: Athena로 해당 시간대 Error 로그 검색.
4. **원인 파악**: 코드/의존성 문제 진단.
5. **조치**: PR로 수정 후 배포.

---

## 5. 필요한 파일 및 경로
- **`internal/infrastructure/logging/logger.go`**: Zap 로깅 설정.
- **`internal/infrastructure/<dependency>/metrics.go`**: 의존성 메트릭.
- **`deploy/k8s/prometheus-config.yaml`**: Prometheus 설정.
- **`deploy/k8s/alert-rules.yaml`**: 경고 규칙.
- **`deploy/k8s/grafana-dashboard.json`**: Grafana 대시보드.
- **`deploy/k8s/fluentd-config.yaml`**: Fluentd 설정.
- **`scripts/log-query.sh`**: 로그 분석 스크립트 (옵션).

---

## 6. 결론
`tastesync-customer-id`의 모니터링과 로깅은 Prometheus, Grafana, Zap, Fluentd를 통해 초저지연과 안정성을 보장합니다. 실시간 메트릭과 구조화된 로그로 서비스 품질을 유지하며, 문제 발생 시 신속히 대응할 수 있습니다.