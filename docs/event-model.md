# TasteSync Event Model

이 문서는 TasteSync 플랫폼의 이벤트 드리븐 아키텍처에서 사용되는 이벤트 모델을 정의합니다. `tastesync-customer-id` 서비스를 중심으로, 2027년 v1.0 출시 시 500~1,000 사용자(식당 중심)를 지원하며, 2040년 1,000만 사용자까지 확장 가능하도록 설계되었습니다. 초저지연(1초 내)과 비동기 처리를 위해 Kafka와 NATS를 병행하며, 고객 식별 이벤트를 기반으로 추천, 알림, 분석 등 후속 작업을 트리거합니다.

---

## 1. 이벤트 모델 개요

### 1.1 목표
- **초저지연**: 이벤트 발행 후 1초 내 구독 서비스 반응.
- **확장성**: 대규모 사용자(1,000만) 이벤트 처리 지원.
- **내구성**: 이벤트 손실 방지 및 재처리 가능성 보장.
- **유연성**: 다양한 서비스 간 비동기 통신 지원.

### 1.2 이벤트 드리븐 아키텍처 특징
- **발행/구독 모델**: 서비스가 이벤트를 발행(Publish)하고, 관련 서비스가 구독(Subscribe).
- **메시지 큐**: 
  - **Kafka**: 대규모 이벤트 스트리밍 및 내구성.
  - **NATS**: 초저지연 실시간 알림.
- **도메인 이벤트**: `CustomerIdentified` 등으로 고객 식별 완료 후 후속 작업 트리거.

---

## 2. 이벤트 정의

### 2.1 CustomerIdentified 이벤트
- **설명**: 고객이 비콘/QRS로 식별된 후 발생하는 도메인 이벤트.
- **발생 시점**: `customer-id` 서비스가 `IdentifyCustomer` 호출을 성공적으로 처리했을 때.
- **구조**:
  ```json
  {
    "event_id": "evt-uuid-1234-5678-9012",
    "event_type": "CustomerIdentified",
    "timestamp": "2025-03-02T12:00:00Z",
    "version": "v1",
    "data": {
      "customer_id": "cust123",
      "location": "Table 3",
      "confidence": 0.95,
      "beacon": {
        "uuid": "550e8400-e29b-41d4-a716-446655440000",
        "major": 100,
        "minor": 3,
        "rssi": -50
      },
      "detected_at": "2025-03-02T12:00:00Z"
    }
  }
  ```
- **필드 설명**:
  - `event_id`: 고유 이벤트 식별자 (UUID 형식).
  - `event_type`: 이벤트 이름 (`CustomerIdentified`).
  - `timestamp`: 이벤트 발생 시각 (ISO 8601, UTC).
  - `version`: 이벤트 스키마 버전 (예: `v1`).
  - `data`: 이벤트 페이로드.
    - `customer_id`: 식별된 고객 ID.
    - `location`: 고객 위치 (예: "Entrance", "Table 3").
    - `confidence`: 식별 신뢰도 (0.0~1.0).
    - `beacon`: 비콘 데이터 (UUID, Major, Minor, RSSI).
    - `detected_at`: 비콘 감지 시각.
- **제약**:
  - `event_id`: 필수, 중복 불가.
  - `timestamp`, `detected_at`: UTC 기준, millisecond 단위까지 가능.
  - `confidence`: 0.0~1.0 범위.

---

## 3. 이벤트 발행 및 구독

### 3.1 발행 (Publishing)
- **서비스**: `customer-id`.
- **구현**: 
  - 경로: `internal/infrastructure/kafka/publisher.go`.
  - 예시:
    ```go
    package kafka

    type EventPublisher struct {
        producer *kafka.Producer
    }

    func (p *EventPublisher) PublishCustomerIdentified(event domain.CustomerIdentified) error {
        msg := &kafka.Message{
            TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
            Value:          json.Marshal(event),
        }
        return p.producer.Produce(msg, nil)
    }
    ```
- **메시지 큐**: 
  - **Kafka**: `CustomerIdentified` 이벤트를 `customer-events` 토픽으로 발행.
    - 설정: `internal/config/config.yaml`의 `kafka.broker`, `kafka.topic`.
  - **NATS**: 실시간 알림 필요 시 `notifications` 주제로 발행 (추후 확장).

