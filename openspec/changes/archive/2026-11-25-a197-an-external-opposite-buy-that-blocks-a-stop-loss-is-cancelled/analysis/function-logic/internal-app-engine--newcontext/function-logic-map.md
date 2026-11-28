# Function Logic Map: `NewContext`

- Source: `internal/app/engine/engine.go` (**편집 전** `452`–`640`, base `b93ed133a`)
- Qualified: `NewContext`
- AST evidence: `ast.json` — branches 18 · returns 12
- Risk scan: `risk-pattern-report.md`
- **편집 뒤 갱신(a197 3.9, 2026-11-25)**: `ast.json` · 위험 보고 재생성, 이 문서의 이 파일 좌표는 base `b93ed133a` → 작업 트리로 difflib 줄 정렬해 옮김(다른 파일 좌표 무변), 분기 번호는 위치라 새 분기 뒤를 재번호함. B6 `:496` 새 분기(토글 audit, 실패면 기동 거절), 옛 B6~B17 → B7~B18. a166 census `shapes.json` 의 이 키 갱신(Manager 승인, analysis/census-repair).
- 작성 시점: **편집 전**(a197 재개 개정 2026-11-25, Manager 결정 O-1 = (나) — 토글 audit 을 기동 audit 단계에 둔다). 분기 주장은 `ast.json` 열거에서 왔다.
- **a166 census 고정 함수**: `internal/accountbudgetruntime/testdata/a166-wiring/shapes.json` 의 `internal/app/engine/engine.go||NewContext`(렌더 본문 고정). 편집하면 census 갱신이 필요하다 — **Manager 승인(2026-11-25, a199 선례)**. 선택 시험에 `./internal/accountbudgetruntime/...` 필수.

**역할.** 엔진 Context 조립: 주문 경로 → audit log → 시계 → automation 상태 · 알림 해석 → **기동 설정 audit(`recordGateSettings`)** → 계좌 해석 → 원장 → apply hook → 예산 → 게이트웨이 → 보호 manifest → Guardian → interlock → 청산 전용 판정 → 전략 projection → Context 반환. 생산 호출자: `cmd/tossctl/engine_assembly.go:53`(`assembleEngine`), `engine.go:441`(`New`).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` · `opts` | 조립 옵션 | `assembleEngine` · `New` | 단계별 오류 반환 또는 `refuseStartup` |
| `cfg.Engine.*` | 로드된 설정 | `NewOrderPath` → `config` | — |
| 불변식 | 설정 audit 은 어떤 거절보다 **먼저** 쓴다(§0.5 — 주석 `:482-486`); audit 실패 = 기동 거절 | 이 함수 | B5 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:454` `NewOrderPath` 오류 | 없음 | `return@455` | 조립 실패 시험(미특정) |
| B2 | `:461` `openAuditLog` 오류 | 없음 | `return@462` | 미특정 |
| B3 | `:465` `clk == nil` | `clock.System()` | — | 미특정 |
| B4 | `:479` `opts.Publisher != nil` | 알림 발행기 주입 | — | 미특정 |
| B5 | `:487` `recordGateSettings(…)` 오류 — **a197 의 토글 audit 이 붙는 자리** | audit 행 기록 | `return@492` 기동 거절(오류 감쌈) | a074 · a163 audit 시험(함수 단위), 기동 시험(미특정) |
| B6 | `:496` a197 토글 audit `RecordChange` 오류 (**a197 새 분기**, B5 바로 뒤) | audit 행 기록(값이 바뀐 기동만) | `return@497` 기동 거절(B5 와 같은 문장) | `TestA197StartupAuditsTheLoadedExternalBuyToggle` |
| B7 | `:508` 계좌 해석 오류 | 거절 기록 | `return@509` `refuseStartup` | 미특정 |
| B8 | `:515` 원장 열기 오류 | 거절 기록 | `return@516` | 미특정 |
| B9 | `:524` apply hook 결합 오류 | 원장 닫기 | `return@526` | 미특정 |
| B10 | `:531` `opts.BudgetFactory != nil` | 예산 생성 | — | a166 |
| B11 | `:533` 예산 오류 | 원장 닫기 | `return@535` | a166 |
| B12 | `:552` 보호 manifest 오류 | 원장 닫기 | `return@554` | 미특정 |
| B13 | `:567` gate 켜짐 ∧ guardian nil ∧ 생산 Guardian 허용 | Guardian 생성 | — | 미특정 |
| B14 | `:569` `factory == nil` | 기본 factory | — | 미특정 |
| B15 | `:575` Guardian 생성 오류 | 원장 닫기 | `return@577` | 미특정 |
| B16 | `:597` interlock · 주문 증명 오류 | 원장 닫기 | `return@599` | interlock 시험 |
| B17 | `:601` `!automation.RunsExitSide()` | Guardian 비움(a199) | — | a199 시험 |
| B18 | `:606` 전략 projection 오류 | 원장 닫기 | `return@608` | 미특정 |

정상 반환 `return@610`(Context 리터럴).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `recordGateSettings` | 기동 설정 audit(값이 바뀐 기동만 행) | 오류 → B5 기동 거절 | AST `:487` |
| `refuseStartup` | 거절 audit · 오류 | — | AST `:509` 등 |
| `buildGateway` · `runInterlock` · `engineOrderProof` 등 | 조립 | 각자 | AST |

Live config binding: `cfg.Engine.ExitSourceAgeReport` 를 audit 인자로 넘김(`:488`) — a197 토글도 같은 방식으로 `cfg.Engine.<새 칸>` 을 읽는다.

## State mutations and fallbacks

- audit log 쓰기(B5), 원장 열기 · 닫기, 게이트웨이 · Guardian 생성. 브로커 mutation 0.

## Safety conclusion

- Safe edit boundary (a197 D8): B5 의 `recordGateSettings` 호출 **바로 뒤**에 같은 실패 의미의 audit 한 줄 — `if _, err := auditLog.RecordChange(<새 동작 상수>, "engine.cancel_external_opposite_buy", strconv.FormatBool(cfg.Engine.<새 칸>), ""); err != nil { return nil, fmt.Errorf("engine: recording the automation-gate audit trail: %w", err) }` 모양. 기존 분기 · 순서 무변경, 거절보다 먼저(§0.5) 유지. `recordGateSettings` 본문 · 인자는 바꾸지 않는다(a163 핀 — 마지막 행 고정, a074 순서 핀, 시험 함수 5 개 호출 무편집).
- census: 렌더 본문이 바뀌므로 `shapes.json` 의 이 항목 갱신(Manager 승인). `sites.json` 재고의 선택자(`Do` · `RoundTrip` · `NewClient` · `Budget` · `official.New*`)는 새 줄이 쓰지 않는다.
- High-risk impact: yes(기동 거절 조건에 audit 행 하나가 더해짐 — audit log 쓰기 실패 시 기동 거절, 기존 의미와 같음).
