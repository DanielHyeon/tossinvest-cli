# Function Logic Map: `ReconcileDriver.RunOnce` (engine)

- Source: `internal/app/engine/reconcileloop.go` (418-472)
- 편집 전: `../../pre-edit-ast/engine--reconciledriver.runonce.ast.json`(414~466, sha b8e9b21373ea, 분기 8).
- AST evidence: `ast.json`(편집 뒤 sha cb2e95a4ccbd, 분기 8, return: return@423 return@430 return@455 return@467 return@471). 명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/app/engine/reconcileloop.go --func ReconcileDriver.RunOnce`.
- Risk scan: `risk-pattern-report.md`(일치 0).
- 재번호: `renumber.txt`(`../renumber.py` — 옛/새 분기를 (종류, 그 줄의 소스 문장)으로 정렬. 손 재번호 안 함).

## 재번호

```
B1:if@418 -> B1:if@422
B2:if@424 -> B2:if@428
B3:if@435 -> B3:if@439
B4:if@441 -> B4:if@445
B5:if@447 -> B5:if@451
B6:if@448 -> B6:if@452
B7:if@457 -> B7:if@463
B8:if@458 -> B8:if@464
```

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `outcome` | Observe 결과 | Tracker | — |
| 불변식 | 주기 순서·오류 처분 무변경 | design D7 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@422) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B2 (if@428) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B3 (if@439) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B4 (if@445) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B5 (if@451) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B6 (if@452) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B7 (if@463) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B8 (if@464) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `d.logOrderTerminalReleases`(call@459) | 해제마다 로그 1줄 + 유지 사유가 바뀔 때 1줄(계좌·주문번호·수량·가격 없음) | 오류 없음 | AST |
| `d.opts.Tracker.Observe`(call@457) | 기존 | 기존 | AST |

## State mutations and fallbacks

- 편집: 호출 1줄(분기 아님). 드라이버 구조체에 `terminalHoldLogged` 필드(타입 편집 — 이 함수 밖).

## Safety conclusion

- Safe edit boundary: 호출 1줄, 오류 갈래(B7·B8) 앞. 로그만 — 판정·순서 무변경.
- High-risk impact: no(관측 전용).
