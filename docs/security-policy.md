# TasteSync Security Policy

이 문서는 TasteSync 플랫폼의 보안 정책을 정의합니다. `tastesync-customer-id` 서비스를 중심으로, 2027년 v1.0 출시 시 500~1,000 사용자(식당 중심)를 지원하며, 2040년 1,000만 사용자까지 확장합니다. 초저지연(1초 내)과 확장성을 유지하면서 데이터 보안, 통신 보안, 인증/권한 관리, 취약점 대응을 철저히 보장합니다. GDPR, PIPA 등 규제 준수를 목표로 하며, OWASP Top 10을 기준으로 설계되었습니다.

---

## 1. 보안 개요

### 1.1 목표
- **데이터 보호**: 고객 식별 데이터의 기밀성, 무결성, 가용성 보장.
- **통신 보안**: HTTPS 및 gRPC 통신의 암호화 및 인증.
- **접근 제어**: 권한 있는 사용자/서비스만 접근 허용.
- **취약점 관리**: 보안 위협 사전 예방 및 신속 대응.
- **규제 준수**: GDPR, PIPA 등 개인정보 보호법 준수.

### 1.2 적용 범위
- **`customer-id`**: 비콘/QRS 기반 고객 식별 서비스.
- **의존성**: Redis, PostgreSQL, Kafka, Envoy API Gateway.
- **클라이언트**: 모바일/패드 앱, POS 앱, 홈페이지, 관리 대시보드.

---

## 2. 통신 보안

### 2.1 HTTPS (외부 통신)
- **프로토콜**: TLS 1.3 (최소 TLS 1.2 허용).
- **인증서**: 
  - Let's Encrypt 또는 AWS Certificate Manager(ACM) 사용.
  - 갱신 주기: 90일마다 자동 갱신.
  - 경로: `deploy/k8s/customer-id.yaml`의 `tls` 설정 참조.
- **설정**:
  - Envoy에서 HTTPS 종료 후 gRPC로 변환.
  - HSTS 헤더 강제 적용 (최대 1년).
- **구현**:
  - 파일: `tastesync-backend-common/envoy.yaml`.
  - 예시:
    ```yaml
    listeners:
    - filter_chains:
      - filters:
        - name: envoy.filters.network.http_connection_manager
          config:
            http_filters:
            - name: envoy.filters.http.hsts
              config: { max_age_seconds: 31536000 }
    ```

### 2.2 gRPC (내부 통신)
- **프로토콜**: gRPC over HTTP/2, TLS 1.3 암호화.
- **인증서**: 
  - 서비스 간 자체 서명 인증서 사용.
  - 경로: `internal/config/keys/prod/grpc-cert.pem`, `grpc-key.pem`.
- **설정**:
  - 상호 TLS(mTLS) 적용으로 클라이언트/서버 인증.
  - 구현: `internal/infrastructure/grpc/tls_config.go`.
- **키 관리**:
  - 갱신 주기: 6개월.
  - 생성:
    ```bash
    openssl req -x509 -newkey rsa:2048 -nodes -keyout internal/config/keys/prod/grpc-key.pem -out internal/config/keys/prod/grpc-cert.pem -days 180
    ```

---

## 3. 인증 및 권한 관리

### 3.1 JWT (JSON Web Token)
- **알고리즘**: RSA 2048비트.
- **구조**:
  - **Header**: `{"alg": "RS256", "typ": "JWT"}`.
  - **Payload**: 
    ```json
    {
      "sub": "client_id",        // 클라이언트/서비스 ID
      "iss": "tastesync-auth",   // 발급자
      "exp": 1648771200,         // 만료 시간 (1시간)
      "iat": 1648767600          // 발급 시간
    }
    ```
  - **Signature**: RSA 개인 키로 서명.
- **키 관리**:
  - 개발: `internal/config/keys/dev/private.pem`, `public.pem`.
  - 운영: `internal/config/keys/prod/private.pem`, `public.pem`.
  - 생성:
    ```bash
    openssl genrsa -out internal/config/keys/prod/private.pem 2048
    openssl rsa -in internal/config/keys/prod/private.pem -pubout -out internal/config/keys/prod/public.pem
    ```
- **구현**: 
  - 경로: `internal/auth/jwt.go`.
  - 토큰 검증: 모든 gRPC 요청에 포함, Envoy에서 사전 검증.

### 3.2 권한 관리
- **역할**:
  - `client`: 모바일/POS 앱 (고객 식별 요청).
  - `service`: 내부 서비스 (gRPC 호출).
  - `admin`: 관리 대시보드 (전체 접근).
