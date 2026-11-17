# Function Logic Map: `buildGateway` (engine)

- Source: `internal/app/engine/gateway.go` (250-388)
- 편집 전: `../../pre-edit-ast/engine--buildgateway.ast.json`(250~386, sha d5329ce26ca4, 분기 10).
- AST evidence: `ast.json`(편집 뒤 sha 92c70f28eb70, 분기 10, return: return@256 return@280 return@288 return@292 return@295 return@316 return@320 return@341 return@351 return@355 return@373). 명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/app/engine/gateway.go --func buildGateway`.
- Risk scan: `risk-pattern-report.md`(일치 0).
- 재번호: `renumber.txt`(`../renumber.py` — 옛/새 분기를 (종류, 그 줄의 소스 문장)으로 정렬. 손 재번호 안 함).

## 재번호

```
B1:if@255 -> B1:if@255
B2:if@277 -> B2:if@279
B3:if@285 -> B3:if@287
B4:if@289 -> B4:if@291
B5:if@313 -> B5:if@315
B6:if@317 -> B6:if@319
B7:if@338 -> B7:if@340
B8:if@346 -> B8:if@348
B9:if@348 -> B9:if@350
B10:if@352 -> B10:if@354
```

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `in.journal` | 운영 장부 | 엔진 조립 | 기존 |
| 불변식 | 추적기 배선 외 무변경 | design D11 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@255) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B2 (if@279) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B3 (if@287) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B4 (if@291) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B5 (if@315) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B6 (if@319) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B7 (if@340) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B8 (if@348) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B9 (if@350) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B10 (if@354) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `tracker.Restore`(call@279) | 복원(기존) | 기존 | AST |

## State mutations and fallbacks

- 편집: 추적기 복합 리터럴에 `Terminal: in.journal` 필드 1(분기 아님) — `TestA192TheProductionTrackerReadsTerminalProofFromItsJournal` 이 고정.

## Safety conclusion

- Safe edit boundary: 필드 1. 분기 10개 문장 불변(줄 +2).
- High-risk impact: yes(배선) — 이 필드가 운영에서 종결 증명 해제를 켬.
