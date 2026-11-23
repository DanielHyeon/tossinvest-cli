# Function Logic Map: `TestA135SameStageOnANewAttestationIsAnnouncedAgain`

> **편집 후 지도 — 2단계 GREEN 뒤(2026-11-23).** 편집 전 원본 `../../pre-edit-ast/internal-app-engine--testa135samestageonanewattestationisannouncedagain.ast.json`, 재번호 `renumber.txt`(옛/새 분기를 (종류, 줄 문장)으로 difflib 정렬 — 생성기 `../../renumber.py.txt`).

- Source: `internal/app/engine/a135_attestation_watch_internal_test.go` (471-504)
- AST evidence: `ast.json` — branches **4** / returns **0** / calls **29**, source SHA-256 `e097c1707e7a80769865f2b46dc0b98a7801632c78dff7725d507bb0075c55c9`(파일 단위, 편집 후).
  명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/app/engine/a135_attestation_watch_internal_test.go --func TestA135SameStageOnANewAttestationIsAnnouncedAgain`.
- Risk scan: `risk-pattern-report.md`.
- a166 census: `attestation_watch.go`·`a135_attestation_watch_internal_test.go` 0건(1.4).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 전반: 만료된 A → 만료된 B 같은 날 | a196 D5 | 이 함수 | — |
| 후반: D-5 A → 발급·만료 +6h(같은 증거) → 증거 +1분(기한 변경) | a196 D4 | 이 함수 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@480) | 전반 사건 ≠ 1 | Fatalf | — | (자기 자신) |
| B2 (if@493) | 같은 증거 재발급 뒤 사건 ≠ 1 | Fatalf | — | — |
| B3 (range@497) | 증거 시각 +1분 | 픽스처 변경 | — | — |
| B4 (if@501) | 기한 변경 뒤 사건 ≠ 2 또는 Key 같음 | Fatalf | — | — |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `a135Watch`·`checkOnce`·`n.got` | 실행자 경유 | — | AST calls 29 |

## State mutations and fallbacks

- 재작성(Q5): 종전 단언(같은 날 다른 만료 파일 2건, 발급만 +1분이면 새 신원)은 a196 D4·D5 와 반대 → 새 규칙으로. renumber: 전반 B1·B2 changed, 후반 B3→B4 same.
- 통과 이유가 바뀐 부분(design §5-12): 종전 후반은 D1 식이 유효 기간을 1분 줄여 통과했음 → 이제 발급·만료를 함께 밀어 유효 기간을 지키고, 기한 변경은 증거 시각으로 명시.

## Safety conclusion

- Safe edit boundary: 시험 전체. High-risk impact: no — 시험.
