# Function Logic Map: `Client.ReconcileConditionalOrder`

- Source: `internal/official/reconcile_detail_reads.go` (176-188, base `e57d06cd6`)
- AST evidence: `ast.json`(편집 전 — 편집 뒤 재추출해 이 문서를 갱신함)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: High-risk(reconciliation — 검증 대사 읽기, 판정이 기록 줄 쓰기를 가름)
- 생산 호출자: cmd `reconcileReads` 위임 하나 ← verifylive `Reconcile` 사슬(design §4, CodeGraph 1.6.0)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` | 기록이 고른 artifact id 바이트 | verifylive `readReconcileByID` | 빈 값이면 `ErrReconcileSchema`(B1) — 요청 0 |
| 응답 | 200 statement · 404 `conditional-order-not-found` silent · 그 밖 오류(a164 D3) | 공식 OpenAPI · M40 실측 | silent 는 오류 아님 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 177 `strings.TrimSpace(id) == ""` | 없음 | 178 `ErrReconcileSchema` | `TestReconcileReadsKeepTheirInputChecks`(a168, 특성) |
| B2 | if 181 `getAcct(…, &raw)` 오류 | 없음 | B3 로 | `TestReconcileByIDRefusesEveryOtherOutcome` · `TestReconcileByIDRefusesATransportError` |
| B3 | if 182 `reconcileDetailSilent(err)` | 없음 | 183 `Silent: true` · 아니면 185 오류 | `TestReconcileByIDClassifiesOnlyTheMeasuredNotFoundAsSilent` · `TestReconcileByIDRefusesEveryOtherOutcome` |
| (반환) | 187 `DecodeReconcileConditionalDetail(raw)` | 없음 | statement · 오류 | `TestReconcileByIDReadsTheReceiptShapedStatement` · `TestReconcileByIDRefusesSchemaHoles` |

반환 좌표(AST, 편집 전): 178 · 183 · 185 · 187

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.getAcct` | 계좌 헤더 붙인 GET → `getWithHeaders` → `send`(401 갱신 재시도 최대 2회, 2xx 아니면 `classifyStatus`, 2xx 면 `unwrapAndDecode`) | 전송·인증·429·5xx·봉투 해독 실패는 오류 반환(자동 재시도 없음 — 401 만 send 안에서) | CodeGraph callees + AST |
| `reconcileDetailSilent` | 404 본문 code 로 silent 판정(a164) | 실패 없음(불리언) | AST |
| `DecodeReconcileConditionalDetail` | statement 엄격 해독(a164, `reconcileObject`) | `ErrReconcileSchema` | CodeGraph callees |

## State mutations and fallbacks

- 지역 변수(`raw`)만 채움. 상태·설정 바인딩 없음. 폴백 없음 — 모든 실패는 오류 반환.

## a168 편집 계획(design D1·D2·D3·D6)

- 관측자 부착. B2·B3(오류·silent 갈래)는 그대로 앞 — silent 분류 불변.
- 해독 뒤 `reconcileAccepted` 판정 갈래 추가(D1 · D2 · D6 — D6 가 Manager 결정 4 의 by-id 단독 변형을 닫음).
- 편집 뒤 예상: 분기 5(B1 빈 id · B2 전송 오류 · B3 silent · B4 해독 오류 · B5 판정 실패), 반환 6.

## Safety conclusion

- Safe edit boundary: 관측자 부착 + 해독 **뒤** 판정 갈래 하나. 기존 갈래(입력 검사 · 전송 오류 · silent · 해독)의 조건·순서·반환 불변.
- High-risk impact: yes — 거절만 늘고(수락 갈래 신설 0) 수락 값은 바이트 단위로 같음. `send`·`unwrapAndDecode`·`reconcileObject` 불변.
