# Function Logic Map: `Client.ReconcileOpenOrdersPage`

- Source: `internal/official/reconcile_reads.go` (206-212, base `e57d06cd6`)
- AST evidence: `ast.json`(**편집 뒤** 재추출, revision current — 편집 전 갈래는 아래 「편집 전」 표와 커밋 `3bfe3c41c` 의 같은 파일)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: High-risk(reconciliation — 검증 대사 읽기, 판정이 기록 줄 쓰기를 가름)
- 생산 호출자: cmd `reconcileReads` 위임 하나 ← verifylive `Reconcile` 사슬(design §4, CodeGraph 1.6.0)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `symbol`·`cursor`·`limit` | artifact 심볼 · 커서 · 한도 | verifylive | 질의 문자열 |
| 응답 본문 | 계약 `/api/v1/orders` 성공 200 | 공식 OpenAPI | 해독 실패 오류 |

## Branches and early returns — 편집 전(base `e57d06cd6`)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 208 `getAcct(…, &raw)` 오류 | 없음 | 209 오류 그대로 | `TestReconcileClientReadsRefuseANullResult` · a168 `TestReconcileReadsRefuseANon200Success`(204) |
| (반환) | 211 `DecodeReconcileOrderPage(raw)` | 없음 | 해독 결과·오류 | `TestReconcileClientReadsSendTheDocumentedGetQueries` |

반환 좌표(AST, 편집 전): 209 · 211

## Branches and early returns — 편집 뒤(현재, 217-234)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 222 전송 오류 | 없음 | 223 오류 그대로 | `TestReconcileReadsRefuseANon200Success` `order-204-empty` |
| B2 | if 226 `DecodeReconcileOrderPage` 오류 | 없음 | 227 오류 | `TestReconcileClientReadsRefuseANullResult` |
| B3 | if 230 `reconcileAccepted` 실패 — a168 새 갈래 | 없음 | 231 오류 | `TestReconcileReadsRefuseAmbiguousBodies` order 사례(R4·W6·W7·R10·R10b) · `TestReconcileReadsRefuseANon200Success` order-201~299 · `TestReconcileReadsJudgeTheLastAttempt` order-401-then-201 · `TestReconcileRefusesAmbiguousBrokerResponsesOverHTTP` R8b(그룹 일반 OPEN 은 정상 — 조건 OPEN 거절 사례) |
| (반환) | 233 `page, nil` | 없음 | 해독 결과 | `TestReconcileReadsKeepTheirAcceptance` order-open-row |

반환 좌표(AST, 편집 뒤): 220 (closure) · 223 · 227 · 231 · 233. closure 반환은 `reconcileAttempts` 에 넘기는 읽기 함수 안의 것 — 관측자를 붙인 ctx 로 기존 호출을 그대로 부름.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `reconcileAttempts` | 관측자 부착 뒤 읽기 함수 호출, 시도 목록 반환(a168 새 함수) | 읽기 오류 그대로 | 새 파일 reconcile_response.go |
| `reconcileAccepted` | 마지막 시도 200 · 본문 모호함 판정(a168) | `ErrReconcileSchema` | 새 파일 |
| `c.getAcct` | 계좌 헤더 붙인 GET → `getWithHeaders` → `send`(401 갱신 재시도 최대 2회, 2xx 아니면 `classifyStatus`, 2xx 면 `unwrapAndDecode`) | 전송·인증·429·5xx·봉투 해독 실패는 오류 반환(자동 재시도 없음 — 401 만 send 안에서) | CodeGraph callees + AST |
| `DecodeReconcileOrderPage` | 일반 주문 목록 엄격 해독(a121) | `ErrReconcileSchema` | CodeGraph callees |

## State mutations and fallbacks

- 지역 변수(`raw`)만 채움. 상태·설정 바인딩 없음. 폴백 없음 — 모든 실패는 오류 반환.

## a168 편집(design D1·D2·D3·D6) — 적용됨

- 관측자 부착, 해독 뒤 `reconcileAccepted` 판정 갈래 추가.
- 편집 뒤 예상: 분기 3(B1 전송 오류 · B2 해독 오류 · B3 판정 실패), 반환 4.

## Safety conclusion

- Safe edit boundary: 관측자 부착 + 해독 **뒤** 판정 갈래 하나. 기존 갈래(입력 검사 · 전송 오류 · silent · 해독)의 조건·순서·반환 불변.
- High-risk impact: yes — 거절만 늘고(수락 갈래 신설 0) 수락 값은 바이트 단위로 같음. `send`·`unwrapAndDecode`·`reconcileObject` 불변.
