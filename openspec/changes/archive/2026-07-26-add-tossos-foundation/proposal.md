# Change: add-tossos-foundation

## Why

TossOS는 tossinvest-cli를 코어로 삼은 독립 product fork로 자동매매 제품을 만든다. 자동매매 엔진을 얹기 전에 fork·빌드·테스트 베이스라인을 고정하고, SDD(OpenSpec) 작업 체계와 역할 분리(Manager/Teammate) 규칙을 저장소에 확립해야 한다. 이 기반 없이는 이후 전략 손실과 주문 시스템 결함을 분리할 수 없고, 다중 에이전트 작업이 세션 간 일관성을 잃는다.

## What Changes

- upstream(JungHoonGhae/tossinvest-cli) 전체 히스토리를 보존한 fork 확립: 커밋 `57348a7` 고정, `upstream` remote 설정, 작업 브랜치 `feat/p0-foundation`
- `go build`/`go vet`/`go test` 베이스라인 기록(docs/baseline.md): 650개 테스트 green, 패키지별 커버리지, upstream 알려진 갭 목록
- 전체 로드맵(docs/ROADMAP.md)과 SDD 작업 규칙(docs/WORKFLOW.md) 추가 — Phase 0~6, change 7건으로 분할
- Makefile에 표준 타겟 보강(test/vet/cover/validate — 기존 upstream 타겟 보존), .gitignore 시크릿·로컬 산출물 보강
- StockOS 재사용 자산 인벤토리(docs/stockos-inventory.md) 추가
- 기존 Go 코드는 무변경 (베이스라인 오염 방지)

## Capabilities

### New Capabilities

- `product-fork`: upstream 히스토리 보존, 베이스라인 고정, 회귀 금지, 선별적 upstream 동기화, 라이선스·시크릿 정책
- `sdd-workflow`: OpenSpec 기반 변경 관리, Manager/Teammate 역할 분리, 문서 gstack 리뷰 강제, TDD, 완료 검증 게이트, 실계좌 보호

### Modified Capabilities

(없음 — 최초 change)

## Impact

- Affected code: 신규 파일만 추가(openspec/, docs/, Makefile 타겟, .gitignore 항목). 기존 Go 코드 무변경
- 이후 모든 Phase(1~6)는 이 change가 확립한 게이트(전체 테스트 통과, `openspec validate --strict`, gstack 리뷰)를 따른다
- Non-goals: 자동매매 기능 추가 없음, Go module 경로 변경 없음(Phase 2 이후 별도 change에서 재평가), GitHub 비공개 origin 생성·push 없음(사용자 수행)
