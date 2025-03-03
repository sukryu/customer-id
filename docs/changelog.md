# TasteSync Changelog

이 문서는 TasteSync 프로젝트의 변경 사항을 기록합니다. `tastesync-customer-id` 서비스를 중심으로, 2027년 v1.0 출시를 목표로 하며, 기능 추가, 버그 수정, 문서 개선 등을 버전별로 정리합니다. 모든 변경은 Semantic Versioning(SemVer)을 따�며, `MAJOR.MINOR.PATCH` 형식으로 버전이 증가합니다.

---

## [Unreleased]
- **날짜**: TBD
- **설명**: 현재 개발 중인 변경 사항.

### Added
- 초기 프로젝트 구조 설정 (`cmd/`, `internal/`, `pkg/` 등).
- gRPC 기반 `IdentifyCustomer` 메서드 정의 (`proto/customer_id.proto`).
- DDD, 헥사고날, 이벤트 드리븐 아키텍처 설계 (`docs/architecture.md`).
- 환경 변수 관리 (`internal/config/config.yaml`, `config.go`).
- JWT RSA 인증 초기 설정 (`internal/auth/jwt.go`).

### Changed
- N/A (초기 설정 단계).

### Fixed
- N/A (초기 설정 단계).

### Removed
- N/A (초기 설정 단계).

---

## [0.1.0] - 2025-03-02
- **설명**: 초기 개발 환경 및 문서화 작업 완료.

### Added
- 프로젝트 초기화 및 GitHub 리포지토리 생성 (`tastesync-customer-id`).
- Go 1.23 기반 개발 환경 설정 (`go.mod`, `.golangci.yml`).
- Docker 및 Kubernetes 배포 설정 (`deploy/docker/Dockerfile`, `deploy/k8s/customer-id.yaml`).
- CI/CD 파이프라인 초기 설정 (`.github/workflows/build.yml`, `deploy.yml`).
- `Makefile`으로 빌드/테스트/배포 자동화 스크립트 추가.
- 문서화 초기 작업:
  - `docs/README.md`: 프로젝트 개요.
  - `docs/architecture.md`: 아키텍처 개요.
  - `docs/development-guide.md`: 개발 가이드.
  - `docs/api-spec.md`: API 정의.
  - `docs/deployment-guide.md`: 배포 가이드.
  - `docs/security-policy.md`: 보안 정책.
  - `docs/event-model.md`: 이벤트 모델.
  - `docs/config-guide.md`: 설정 가이드.
  - `docs/contributing.md`: 기여 가이드.
  - `docs/monitoring-logging.md`: 모니터링 및 로깅 전략.

### Changed
- N/A (최초 릴리스).

### Fixed
- N/A (최초 릴리스).

### Removed
- N/A (최초 릴리스).

---

## Notes
- **버전 관리**: SemVer를 준수하며, `[Unreleased]` 섹션은 개발 중인 변경 사항을 추적합니다.
- **날짜**: 실제 배포 시 ISO 8601 형식 (YYYY-MM-DD)으로 기록.
- **형식**: 각 버전은 `Added`, `Changed`, `Fixed`, `Removed` 하위 섹션으로 구성.

---

## Contributing
변경 사항을 제안하거나 추가하려면 `docs/contributing.md`를 참고하여 PR을 제출하세요. 모든 기여는 이 changelog에 기록됩니다.