# Function Logic Map: `TestA135StageBoundaries`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-23, base `064f59fd9`.** 시험 함수 — gate ⑤ 가 시험 파일 함수도 센다. 편집 여부는 Manager 결정(아래 「편집 경로」).

- Source: `internal/app/engine/a135_attestation_watch_internal_test.go` (64-100)
- AST evidence: `ast.json` — branches **4** / returns **0** / calls **8**, source SHA-256 `d48d3091352d1daae0c579962f39930fcda7d315f67ffe71d350d9c867e57e1d`(파일 단위).
  명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/app/engine/a135_attestation_watch_internal_test.go --func TestA135StageBoundaries`.
- Risk scan: `risk-pattern-report.md` — go-panic 2건(:376 `TestA135CheckPanicDoesNotStopTheWatch`, :539 `a135PanickingNotifier.Notify`) — **이 함수 밖**(64-100 아님), 의도된 패닉.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 표 12행(remaining·stage·days·critical) | L73-84 | 이 함수 | — |
| `a135Attestation(remaining)` | 만료 = now+remaining, 발급 = 만료−30d, **감독 증거 At = 발급−1h**(L37·L46-47) | 시험 헬퍼 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (range@86) | 표 순회 | `t.Run`(call@87) | — | (자기 자신) |
| B2 (if@89) | stage 불일치 | `t.Fatalf`(@90) | — | — |
| B3 (if@92) | 만료 아닌 행의 days 불일치 | `t.Fatalf`(@93) | — | — |
| B4 (if@95) | critical 불일치 | `t.Fatalf`(@96) | — | — |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `judgeAttestation(a135Attestation(r), nil, a135Now, a135Account)`(call@88) | 판정 경계 | 순수 | AST |

## State mutations and fallbacks

- 없음.
- **a196 편집 뒤 뒤집히는 행(overlay 프로브 실측, D1 식 그대로 적용)**: 감독 증거의 기한 = 만료 − 1h 라서
  - 「D-8 은 아직 알리지 않음」 7d+1s → 기한 167h0m1s → D-7 notice(기대 healthy·8) — **FAIL 예상**
  - 「2일 1초 = D-3」 48h0m1s → 47h0m1s → D-2 imminent(기대 notice·3·non-critical) — **FAIL 예상**
  - 「1초 = D-1」 → −59m59s → `daysLeftCeil` 1 → imminent·1 — 우연 통과(0 이하 처리 방식에 의존)
  - 나머지 9행 통과(정확히 7일 → 167h → D-7 등, 프로브 표 review.md 1단계 절).

## 편집 경로 (Manager 결정)

1. 이 함수의 표를 고친다(경계 행을 「만료가 기한」·「감독 증거가 기한」 둘로) — 번들: 이 함수.
2. 헬퍼 `a135Attestation` 의 감독 증거 시각을 발급 시각과 같게(기한 = 만료) — 번들: `a135Attestation`(이 change 에 아직 없음), 이 함수는 무편집.
   발급자 계약(acceptSupervised `age < 0` 만 제외)상 At = IssuedAt 은 실제로 가능한 값.

## Safety conclusion

- Safe edit boundary: 표 데이터. 단언 구조(B2~B4)는 유지.
- High-risk impact: no — 시험.