### 3.2 구독 (Subscribing)
- **구독 서비스**: 
  - `recommendation`: 추천 생성.
  - `notification`: 실시간 알림 전송.
  - `analytics`: 방문/위치 분석.
- **구현**: 
  - Kafka Consumer로 `customer-events` 토픽 구독.
  - 예시:
    ```go
    package kafka

    type Consumer struct {
        consumer *kafka.Consumer
    }

    func (c *Consumer) SubscribeCustomerEvents(handler func(event domain.CustomerIdentified)) {
        c.consumer.SubscribeTopics([]string{"customer-events"}, nil)
        for {
            msg, err := c.consumer.ReadMessage(-1)
            if err == nil {
                var event domain.CustomerIdentified
                json.Unmarshal(msg.Value, &event)
                handler(event)
            }
        }
    }
    ```
- **NATS**: 초저지연 요구 시 별도 구현 가능.

---

## 4. 메시지 큐 설정

### 4.1 Kafka
- **토픽**: `customer-events`.
- **파티션**: 최소 3개 (스케일링 대비).
- **복제**: 복제 팩터 3 (내구성 보장).
- **설정**: 
  - `deploy/docker/docker-compose.yml`에 Kafka 브로커 정의.
  - 프로덕션: AWS MSK 또는 자체 Kafka 클러스터.
- **예시**:
  ```yaml
  kafka:
    image: confluentinc/cp-kafka:latest
    environment:
      KAFKA_NUM_PARTITIONS: 3
      KAFKA_REPLICATION_FACTOR: 3
  ```

### 4.2 NATS
- **주제**: `notifications` (추후 확장 시 사용).
- **설정**: 
  - 초저지연 알림 전용.
  - 파일: `deploy/docker/docker-compose.yml`에 NATS 추가 가능.
- **예시**:
  ```yaml
  nats:
    image: nats:latest
    ports: [ "4222:4222" ]
  ```

### 4.3 병행 사용 전략
- **Kafka**: 대규모 이벤트 스트리밍, 내구성, 재처리.
- **NATS**: 실시간 알림(예: "고객 A 입장") 초저지연.
- **구분**: 
  - `CustomerIdentified` → Kafka (기본).
  - 실시간 요구 시 NATS 병행 (v2.0 이후 고려).

---

## 5. 이벤트 처리 정책

### 5.1 발행 정책
- **멱등성**: `event_id`로 중복 발행 방지.
- **순서 보장**: Kafka 파티션별 순서 유지.
- **재시도**: 발행 실패 시 최대 3회 재시도 (지수 백오프).

### 5.2 구독 정책
- **최소 1회 전달**: Kafka로 이벤트 손실 방지.
- **중복 처리**: `event_id`로 중복 구독 감지 및 스킵.
- **오류 복구**: Consumer 그룹으로 실패 시 재시작.

### 5.3 모니터링
- **메트릭**: 
  - 이벤트 발행 속도, 지연 시간, 오류율.
  - Prometheus로 수집 (`deploy/k8s/prometheus-config.yaml`).
- **로그**: 
  - 발행/구독 로그 기록 (`internal/infrastructure/logging/`).

---

## 6. 필요한 파일 및 경로
- **`proto/customer_id.proto`**: 이벤트 관련 메시지 정의 (참고용).
- **`internal/infrastructure/kafka/publisher.go`**: Kafka 발행 로직.
- **`internal/infrastructure/kafka/consumer.go`**: Kafka 구독 로직.
- **`internal/config/config.yaml`**: Kafka/NATS 설정.
- **`deploy/docker/docker-compose.yml`**: 로컬 메시지 큐 설정.
- **`deploy/k8s/kafka-config.yaml`**: 프로덕션 Kafka 설정.

---

## 7. 결론
TasteSync의 이벤트 모델은 `CustomerIdentified` 이벤트를 중심으로 Kafka와 NATS를 활용해 초저지연과 내구성을 동시에 충족합니다. 이 모델을 통해 `customer-id`는 비동기적으로 추천, 알림, 분석 서비스와 통합되며, 1,000만 사용자 규모에서도 안정적으로 동작합니다.
