# a109 — 형제 endpoint들도 회복한다

> **상태: 착수(2026-08-15).** a108의 A1 적대 리뷰가 실측으로 연 change다. a108 land
> (3615f793) 확인 후 base를 016da624로 재고정했고, a108이 strategy projection에 세운
> 계약(stage+rename 발행, connect probe 생존 판정, 회수의 전체성)을 같은 의례로
> 이식한다. 설계는 design.md D1–D5.

## Why

a108은 2026-08-13 재부팅 사고의 endpoint(strategy projection)만 고쳤다. A1 리뷰가 나머지
세 엔진 endpoint(position policy command·position policy runtime·alert control)를 실측한
결과, **같은 병이 표면만 다르게 존재한다**:

- **pre-chmod socket 잔재의 영구 거부** — policy runtime과 alert control은
  `PreparePrivateSocket`/`PrepareRuntimeSocket`이 정확-0600이 아닌 socket을 거부한다.
  listen→chmod 사이의 죽음이 umask 077(컨테이너 실측)에서 0700 socket을 남기면 매 부팅이
  거부되고, 세 endpoint 모두 `cmd/tossctl/engine.go`(:272/:277/:294 부근)에서 **fatal**이라
  a108 이전의 strategy projection과 똑같이 엔진 기동 루프 전체가 죽는다. A1이 각 3회
  반복으로 재현했다.
- **alert control은 살아 있는 주인의 socket 위에 두 번째 서버가 올라선다** — strategy
  projection이 a108에서 세운 규칙(수락 중인 socket은 건드리지 않는다)과 정반대다.
  실제 방어는 journal flock뿐인데 코드도 문서도 그것을 근거로 들지 않는다.

a108의 tasks 2.5가 이것을 잡지 못한 이유도 기록돼 있다: 핀이 관용이 자명한 잔재 모양만
만들어 **실패할 수 없었다**(a108 review.md §1 F4). 사고급 모양(pre-chmod socket·산 주인)의
핀은 이 change가 소유한다.

## What Changes

- 두 socket endpoint의 발행을 a108과 같은 stage+rename 의례로 통일한다(pre-chmod
  상태의 원천 소거). 형제 staging basename은 11자(`alerts.sock` 11자 이하) — design D1a.
- 잔재 회수를 a108의 전체성 기준으로 맞춘다 — 검증-사망 잔재의 좁은 perm 완화,
  staging 잔재(구버전 CreateTemp 이름 포함) 회수, 낯선 엔트리 거부 포함.
- 산-주인 탈취를 **거부로 바꾼다**(두 socket endpoint 공통 — connect probe 사망 증명
  없이 unlink 금지). journal flock이 1차 방어임은 코드 주석이 명시 인용한다(design D2).
- `engine.go`의 세 fatal을 강등한다 — endpoint별 보수성 판정은 design D3 표,
  보고는 a108 D3-2 교리(critical·등급표·outbox 3금지)를 일반화한다.
- httpapi 소비자측 lazy 재-dial(a108 gstack 이관 2b.1)과 defer 순서 핀(2b.2),
  롤백 절차 측정·기록(2b.3)을 함께 닫는다.
- 사고급 모양의 crash-shape 핀(pre-chmod socket·산 주인·staging 잔재·낯선 엔트리)을
  깐다 — a108 2.5의 실패-불가 핀을 대체하되 기존 관용 핀은 유지한다.

## Non-goals

- strategy projection endpoint의 재작성 — a108이 소유한다(staging 12자 유지 포함).
- descriptor 발행 3벌의 fold — 병이 없는 표면의 High-risk 리팩터링(선언된 생략,
  design 말미).
- (등록 시점의 유보였던 강등 판정은 설계에서 완료했다 — design D3. 등록문이 우려한
  "명령 표면" 논증은 표면 부재의 보수성 논증으로 대체됐다.)

## Impact

- Affected specs: `engine-safety` (ADDED 2 — 일반 불변식 + 기동 생존),
  `http-api-service` (ADDED 1 — 전략 화면 재부착)
- Affected code: `internal/positionpolicyrpc`, `internal/app/engine`의 세 transport,
  `cmd/tossctl/engine.go`·`cmd/tossctl/httpapi.go`
- High-risk: 엔진 기동 경로 — FLM AST 13개 생성 완료, 맵 완성·Pre-Edit는 각 Teammate
  편집 전.
