# Function Logic Map: `Client.ReconcileInstrument`

- Source: `internal/official/reconcile_reads.go` (215-228, base `e57d06cd6`)
- AST evidence: `ast.json`(**편집 뒤** 재추출, revision current — 편집 전 갈래는 아래 「편집 전」 표와 커밋 `3bfe3c41c` 의 같은 파일)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: High-risk(reconciliation — 검증 대사 읽기, 판정이 기록 줄 쓰기를 가름)
- 생산 호출자: cmd `reconcileReads` 위임 하나 ← verifylive `Reconcile` 사슬(design §4, CodeGraph 1.6.0)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `symbol` | artifact 줄 심볼 바이트 | verifylive `checkReconcileInstrument` | 질의 `symbols` |
| 응답 본문 | `/api/v1/stocks` 성공 200, result 는 배열 | 공식 OpenAPI | 해독 실패 `ErrServer` |

## Branches and early returns — 편집 전(base `e57d06cd6`)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 221 `c.get(…, &raw)` 오류 | 없음 | 222 오류 그대로 | a168 `TestReconcileReadsRefuseANon200Success`(204 빈 본문) |
| B2 | if 224 `len(raw) != 1` | 없음 | 225 `ErrReconcileSchema` | `TestReconcileReadsKeepTheirInputChecks`(a168, 특성 — 두 항목) |
| (반환) | 227 `raw[0].Symbol` | 없음 | 메아리 | `TestReconcileInstrumentReturnsTheBrokersEchoNotItsInput` |

반환 좌표(AST, 편집 전): 222 · 225 · 227

## Branches and early returns — 편집 뒤(현재, 237-257)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 246 전송 오류 | 없음 | 247 오류 | `TestReconcileReadsRefuseANon200Success` `inst-204-empty` |
| B2 | if 249 `len(raw) != 1` | 없음 | 250 ErrReconcileSchema | `TestReconcileReadsKeepTheirInputChecks` |
| B3 | if 253 `reconcileAccepted` 실패 — a168 새 갈래 | 없음 | 254 오류 | `TestReconcileReadsRefuseAmbiguousBodies` I1 · `TestReconcileReadsRefuseANon200Success` inst-201~299 · `TestReconcileReadsJudgeTheLastAttempt` inst-401-then-201 |
| (반환) | 256 `raw[0].Symbol` | 없음 | 메아리 | `TestReconcileReadsKeepTheirAcceptance` inst · `TestReconcileInstrumentReturnsTheBrokersEchoNotItsInput` |

반환 좌표(AST, 편집 뒤): 244 (closure) · 247 · 250 · 254 · 256. closure 반환은 `reconcileAttempts` 에 넘기는 읽기 함수 안의 것 — 관측자를 붙인 ctx 로 기존 호출을 그대로 부름.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `reconcileAttempts` | 관측자 부착(a168) | 읽기 오류 그대로 | 새 파일 |
| `reconcileAccepted` | 마지막 시도 판정(a168) | `ErrReconcileSchema` | 새 파일 |
| `c.get` | 계좌 헤더 없는 GET → `getWithHeaders` → `send` | getAcct 와 같은 send 계약 | AST |

## State mutations and fallbacks

- 지역 변수(`raw`)만 채움. 상태·설정 바인딩 없음. 폴백 없음 — 모든 실패는 오류 반환.

## a168 편집(design D1·D2·D3·D6) — 적용됨

- 관측자 부착, 항목 수 판정(B2) 뒤에 `reconcileAccepted` 판정 갈래 추가(Manager 결정 2).
- 편집 뒤 예상: 분기 3(B1 · B2 · B3 판정 실패), 반환 4.

## Safety conclusion

- Safe edit boundary: 관측자 부착 + 해독 **뒤** 판정 갈래 하나. 기존 갈래(입력 검사 · 전송 오류 · silent · 해독)의 조건·순서·반환 불변.
- High-risk impact: yes — 거절만 늘고(수락 갈래 신설 0) 수락 값은 바이트 단위로 같음. `send`·`unwrapAndDecode`·`reconcileObject` 불변.
