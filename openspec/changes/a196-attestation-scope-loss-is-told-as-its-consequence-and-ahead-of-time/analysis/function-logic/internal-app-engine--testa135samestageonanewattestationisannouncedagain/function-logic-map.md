# Function Logic Map: `TestA135SameStageOnANewAttestationIsAnnouncedAgain`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-23, base `064f59fd9`.** 시험 함수 — 이 change 의 D4·D5 와 정면으로 반대되는 단언을 가진다.

- Source: `internal/app/engine/a135_attestation_watch_internal_test.go` (467-496)
- AST evidence: `ast.json` — branches **3** / returns **0** / calls **23**, source SHA-256 `d48d3091352d1daae0c579962f39930fcda7d315f67ffe71d350d9c867e57e1d`(파일 단위).
  명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/app/engine/a135_attestation_watch_internal_test.go --func TestA135SameStageOnANewAttestationIsAnnouncedAgain`.
- Risk scan: `risk-pattern-report.md` — go-panic 2건은 이 함수 밖(:376·:539).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 전반: 만료된 A(`a135Attestation(-1h)`) → 만료된 B(`-2h`) | 같은 날·같은 단계·다른 발급/만료 | L470·L474 | — |
| 후반: D-5 의 A → `IssuedAt` 만 +1분(만료 그대로) | L488·L491 | — | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@478) | 전반 사건 수 ≠ 2 | `t.Fatalf`(@479) 「different expired attestation … swallowed」 | — | (자기 자신) |
| B2 (if@481) | 전반 두 Key 같음 | `t.Fatalf`(@482) | — | — |
| B3 (if@493) | 후반 사건 수 ≠ 2 또는 Key 같음 | `t.Fatalf`(@494) 「issued-at change was not a new identity」 | — | — |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `a135Watch`·`checkOnce`(call@471·@472·@476·@489·@490·@492) | 실행자 경유 | — | AST |

## State mutations and fallbacks

- **a196 편집 뒤**:
  - 전반(B1·B2): D5 는 거절 신원에서 발급 신원을 뺀다 → 같은 날 같은 `expired` 는 1건, Key 같음 → **B1 에서 FAIL 예상**. 이 change 가 의도해 뒤집는 단언(design §5-11).
  - 후반(B3): D1 식은 유효 기간을 `ExpiresAt − IssuedAt` 로 읽으므로 `IssuedAt` 만 +1분 하면 유효 기간이 1분 줄고 감독 증거의 기한도 1분 당겨진다 → 신원(기한 시각)이 바뀌어 **2건·다른 Key 로 우연히 통과한다.** 실제 갱신(발급·만료가 같이 6시간 이동, 같은 감독 증거)에서는 D4 로 0건이어야 하므로, 이 후반의 이름·의도(「발급만 달라도 새 신원」)는 D4 와 반대다. 통과 이유가 바뀐 시험(기억 「통과는 증거가 아니다」).
- 편집 경로(Manager 결정): 함수 전체를 a196 의 의미로 바꾼다(전반 → 같은 날 1건, 후반 → 발급·만료 동시 이동 + 같은 감독 증거 → 추가 0건), 또는 삭제하고 design §5-2·§5-6 새 시험으로 대체.

## Safety conclusion

- Safe edit boundary: 시험 전체.
- High-risk impact: no — 시험.
