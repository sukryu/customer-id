# TasteSync Data Model

이 문서는 TasteSync 플랫폼의 데이터 모델을 정의합니다. `tastesync-customer-id` 서비스를 중심으로, 2027년 v1.0 출시 시 500~1,000 사용자(식당 중심)를 지원하며, 2040년 1,000만 사용자까지 확장 가능하도록 설계되었습니다. DDD 원칙을 따라 도메인 중심으로 엔티티, 애그리게이트, 값 객체를 정의하며, 초저지연(1초 내 응답)과 대규모 데이터 처리를 위해 인덱스, 파티션 등의 최적화를 적용합니다. 데이터는 PostgreSQL(구조화 저장), Redis(캐싱), DynamoDB(트랜잭션/분석)에 저장됩니다.

---

## 1. 데이터 모델 개요

### 1.1 목표
- **초저지연**: 1초 내 데이터 조회/저장 보장.
- **확장성**: 1,000만 사용자 데이터 처리 지원.
- **무결성**: 도메인 규칙에 따른 데이터 일관성 유지.
- **소상공인 배려**: 고객 식별 및 위치 데이터를 기반으로 추천/분석 제공.

### 1.2 설계 원칙
- **DDD**: 도메인 중심 엔티티/애그리게이트 정의.
- **헥사고날**: 저장소 독립적인 데이터 모델.
- **성능 최적화**: 인덱스, 파티션, 캐싱으로 효율성 극대화.

---

## 2. 도메인 모델

### 2.1 엔티티 (Entities)

#### 2.1.1 Customer
- **설명**: 식별된 고객을 나타냄.
- **속성**:
  - `CustomerID` (string): 고유 식별자 (예: "cust123").
  - `LastSeen` (timestamp): 마지막 식별 시간 (예: "2025-03-02T12:00:00Z").
  - `Preferences` (map[string]string): 고객 선호도 (예: {"drink": "coffee"}).
- **제약**:
  - `CustomerID`: 필수, 최대 64자.
  - `LastSeen`: UTC 기준, 업데이트 시 변경.

#### 2.1.2 Beacon
- **설명**: 비콘 장치를 나타냄.
- **속성**:
  - `BeaconID` (string): 고유 식별자 (UUID, 예: "550e8400-e29b-41d4-a716-446655440000").
  - `StoreID` (string): 매장 식별자 (예: "store100").
  - `Major` (int32): 매장 내 그룹 (예: 100).
  - `Minor` (int32): 세부 위치 (예: 3).
  - `Location` (string): 비콘 설치 위치 (예: "Table 3").
  - `Status` (string): 상태 (예: "active", "inactive").
- **제약**:
  - `BeaconID`: 필수, UUID 형식.
  - `Major`, `Minor`: 0~65535 범위.
  - `Status`: 열거형 (`active`, `inactive`, `maintenance`).

### 2.2 애그리게이트 (Aggregates)

#### 2.2.1 CustomerIdentity
- **설명**: 고객 식별과 관련된 핵심 데이터 집합.
- **구성**:
  - 루트 엔티티: `Customer`.
  - 포함 엔티티: `Beacon`.
- **속성**:
  - `CustomerID` (string): 고객 ID.
  - `BeaconID` (string): 연관된 비콘 ID.
  - `Location` (string): 식별된 위치.
  - `Confidence` (float32): 식별 신뢰도 (0.0~1.0).
  - `DetectedAt` (timestamp): 식별 시각.
- **도메인 규칙**:
  - `Confidence` ≥ 0.8 요구.
  - 동일 `CustomerID`에 대해 1분 내 중복 식별 금지.

### 2.3 값 객체 (Value Objects)

#### 2.3.1 BeaconData
- **설명**: 비콘 감지 시 수신된 원시 데이터.
- **속성**:
  - `UUID` (string): 비콘 UUID.
  - `Major` (int32): 비콘 Major 값.
  - `Minor` (int32): 비콘 Minor 값.
  - `RSSI` (int32): 신호 강도 (-100~0 dBm).
- **제약**: 모든 필드 필수.

#### 2.3.2 Location
- **설명**: 고객의 위치 정보.
- **속성**:
  - `Name` (string): 위치 이름 (예: "Table 3").
  - `Type` (string): 위치 유형 (예: "entrance", "table").
- **제약**:
  - `Name`: 최대 32자.
  - `Type`: 열거형 (`entrance`, `table`, `counter`).

