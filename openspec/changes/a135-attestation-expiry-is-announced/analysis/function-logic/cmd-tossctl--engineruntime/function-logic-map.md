# Function Logic Map: `engineRuntime`

- Source: `cmd/tossctl/engine.go` (655-748, 편집 뒤 — base `421bf53c` 653-745; 2줄 이동은 a134 `85752cd6` 의 `newEngineCmd` 편집)
- Qualified: `engineRuntime`
- AST evidence: `ast.json`(편집 뒤 재추출, `source_sha256` 458cde7b2cbdef09… · base 446e27a347768b0b…) — AST branches 6 · returns 7(편집 전후 동일)
- Risk scan: `risk-pattern-report.md`(configured risk pattern 0)

**역할.** 엔진 런타임을 조립함 — 감독 루프 넷(reconcile · exit · filldetect · strategy-entry) + 보조 실행자. a135 편집은
`Auxiliary` 리터럴에 `ectx.AttestationWatch(clk)` 한 원소(+ 주석 한 줄)를 더한 것뿐. 분기 · early return · 오류 경로 무변경 —
`AttestationWatch` 는 오류를 돌려주지 않는 생성자라 새 분기가 없음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ectx` | 생성된 `*engine.Context`(게이트 판정 · 알림기 · 저널 포함) | `engine.NewContext` | 각 생성자 오류에서 그대로 반환(B1~B6) |
| `clk` | 시계(생산 = system) | 호출자 | 감시 실행자도 같은 시계로 잠듦 |

## Branches and early returns

| Branch | AST line(편집 뒤) | Condition | Return/error | a135 |
|---|---|---|---|---|
| B1 | 658 | 체결 감지 생성 실패 | `nil, err` | 무편집 |
| B2 | 667 | 대사 드라이버 생성 실패 | `nil, err` | 무편집 |
| B3 | 682 | exit 관측자 생성 실패 | `nil, err` | 무편집 |
| B4 | 687 | 복구 생성 실패 | `nil, err` | 무편집 |
| B5 | 691 | 전략 진입 외곽 생성 실패 | `nil, err` | 무편집 |
| B6 | 703 | 배달 실행자 생성 실패 | `nil, err` | 무편집 |

Exact AST return positions: 편집 전후 7개, 위치는 2줄 이동(a134) 외 동일. 끝 return(`engine.NewRuntime(...)`)의 인자 리터럴만 a135 가 바꿈.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ectx.AttestationWatch(clk)`(a135 추가) | attestation 만료 · 범위 상실 감시 보조 실행자 | 오류 없음, 생성만. 실행 시 1시간마다 로컬 파일 1회 읽기 · 브로커 호출 0 · `OnStop` 없음 | `internal/app/engine/attestation_watch.go` · AST calls |
| `engine.NewRuntime` | 런타임 조립 · 보조 실행자 이름 검증(빈 이름 · 중복 · Run nil 거절) | 오류 반환 | `internal/app/engine/runtime.go` 보조 실행자 검증 |
| 나머지(`engineFillDetector` · `ReconcileDriver` · `ExitObserver` · `Recovery` · 전략 외곽 · `AlertDeliverer` · `NormalAlertRelayExecutor`) | 무편집 | 무편집 | AST calls |

## State mutations and fallbacks

- 지역 값만 만듦. fallback 없음.

## Safety conclusion

- Safe edit boundary: 보조 실행자 하나 추가. 감독 루프 집합(넷) · 순서 · 오류 경로 무변경. 보조 실행자는 감독 밖이라 정지해도 다른 루프를 내리지 않음(auxiliary.go 계약).
- High-risk impact: no — 새 실행자는 관측 · 알림만. 주문 · 손절 · 게이트 판정을 쓰지 않음.
