# TasteSync API Specification

이 문서는 TasteSync 플랫폼의 API 인터페이스를 정의합니다. `tastesync-customer-id` 서비스를 중심으로, 고객 식별 기능을 제공하는 gRPC API를 상세히 설명하며, 2027년 v1.0 출시 시 500~1,000 사용자를 지원합니다. 외부 클라이언트는 HTTPS로 요청을 보내고, Envoy API Gateway에서 gRPC로 변환되어 내부 서비스와 통신합니다. 추후 RESTful API 추가 가능성을 고려해 확장성을 유지합니다.

---

## 1. API 개요

### 1.1 목표
- **초저지연**: 비콘/QRS 감지 후 1초 내 응답.
- **명확성**: 클라이언트와 서버 간 통신 규약 정의.
- **확장성**: gRPC 기반으로 대규모 트래픽 처리.

### 1.2 통신 방식
- **외부**: HTTPS (TLS 1.3) → Envoy에서 gRPC로 변환.
- **내부**: gRPC (Go 1.23 기반, TLS 암호화).
- **인증**: JWT (RSA 2048비트) 토큰 사용.

### 1.3 기본 엔드포인트
- **HTTPS**: `https://api.tastesync.com/customer-id/identify` (외부 호출용).
- **gRPC**: `grpc://customer-id.tastesync.svc:50051` (내부 호출용).

---

## 2. gRPC API 정의

### 2.1 서비스 정의
- **파일**: `proto/customer_id.proto`.
- **패키지**: `customerid`.
- **기본 구조**:
  ```proto
  syntax = "proto3";
  package customerid;
  option go_package = "github.com/yourusername/tastesync-customer-id/proto";

  service CustomerID {
    // IdentifyCustomer identifies a customer based on beacon or QRS data.
    rpc IdentifyCustomer (IdentifyRequest) returns (IdentifyResponse) {}
  }
  ```

### 2.2 메시지 정의

#### IdentifyRequest
- **설명**: 비콘 또는 QRS 데이터를 기반으로 고객 식별 요청.
- **구조**:
  ```proto
  message IdentifyRequest {
    // Unique identifier of the beacon (UUID).
    string uuid = 1;
    // Major identifier (e.g., store identifier).
    int32 major = 2;
    // Minor identifier (e.g., entrance or table number).
    int32 minor = 3;
    // Received Signal Strength Indicator (RSSI) for distance estimation.
    int32 rssi = 4;
    // Timestamp of the beacon detection (ISO 8601 format).
    string timestamp = 5;
  }
  ```
- **제약**:
  - `uuid`: 필수, 36자 UUID 형식 (예: `550e8400-e29b-41d4-a716-446655440000`).
  - `major`, `minor`: 0~65535 범위.
  - `rssi`: -100~0 dBm 범위.
  - `timestamp`: UTC 기준 (예: `2025-03-02T12:00:00Z`).

#### IdentifyResponse
- **설명**: 고객 식별 결과 반환.
- **구조**:
  ```proto
  message IdentifyResponse {
    // Identified customer ID.
    string customer_id = 1;
    // Location where the customer was identified (e.g., "Entrance", "Table 3").
    string location = 2;
    // Confidence score of identification (0.0~1.0).
    float confidence = 3;
  }
  ```
- **제약**:
  - `customer_id`: 고유 식별자 (최대 64자).
  - `location`: 최대 32자.
  - `confidence`: 0.0~1.0 (1.0 = 100% 확신).

### 2.3 메서드 상세

#### IdentifyCustomer
- **설명**: 비콘/QRS 데이터를 받아 고객을 식별하고 위치를 반환.
- **입력**: `IdentifyRequest`.
- **출력**: `IdentifyResponse`.
- **에러로그**:
  - `INVALID_ARGUMENT` (3): 요청 데이터 형식 오류 (예: UUID 누락).
  - `NOT_FOUND` (5): 고객 식별 실패.
  - `INTERNAL` (13): 서버 내부 오류.
- **예시**:
  ```proto
  // 요청
  IdentifyRequest {
    uuid: "550e8400-e29b-41d4-a716-446655440000",
    major: 100,
    minor: 3,
    rssi: -50,
    timestamp: "2025-03-02T12:00:00Z"
  }

  // 응답
  IdentifyResponse {
    customer_id: "cust123",
    location: "Table 3",
    confidence: 0.95
  }
  ```

---

## 3. HTTPS 호출 방식

### 3.1 엔드포인트
- **URL**: `POST https://api.tastesync.com/customer-id/identify`.
- **헤더**:
  - `Authorization: Bearer <JWT_TOKEN>` (RSA 기반).
  - `Content-Type: application/json`.

### 3.2 요청 본문
- **형식**: JSON.
- **예시**:
  ```json
  {
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "major": 100,
    "minor": 3,
    "rssi": -50,
    "timestamp": "2025-03-02T12:00:00Z"
  }
  ```

### 3.3 응답 본문
- **형식**: JSON.
- **예시**:
  ```json
  {
    "customer_id": "cust123",
    "location": "Table 3",
    "confidence": 0.95
  }
  ```
- **에러 응답**:
  ```json
  {
    "error": {
      "code": 3,
      "message": "Invalid UUID format"
    }
  }
  ```

### 3.4 상태 코드
- **200**: 성공.
- **400**: 요청 형식 오류.
- **404**: 고객 미식별.
- **500**: 서버 오류.

---

## 4. 인증

### 4.1 JWT 토큰
- **알고리즘**: RSA 2048비트.
- **구조**:
  - **Header**: `{"alg": "RS256", "typ": "JWT"}`.
  - **Payload**: `{"sub": "client_id", "exp": 1648771200}`.
  - **Signature**: RSA 개인 키로 서명.
- **키 관리**: 
  - 개발: `internal/config/keys/dev/`.
  - 운영: `internal/config/keys/prod/`.
- **획득**: 클라이언트는 별도 인증 서버에서 발급 (미구현 시 임시 키 제공).

### 4.2 사용 방법
- HTTPS 요청에 `Authorization` 헤더로 포함.
- gRPC는 Envoy에서 JWT 검증 후 전달.

---

## 5. 확장 고려사항

### 5.1 RESTful API 추가
- **방법**: Envoy에 REST 라우팅 추가 (예: `/v1/customer-id/identify`).
- **구현 시점**: 필요 시 별도 정의 후 통합.

### 5.2 버전 관리
- gRPC 프로토콜은 `v1`부터 시작, 하위 호환성 유지.
- 경로: `proto/v1/customer_id.proto`로 확장 가능.

---

## 6. 결론
`tastesync-customer-id`의 API는 gRPC를 중심으로 초저지연을 구현하며, HTTPS를 통해 외부 클라이언트와 통신합니다. JWT RSA 인증으로 보안을 강화하고, RESTful API 추가 가능성을 열어둡니다. 클라이언트 개발자는 이 스펙을 참고하여 통신을 설계하세요.