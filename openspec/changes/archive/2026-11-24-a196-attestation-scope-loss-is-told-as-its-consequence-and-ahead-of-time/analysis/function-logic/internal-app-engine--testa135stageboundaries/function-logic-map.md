# Function Logic Map: `TestA135StageBoundaries`

> **편집 후 지도 — 2단계 GREEN 뒤(2026-11-23).** 편집 전 원본 `../../pre-edit-ast/internal-app-engine--testa135stageboundaries.ast.json`, 재번호 `renumber.txt`(옛/새 분기를 (종류, 줄 문장)으로 difflib 정렬 — 생성기 `../../renumber.py.txt`).

- Source: `internal/app/engine/a135_attestation_watch_internal_test.go` (64-103)
- AST evidence: `ast.json` — branches **4** / returns **0** / calls **8**, source SHA-256 `e097c1707e7a80769865f2b46dc0b98a7801632c78dff7725d507bb0075c55c9`(파일 단위, 편집 후).
  명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/app/engine/a135_attestation_watch_internal_test.go --func TestA135StageBoundaries`.
- Risk scan: `risk-pattern-report.md`.
- a166 census: `attestation_watch.go`·`a135_attestation_watch_internal_test.go` 0건(1.4).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 표 13행 | 기한(= 만료 − 1h, 픽스처 증거가 발급 1시간 전) 기준 경계 | 이 함수 | — |
| 헬퍼 `a135Attestation` | **무편집**(Q5) | — | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (range@89) | 표 순회 | t.Run | — | (자기 자신) |
| B2 (if@92) | stage 불일치 | Fatalf | — | — |
| B3 (if@95) | days 불일치(만료 아닌 행) | Fatalf | — | — |
| B4 (if@98) | critical 불일치 | Fatalf | — | — |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `judgeAttestation(a135Attestation(r), …)` | 판정 경계 | 순수 | AST |

## State mutations and fallbacks

- 표만 바뀜: remaining 에 1h 를 더해 행 이름을 「기한」 기준으로, 「기한 정각(만료 1시간 전) = 지남」 행 추가(proof_overdue). 단언 구조 무변경(renumber same 4).
- 통과 이유가 바뀐 행(design §5-12): 종전 「1초 = D-1」 은 기한이 이미 지나 있었음 → 이제 「기한 정각」 행이 그 사실을 직접 단언.

## Safety conclusion

- Safe edit boundary: 표 데이터. High-risk impact: no — 시험.
