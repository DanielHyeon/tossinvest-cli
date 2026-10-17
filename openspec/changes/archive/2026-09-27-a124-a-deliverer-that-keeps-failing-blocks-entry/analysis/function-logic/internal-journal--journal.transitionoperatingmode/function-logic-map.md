# Function Logic Map: `Journal.TransitionOperatingMode`

- Source: `internal/journal/operating_mode.go`
- AST evidence: `ast.json` — **편집 전**, :346–485, 분기 28 · 반환 20.
  source_sha256 `ed35ea67c26b…`, 추출 base `4798d399` (2026-09-26, Teammate — 재freeze 로트).
- Risk scan: `risk-pattern-report.md`

**a124 는 이 함수를 편집하지 않는다.** design 3판 D1 · D7 이 「같은 모드면 행을 안 쓴다 · 더 엄하면 그대로 · 투영은 커밋 뒤 ·
오류 문구에 계좌가 들어간다」를 근거로 쓰기 때문에 문서보다 먼저 만들었다. `EscalateOperatingMode`(:496-511)는 이 함수를
`Actor: auto` 로 부르는 얇은 포장이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `req.AccountRef` · `Mode` · `Actor` · `Cause` | 비어 있지 않음·열거 안 | 호출자 | B2~B5 |
| 자동 요청 | HALT_ALL·NORMAL 금지, 열거 트리거만 | 정본 | B8~B10 |
| 현재 모드 | 원장 최신 행 | `currentModeTx` 트랜잭션 안 | B12 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 요청 검증 switch (:353) | — | — | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B2 | 계좌 공백 (:354) | 없음 | ErrInvalidRequest (:355) | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B3 | 모드 무효 (:357) | 없음 | ErrInvalidRequest (:358) | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B4 | 행위자 무효 (:360) | 없음 | ErrInvalidRequest (:361) | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B5 | 원인 공백 (:364) | 없음 | ErrInvalidRequest (:365) | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B6 | `actor == auto` (:371) | — | — | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B7 | 자동의 목표 모드 switch (:372) | — | — | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B8 | 자동 → HALT_ALL (:373) | 없음 | ErrHaltAllIsNeverAutomatic (:374) | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B9 | 자동 → NORMAL (:375) | 없음 | ErrModeRelaxationRequiresOperator (:378) | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B10 | 자동인데 열거된 트리거 아님 (:381) | 없음 | ErrModeTriggerUnknown (:382) | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B11 | `BeginTx`(IMMEDIATE) 오류 (:392) | 없음 | 오류 — **문구에 계좌 포함** (:393-394) | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B12 | 현재 모드 읽기 오류 (:399) | 롤백 | 오류 (:400) | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B13 | 방향 판정 오류 (:406) | 롤백 | 오류 (:407) | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B14 | 방향 switch (:409) | — | — | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B15 | `direction == 0` — 이미 그 모드 (:410) | 행 안 씀 | `changed=false`, nil (:415) | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B16 | 더 엄한 모드 + 자동 (:417) | 행 안 씀 | `changed=false`, nil (:421) | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B17 | 완화 (:423) | — | — | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B18 | 완화인데 승인 없음 (:424) | 없음 | ErrModeApprovalRequired (:425) | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B19 | 완화인데 감사 기록기 없음 (:428) | 없음 | ErrInvalidRequest (:429) | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B20 | 행 수 읽기 오류 (:437) | 롤백 | 오류 (:438) | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B21 | id 공백 → 생성 (:441) | — | — | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B22 | INSERT 오류 (:444) | 롤백 | 오류 — 계좌 포함 (:448-449) | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B23 | 감사 기록기 있음 (:460) | 감사 줄 | — | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B24 | 감사 기록 오류 (:461) | 롤백 | 오류 (:463) | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B25 | 커밋 오류 (:468) | 롤백 | 오류 (:469) | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B26 | 투영기 있음 (:475) | **커밋 뒤** `ProjectOperatingMode` — 게이트 투영 | — | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B27 | announcer 있음 (:478) | 커밋 뒤 알림 | — | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |
| B28 | 알림 오류 (:479) | 전이는 이미 커밋 | `ErrModeAnnouncementFailed` (:480) | 기존 operating_mode 시험 (1.4 — 이 change 는 채우지 않는다) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.db.BeginTx` :390 | BEGIN IMMEDIATE — 원장 연결 하나 | 오류 → B11 | AST |
| `currentModeTx` · `modeDirection` | 트랜잭션 안의 현재 모드와 방향 | 오류 → B12 · B13 | AST |
| `j.modeProjectorRef().ProjectOperatingMode` :476 | 게이트 투영 | **커밋 뒤** (B25 다음) | AST |
| `req.Announcer.AnnounceOperatingMode` :479 | 알림 | 커밋 뒤, 오류는 전이를 되돌리지 않음 | AST |

## State mutations and fallbacks

- 자동 승격의 반복 호출은 B15(같은 모드) · B16(더 엄함)에서 **행을 쓰지 않는다** — 매 사이클 재승격은 폭풍이 아니다.
- 투영(B26)과 알림(B27)은 커밋 **뒤**다 — 원장 트랜잭션을 쥔 채 게이트 뮤텍스나 `Notify` 로 가지 않는다(교착 논증의 근거).
- 오류 문구 B11 · B22 는 계좌를 담는다 — 그 오류를 게이트 detail 로 옮기면 안 된다(design D9).

## Safety conclusion

- Safe edit boundary: 편집하지 않는다.
- High-risk impact: yes — 운영 모드(읽기 전용 근거).