---

## 3. 데이터 저장소 설계 및 최적화

### 3.1 PostgreSQL (구조화 저장)
- **테이블 설계**:
  - **`customers`**:
    ```sql
    CREATE TABLE customers (
        customer_id VARCHAR(64) PRIMARY KEY,
        last_seen TIMESTAMP WITH TIME ZONE NOT NULL,
        preferences JSONB DEFAULT '{}',
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    CREATE INDEX idx_customers_last_seen ON customers(last_seen);
    ```
    - **최적화**:
      - `idx_customers_last_seen`: 최근 활동 기반 조회 속도 향상.
      - `JSONB`: 선호도 데이터의 동적 확장 지원.

  - **`beacons`**:
    ```sql
    CREATE TABLE beacons (
        beacon_id VARCHAR(36) PRIMARY KEY,
        store_id VARCHAR(64) NOT NULL,
        major INT NOT NULL CHECK (major >= 0 AND major <= 65535),
        minor INT NOT NULL CHECK (minor >= 0 AND minor <= 65535),
        location VARCHAR(32),
        status VARCHAR(16) NOT NULL DEFAULT 'active',
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        CONSTRAINT valid_status CHECK (status IN ('active', 'inactive', 'maintenance'))
    );
    CREATE INDEX idx_beacons_store_id ON beacons(store_id);
    CREATE INDEX idx_beacons_status ON beacons(status);
    ```
    - **최적화**:
      - `idx_beacons_store_id`: 매장별 비콘 조회 최적화.
      - `idx_beacons_status`: 활성 상태 필터링 속도 향상.

  - **`customer_identities`**:
    ```sql
    CREATE TABLE customer_identities (
        id BIGSERIAL PRIMARY KEY,
        customer_id VARCHAR(64) NOT NULL REFERENCES customers(customer_id),
        beacon_id VARCHAR(36) NOT NULL REFERENCES beacons(beacon_id),
        location VARCHAR(32),
        confidence REAL NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
        detected_at TIMESTAMP WITH TIME ZONE NOT NULL
    ) PARTITION BY RANGE (detected_at);
    CREATE TABLE customer_identities_2025 PARTITION OF customer_identities
        FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');
    CREATE INDEX idx_customer_identities_customer_id ON customer_identities(customer_id);
    CREATE INDEX idx_customer_identities_detected_at ON customer_identities(detected_at);
    CREATE UNIQUE INDEX idx_customer_identities_unique ON customer_identities(customer_id, detected_at);
    ```
    - **최적화**:
      - **파티션**: `detected_at` 기준 연도별 범위 파티셔닝으로 대규모 데이터 관리 효율화 (예: 1,000만 레코드 시 조회 속도 개선).
      - `idx_customer_identities_customer_id`: 고객별 식별 이력 조회 최적화.
      - `idx_customer_identities_detected_at`: 시간 기반 조회 속도 향상.
      - `idx_customer_identities_unique`: 중복 식별 방지 (1분 내 동일 고객 체크).

- **인덱스 전략**:
  - B-Tree 인덱스 사용하여 범위 조회 및 정렬 최적화.
  - 복합 인덱스 대신 단일 컬럼 인덱스로 유지보수성과 유연성 확보.

### 3.2 Redis (캐싱)
- **키 설계**:
  - `customer:<customer_id>`: 최근 식별 데이터 (TTL 1시간).
    - 예: `customer:cust123` → `{"location": "Table 3", "confidence": 0.95, "detected_at": "2025-03-02T12:00:00Z"}`.
  - `beacon:<beacon_id>`: 비콘 메타데이터 (TTL 24시간).
    - 예: `beacon:550e8400-e29b-41d4-a716-446655440000` → `{"store_id": "store100", "location": "Table 3"}`.
- **최적화**:
  - **Hash 구조**: `HSET`으로 키-값 쌍 저장, 메모리 사용량 최소화.
    ```bash
    HSET customer:cust123 location "Table 3" confidence 0.95 detected_at "2025-03-02T12:00:00Z"
    ```
  - **TTL**: 자주 사용되는 데이터만 캐싱, 캐시 무효화 속도 향상.
  - **클러스터링**: 1,000만 사용자 대비 Redis Cluster로 샤딩 준비 (v2.0 이후).

- **구현**: `internal/infrastructure/redis/cache.go`.