- **정책**: 
  - 최소 권한 원칙(PoLP) 적용.
  - 역할별 접근 제어 리스트(ACL) 정의.
  - 파일: `internal/auth/acl.yaml`.
- **예시**:
  ```yaml
  roles:
    client:
      permissions:
      - "customer-id:identify"
    service:
      permissions:
      - "customer-id:identify"
      - "recommendation:recommend"
    admin:
      permissions: ["*"]
  ```

---

## 4. 데이터 보안

### 4.1 데이터 암호화
- **저장**: 
  - 민감 데이터(고객 ID, 위치): AES-256-GCM 암호화.
  - 키: `internal/config/keys/prod/data-encryption-key` (256비트).
  - 구현: `pkg/util/encryption.go`.
- **캐싱**: 
  - Redis에 암호화된 데이터 저장, TTL 1시간.
- **전송**: 
  - HTTPS/gRPC 통신 중 TLS로 보호.

### 4.2 데이터 접근 제어
- **PostgreSQL**: 
  - 최소 권한 계정(`tastesync`) 사용.
  - 암호: Kubernetes Secrets로 관리 (`customer-id-secrets`).
- **Redis**: 
  - ACL 설정으로 읽기/쓰기 분리.
  - 설정: `deploy/docker/docker-compose.yml`.

### 4.3 데이터 백업 및 복구
- **백업**: 
  - PostgreSQL: 매일 백업 → S3 (`s3://tastesync-backups/`).
  - 스크립트: `scripts/backup.sh`.
- **복구**: 
  - RPO: 24시간, RTO: 5분 목표.
  - 절차: `scripts/restore.sh`.

---

## 5. 취약점 관리

### 5.1 정기 점검
- **도구**: OWASP ZAP, Trivy (컨테이너 스캔).
- **주기**: 주 1회 자동 실행.
- **설정**: `tests/security_test.sh`.
- **실행**:
  ```bash
  make security-scan
  ```

### 5.2 취약점 대응
- **탐지**: 
  - Prometheus 경고 규칙으로 비정상 트래픽 감지.
  - 파일: `deploy/k8s/prometheus-config.yaml`.
- **대응**: 
  - CVSS 7.0 이상 취약점 발견 시 24시간 내 패치.
  - 인시던트 로그: `s3://tastesync-logs/security-incidents/`.
- **보고**: 팀 채널에 즉시 알림.

### 5.3 OWASP Top 10 대응
- **A01 - 접근 제어 오류**: JWT/ACL로 방지.
- **A03 - 인젝션**: 준비된 문장(Prepared Statement) 사용.
- **A07 - 인증 오류**: RSA JWT로 강력한 인증.

---

## 6. 보안 정책 구현

### 6.1 코드 수준
- **입력 검증**: 모든 gRPC 요청에서 필수 필드 확인.
- **에러 처리**: 민감 정보 노출 방지 (예: 스택 트레이스 숨김).
- **구현**: `internal/infrastructure/grpc/validate.go`.

### 6.2 배포 수준
- **컨테이너 보안**: 
  - 최소 이미지 사용 (`alpine` 기반).
  - 루트 권한 비활성화.
  - 파일: `deploy/docker/Dockerfile`.
- **네트워크 정책**: 
  - Kubernetes NetworkPolicy로 서비스 간 접근 제한.
  - 파일: `deploy/k8s/network-policy.yaml`.

### 6.3 모니터링 및 감사
- **로그**: 모든 요청/응답 로깅 (`internal/infrastructure/logging/`).
- **감사**: 월 1회 보안 감사 보고서 작성 (`docs/security-audit-<date>.md`).

---

## 7. 필요한 파일 및 경로
- **`internal/config/keys/dev/`, `prod/`**: JWT RSA 키.
- **`internal/infrastructure/grpc/tls_config.go`**: TLS 설정.
- **`internal/auth/jwt.go`**: JWT 인증.
- **`internal/auth/acl.yaml`**: 권한 제어 정책.
- **`pkg/util/encryption.go`**: 데이터 암호화.
- **`deploy/k8s/customer-id.yaml`**: 배포 설정.
- **`deploy/k8s/network-policy.yaml`**: 네트워크 정책.
- **`tests/security_test.sh`**: 취약점 점검 스크립트.
- **`scripts/backup.sh`, `restore.sh`**: 백업/복구 스크립트.

---

## 8. 결론
TasteSync의 보안 정책은 `customer-id`를 중심으로 데이터, 통신, 인증을 철저히 보호하며, OWASP 기준과 규제 준수를 충족합니다. 이 정책을 엄격히 준수하여 1,000만 사용자 규모에서도 안전한 서비스를 제공하세요.