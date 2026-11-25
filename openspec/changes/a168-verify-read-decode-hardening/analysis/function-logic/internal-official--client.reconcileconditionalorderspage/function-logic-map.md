# Function Logic Map: `Client.ReconcileConditionalOrdersPage`

- Source: `internal/official/reconcile_reads.go` (194-203, base `e57d06cd6`)
- AST evidence: `ast.json`(편집 전 — 편집 뒤 재추출해 이 문서를 갱신함)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: High-risk(reconciliation — 검증 대사 읽기, 판정이 기록 줄 쓰기를 가름)
- 생산 호출자: cmd `reconcileReads` 위임 하나 ← verifylive `Reconcile` 사슬(design §4, CodeGraph 1.6.0)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `status` | `OPEN` · `CLOSED`(호출자 `readReconcileConditionalGroup` 이 고정) | verifylive 상수 | 빈 값이면 `ErrOrderStatusRequired`(B1) — 요청 0 |
| `symbol`·`cursor`·`limit` | artifact 줄 심볼 · 직전 페이지 커서 · `reconcileListLimit` | verifylive | 질의 문자열로만 감 |
| 응답 본문 | 계약 `PaginatedConditionalOrderResponse`(성공 200 만) | 공식 OpenAPI | 해독 실패는 `ErrReconcileSchema`/`ErrServer` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 195 `strings.TrimSpace(status) == ""` | 없음 | 196 `ErrOrderStatusRequired` 감싼 오류 | `TestReconcileReadsKeepTheirInputChecks`(a168, 특성) |
| B2 | if 199 `getAcct(…, &raw)` 오류 | 없음 | 200 오류 그대로 | `TestReconcileClientReadsRefuseANullResult`(봉투 null → 해독 오류는 B2 아님, 아래 반환 202) · a168 `TestReconcileReadsRefuseANon200Success`(204 빈 본문 ErrServer) |
| (반환) | 202 `DecodeReconcileConditionalPage(raw)` | 없음 | 해독 결과·오류 | `TestReconcileClientReadsSendTheDocumentedGetQueries` · `TestReconcileClientReadsRefuseANullResult` |

반환 좌표(AST, 편집 전): 196 · 200 · 202

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.getAcct` | 계좌 헤더 붙인 GET → `getWithHeaders` → `send`(401 갱신 재시도 최대 2회, 2xx 아니면 `classifyStatus`, 2xx 면 `unwrapAndDecode`) | 전송·인증·429·5xx·봉투 해독 실패는 오류 반환(자동 재시도 없음 — 401 만 send 안에서) | CodeGraph callees + AST |
| `reconcileListQuery` | 질의 조립 | 실패 없음 | AST |
| `DecodeReconcileConditionalPage` | 봉투 안 result 엄격 해독(a121) | `ErrReconcileSchema` | CodeGraph callees |

## State mutations and fallbacks

- 지역 변수(`raw`)만 채움. 상태·설정 바인딩 없음. 폴백 없음 — 모든 실패는 오류 반환.

## a168 편집 계획(design D1·D2·D3·D6)

- 관측자 부착: `getAcct` 호출을 `reconcileAttempts(ctx, func(ctx) error { … })` 로 감쌈(새 파일 함수). 오류 갈래 B2 불변.
- 해독(반환 202 의 호출) 뒤에 `reconcileAccepted(attempts)` 판정 갈래 하나 추가 — 실패면 빈 페이지 + 오류(D1 200 · D2 접히는 짝 · D6 계약 키 변형).
- 편집 뒤 예상: 분기 4(B1 빈 status · B2 전송 오류 · B3 해독 오류 · B4 판정 실패), 반환 5.

## Safety conclusion

- Safe edit boundary: 관측자 부착 + 해독 **뒤** 판정 갈래 하나. 기존 갈래(입력 검사 · 전송 오류 · silent · 해독)의 조건·순서·반환 불변.
- High-risk impact: yes — 거절만 늘고(수락 갈래 신설 0) 수락 값은 바이트 단위로 같음. `send`·`unwrapAndDecode`·`reconcileObject` 불변.
