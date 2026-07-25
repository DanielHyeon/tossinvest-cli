# Design: add-tossos-foundation

## Context

TossOS는 tossinvest-cli(커밋 `57348a7`, 650 테스트 green)의 product fork다. 이 change는 코드가 아니라 "작업이 굴러가는 기반"을 만든다. 정찰 결과 upstream은 표준 라이브러리 중심(외부 의존: cobra + charmbracelet뿐)이고, 주문 정책(`internal/trading`)·의도 정규화(`internal/orderintent`)·브로커 3종(hybrid/official/WTS)이 이미 인터페이스로 분리돼 있어 엔진이 직접 호출하기 좋은 구조다. 반면 배선(`newAppContext`)과 조건주문 게이트는 `cmd/tossctl`에 갇혀 있고, 위험 한도·주문 상태기계·SSE→상태 반영은 코드에 존재하지 않는다 — 이는 Phase 1~2의 입력이다.

## Goals / Non-Goals

**Goals**
- fork·베이스라인 고정(회귀 기준점), SDD 체계·역할 분리 규칙의 저장소 고정
- 이후 Phase가 따를 완료 게이트 확립

**Non-Goals**
- 자동매매 기능·기존 Go 코드 수정 (베이스라인 오염 방지)
- Go module 경로 변경, GitHub origin 생성·push

## Decisions

### D1. Go module 경로를 P0에서 변경하지 않는다
대안 (a) 즉시 `module tossos` rename + 전체 import 수정 / (b) 유지. **선택: (b)**.
rename은 전 파일 diff를 오염시키고 upstream 선별 merge 충돌 표면을 최대화한다. 신규 패키지는 기존 module 경로 하위 `internal/`에 추가하므로 rename 없이 개발 가능. Phase 2 이후 별도 change에서 재평가.

### D2. main은 upstream 추적, 작업은 feat/p&lt;N&gt;-* 브랜치
upstream 선별 merge를 main에서 받아 충돌을 한 곳에서 해소하고, Phase 작업은 브랜치에서 검증 후 main에 반영. 비공개 origin은 사용자가 직접 생성·연결(에이전트 push 금지).

### D3. SDD 게이트를 저장소 문서로 강제
다중 에이전트 체계에서 규칙이 대화 컨텍스트에만 있으면 세션 간 소실된다. docs/WORKFLOW.md + sdd-workflow spec으로 저장소에 고정하고, 모든 change 완료 게이트에 (1) 전체 테스트 통과, (2) `openspec validate --strict`, (3) gstack 리뷰를 포함한다.

### D4. CLI·MCP를 포함한 upstream 전체 보존 (절단 없음)
정찰 보고는 `internal/output`(4,600 LOC)·`tui`·`i18n` 등을 "버리는" 절단선을 제안했으나, 제품 결정은 **CLI를 관리자 콘솔로 보존**하는 것이다. 엔진은 해당 패키지를 사용하지 않을 뿐 삭제하지 않는다. 삭제는 upstream merge 충돌만 늘린다.

### D5. NTFS 환경 대응
저장소가 NTFS(fuseblk) 마운트에 있어 `core.filemode=false`가 필수다(이미 설정됨). 실행 비트가 보존되지 않으므로 스크립트 실행은 `bash script.sh` 형태를 표준으로 한다. **영속 데이터(원장·intent journal)는 SQLite 내구성 문제 때문에 저장소 밖 ext4 경로(XDG data dir)에 두고, fuseblk 경로는 기동 시 거부한다** (gstack 리뷰 반영).

### D6. Fork 자세: 선별 cherry-pick + shim (gstack 리뷰 반영)
리뷰가 D1/D4(머지 용이성)와 Phase 1 리팩터(newAppContext·MutationResult·조건주문 게이트 이동)의 모순을 지적했다. 결정: **선별 cherry-pick 중심의 fork 자세**를 명시하되, Phase 1 리팩터는 **위임 shim**(cmd/tossctl에 동명 얇은 래퍼, `type MutationResult = domain.MutationResult` alias)으로 수행해 upstream 파일 원형과 기존 테스트를 보존한다. upstream 반영은 `upstream-sync` 브랜치 전용. Go module 경로 변경은 여전히 보류 — 사용자 결정 항목(review.md 미결정 사항).

### D7. 리뷰 게이트 등급제와 기록 artifact (gstack 리뷰 반영)
"모든 문서 커밋 전 리뷰"는 실행 불가능해 조용히 무시된다는 4보이스 합의에 따라, 게이트를 proposal-freeze와 Requirement 수준 수정 2지점으로 한정하고 `openspec/changes/<id>/review.md`를 필수 기록으로 표준화했다. 완료 게이트는 tools/gate.sh로 자동화한다. 상세는 docs/WORKFLOW.md.

## Risks / Trade-offs

- [upstream 갭: amend 상태 판별·매도 경계 미검증] → Phase 1에서 계약 테스트로 보강, 실계좌 검증은 사용자 승인 수동 프로토콜로 분리
- [WTS 세션 연장은 폰 승인 필수 → 무인 운영 중단 위험] → Phase 2+ 설계 원칙: 핵심 매매 루프는 공식 API만으로 동작, WTS 의존 기능(수급·스크리너·AI시그널)은 성능 저하 모드로 강등
- [App-Version 웹 번들 스크레이핑(WTS 주문 선행 단계)의 취약성] → WTS는 조회 전용으로만 사용하는 권위 원칙이 이 리스크를 주문 경로에서 제거
- [openspec 1.4.1 규격 불일치 가능성] → `validate --strict`로 즉시 검출

## Migration Plan

신규 파일 추가만 있으므로 롤백은 커밋 revert로 충분. 기존 Go 코드 무변경이 곧 롤백 안전성이다.

## Open Questions

- 비공개 origin 생성 시점과 저장소 이름 (사용자 결정 대기)
- Phase 5 웹 UI의 데이터 페칭 전략: StockOS는 React Query 미사용(수동 nonce 갱신) — TanStack Query 도입 여부는 Phase 5 change에서 결정