### 3.3 DynamoDB (트랜잭션/분석)
- **테이블 설계**: `CustomerEvents`.
- **구조**:
  - **Partition Key**: `CustomerID` (string).
  - **Sort Key**: `DetectedAt` (ISO 8601 timestamp).
  - **Attributes**:
    - `Location` (string).
    - `Confidence` (number).
    - `BeaconID` (string).
- **최적화**:
  - **파티션**: `CustomerID`로 데이터 분산, 핫 파티션 방지 위해 랜덤 접미사 추가 가능 (예: `cust123-1`).
  - **GSI (Global Secondary Index)**:
    - 이름: `DetectedAtIndex`.
    - Partition Key: `DetectedAt`.
    - 속성: `CustomerID`, `Location`.
    - 용도: 시간 기반 분석 쿼리 최적화.
  - **읽기/쓰기 용량**: 
    - 초기: 10 RCU(Read Capacity Units), 10 WCU(Write Capacity Units).
    - 자동 스케일링: 최대 1,000 RCU/WCU로 설정 (1,000만 사용자 대비).
- **예시**:
  ```json
  {
    "CustomerID": "cust123",
    "DetectedAt": "2025-03-02T12:00:00Z",
    "Location": "Table 3",
    "Confidence": 0.95,
    "BeaconID": "550e8400-e29b-41d4-a716-446655440000"
  }
  ```

---

## 4. 데이터 모델 사용 예시

### 4.1 고객 식별 프로세스
- **입력**: `BeaconData` (UUID, Major, Minor, RSSI).
- **처리**:
  1. Redis에서 `beacon:<beacon_id>` 조회 → `StoreID`, `Location` 확인.
  2. 없으면 PostgreSQL `beacons` 조회 → 캐시 업데이트.
  3. `CustomerIdentity` 생성 → `CustomerID`, `Confidence` 계산.
  4. PostgreSQL `customers` 업데이트 → `LastSeen`, `Preferences`.
- **출력**: `CustomerIdentified` 이벤트 발행 → DynamoDB에 기록.

### 4.2 캐싱 전략
- **Redis 저장**:
  ```go
  cache.HSet("customer:cust123", map[string]interface{}{
      "location": identity.Location,
      "confidence": identity.Confidence,
      "detected_at": identity.DetectedAt,
  }, time.Hour)
  ```
- **조회**:
  ```go
  if data, err := cache.HGetAll("customer:cust123"); err == nil {
      // 데이터 파싱
  }
  ```

---

## 5. 데이터 모델 관리 정책

### 5.1 데이터 무결성
- **검증**: `CustomerIdentity` 생성 시 `Confidence` ≥ 0.8 확인.
- **중복 방지**: PostgreSQL의 복합 키(`customer_id`, `detected_at`)로 중복 차단.

### 5.2 데이터 보존
- **PostgreSQL**: 고객 데이터 5년 보존 (GDPR 준수).
- **Redis**: TTL 1시간 (캐싱), 24시간 (비콘 메타데이터).
- **DynamoDB**: 1년 보존 후 Glacier로 아카이빙.

### 5.3 데이터 마이그레이션
- **스크립트**: `scripts/migrate.sh`.
- **예시**:
  ```bash
  psql -h localhost -U tastesync -d tastesync -f scripts/migrate-2025-partition.sql
  ```
- **파티션 추가**: 연도별 테이블 생성 (`customer_identities_2026` 등).

---

## 6. 필요한 파일 및 경로
- **`internal/domain/entities/`**: `Customer`, `Beacon` 정의.
- **`internal/domain/aggregates/`**: `CustomerIdentity` 정의.
- **`internal/infrastructure/redis/cache.go`**: Redis 캐싱 로직.
- **`internal/infrastructure/db/schema.sql`**: PostgreSQL 스키마.
- **`scripts/migrate.sh`**: 마이그레이션 스크립트.
- **`deploy/k8s/dynamodb-config.yaml`**: DynamoDB 설정 (옵션).

---

## 7. 결론
TasteSync의 데이터 모델은 `Customer`, `Beacon`, `CustomerIdentity`를 중심으로 초저지연과 확장성을 구현하며, 인덱스와 파티션으로 최적화된 PostgreSQL, Redis, DynamoDB를 통해 효율적으로 관리됩니다. 이 모델은 소상공인 배려와 서비스 품질을 뒷받침합니다.