# TasteSync Contributing Guide

TasteSync는 소상공인 실시간 고객 경험(CX) 최적화 플랫폼으로, 오픈소스 프로젝트로서 커뮤니티 기여를 환영합니다. 이 문서는 `tastesync-customer-id` 서비스를 중심으로 기여 방법을 안내하며, 2027년 v1.0 출시를 목표로 500~1,000 사용자(식당 중심)를 지원합니다. 코드, 문서, 버그 보고 등 모든 기여를 통해 프로젝트를 함께 발전시켜 주세요.

---

## 1. 기여 개요

### 1.1 목표
- **품질**: 고품질 코드와 문서로 프로젝트 안정성 유지.
- **협업**: 명확한 프로세스로 커뮤니티 참여 촉진.
- **일관성**: TasteSync의 설계 원칙(DDD, 헥사고날, 이벤트 드리븐) 준수.

### 1.2 기여 유형
- **코드**: 기능 추가, 버그 수정.
- **문서**: README, 가이드 개선.
- **버그 보고**: 이슈 제안 및 재현 방법 제공.
- **피드백**: 아이디어 제안, 사용자 경험 개선.

---

## 2. 시작하기

### 2.1 필수 준비
- **Git**: 버전 관리 도구.
- **Go 1.23**: 개발 환경 (`docs/development-guide.md` 참조).
- **GitHub 계정**: 기여를 위한 계정 생성.

### 2.2 리포지토리 설정
1. **Fork**:
   - `https://github.com/yourusername/tastesync-customer-id`를 개인 계정으로 Fork.
2. **클론**:
   ```bash
   git clone https://github.com/<your-username>/tastesync-customer-id.git
   cd tastesync-customer-id
   ```
3. **원격 설정**:
   ```bash
   git remote add upstream https://github.com/yourusername/tastesync-customer-id.git
   ```

---

## 3. 기여 프로세스

### 3.1 이슈 제안
- **경로**: GitHub Issues (`https://github.com/yourusername/tastesync-customer-id/issues`).
- **템플릿**:
  ```
  **제목**: [BUG] IdentifyCustomer에서 UUID 검증 실패
  **설명**: 특정 UUID 형식에서 오류 발생.
  **재현 방법**:
  1. `IdentifyCustomer` 호출
  2. `uuid: "invalid-uuid"` 입력
  **예상 결과**: 오류 메시지 반환
  **실제 결과**: 서버 충돌
  **환경**: Go 1.23, Docker
  ```
- **주의**: 중복 이슈 확인 후 제출.

### 3.2 코드 기여
1. **브랜치 생성**:
   - 타입: `feature/`, `bugfix/`, `docs/` (예: `feature/add-logging`).
   ```bash
   git checkout -b feature/<your-feature>
   ```
2. **코드 작성**:
   - 스타일: `docs/development-guide.md`의 K8s 스타일 준수.
   - 테스트: 단위/통합 테스트 추가 (`make test-unit`).
3. **커밋**:
   - 형식: `<type>(<scope>): <description>` (예: `feat(grpc): add IdentifyCustomer timeout`).
   - 타입: `feat`, `fix`, `docs`, `test`, `chore`.
4. **푸시**:
   ```bash
   git push origin feature/<your-feature>
   ```
5. **Pull Request (PR)**:
   - 대상: `main` 브랜치.
   - 템플릿:
     ```
     **제목**: feat(grpc): add IdentifyCustomer timeout
     **설명**: 요청 타임아웃을 5초로 설정.
     **관련 이슈**: #123
     **변경 사항**:
     - `internal/infrastructure/grpc/server.go`에 타임아웃 추가.
     **테스트**:
     - `tests/unit/server_test.go`에서 타임아웃 검증.
     ```
   - 요구 사항: `make lint`, `make test-unit` 통과.

### 3.3 문서 기여
- **대상**: `docs/` 내 파일 (예: `README.md`, `architecture.md`).
- **방법**: 코드 기여와 동일한 프로세스, PR 제목에 `docs:` 접두사 사용 (예: `docs(config): update config guide`).

---

## 4. 행동 강령 (Code of Conduct)

### 4.1 기본 원칙
- **존중**: 모든 참여자에게 예의와 배려.
- **포용성**: 다양한 배경과 의견 환영.
- **협력**: 건설적인 피드백과 팀워크 강조.

### 4.2 금지 사항
- 모욕, 차별, 괴롭힘 등 부정적 행동.
- 위반 시: PR/이슈 차단 및 커뮤니티 제외.

---

## 5. 코드 품질 기준

### 5.1 코드 스타일
- **K8s 스타일**: `docs/development-guide.md`의 2.2절 준수.
  - 변수: `lowerCamelCase` (예: `customerID`).
  - 함수: `UpperCamelCase` (예: `IdentifyCustomer`).
  - 주석: GoDoc 형식 필수.

### 5.2 테스트 요구사항
- **단위 테스트**: 새 기능/수정 시 100% 커버리지 목표.
- **통합 테스트**: gRPC 호출 및 의존성 검증.
- **실행**: `make test-unit`, `make test-integration`.

### 5.3 보안
- **입력 검증**: 모든 gRPC 요청에서 필수 필드 체크.
- **에러 처리**: 민감 정보 노출 금지.
- **참조**: `docs/security-policy.md`.

---

## 6. 리뷰 및 병합 프로세스

### 6.1 리뷰 요구사항
- **최소 리뷰어**: 1명 (핵심 유지관리자).
- **체크리스트**:
  - 코드 스타일 준수.
  - 테스트 통과.
  - 보안 검토 완료.
- **승인**: 리뷰어 "Approve" 후 병합 가능.

### 6.2 병합
- **방법**: squash merge로 커밋 히스토리 간소화.
- **담당**: PR 작성자 또는 유지관리자.

---

## 7. 기여 팁

### 7.1 시작하기 좋은 작업
- **이슈 탐색**: "good first issue" 라벨 확인.
- **문서 개선**: `docs/` 내 오타 수정, 설명 추가.

### 7.2 도구 활용
- **Lint**: `make lint`로 품질 점검.
- **Debug**: `logging.level: debug`로 디버깅 로그 활성화.

### 7.3 질문 및 지원
- **채널**: GitHub Discussions 또는 팀 Slack (추후 지정).
- **멘토링**: 신규 기여자는 PR에 "mentor needed" 코멘트 추가.

---

## 8. 필요한 파일 및 경로
- **`docs/development-guide.md`**: 코드 스타일/테스트 가이드.
- **`docs/security-policy.md`**: 보안 기준.
- **`.golangci.yml`**: 린팅 설정.
- **`Makefile`**: 빌드/테스트 명령어.
- **`.github/PULL_REQUEST_TEMPLATE.md`**: PR 템플릿 (별도 생성 권장).

---

## 9. 결론
TasteSync는 커뮤니티의 기여를 통해 초저지연과 확장성을 갖춘 플랫폼으로 성장합니다. 이 가이드를 따라 코드, 문서, 피드백 등 다양한 방식으로 참여해 주세요. 질문이 있으면 언제든 Issues나 Discussions에서 문의하세요!
