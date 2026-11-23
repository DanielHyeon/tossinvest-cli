# Function Logic Map: `stageKey`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-23, base `064f59fd9`.** 2단계 GREEN 뒤 AST·지도를 다시 만든다.

- Source: `internal/app/engine/attestation_watch.go` (303-308)
- AST evidence: `ast.json` — branches **1** / returns **2** / calls **4**, source SHA-256 `29252ace6e7338abe7fa91bf9a7cfe894b8edde2c87657e60e9f54e1e9c7d68d`(파일 단위).
  명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/app/engine/attestation_watch.go --func stageKey`.
- Risk scan: `risk-pattern-report.md`(일치 0). a166 census 0건. 호출자: `attestationWatch.prepare`(call stageKey@294) 1(CodeGraph 1.6.0 + grep).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `v.Stage`·`v.DaysLeft` | judge 결과 | `judgeAttestation` | — |
| `now` | 감시 시계 | prepare call@282 | — |
| `kst` | `time.FixedZone("KST", 9h)`(:311) — tz 데이터 무관 | 전역 | — |
| 불변식 | 키에 계좌·경로 없음(`TestA135AlertCarriesNoAccount` 가 Key 까지 검사) | engine-safety | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@304) | `notice \|\| imminent` | 없음 | return@305 `fmt.Sprintf("%s:%d", Stage, DaysLeft)` — 예 `notice:7` | `TestA135WatchAnnouncesEachStageOnce` |
| (끝) | 거절 단계·healthy(healthy 는 prepare B4 가 먼저 걸러 여기 안 옴) | 없음 | return@307 `string(Stage)+":"+now.In(kst).Format("2006-01-02")` — 예 `scope_lost:2026-11-23` | `TestA135InvalidStatesAreRemindedDaily` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fmt.Sprintf`(call@305)·`now.In(kst).Format`(call@307) | 문자열 조립 | 순수 | AST |

## State mutations and fallbacks

- 없음(순수). **키 자체에는 발급 신원이 없다** — 발급 신원은 prepare B3(「보냄」 재설정)와 eventFor L370(outbox Key 접미사)에 있다.
- a196 편집 예정(D5): 범위 상실 키에 정렬된 누락 endpoint 집합을 더한다(`v.Missing` 은 judge B5 에서 이미 정렬됨). 날수 키(D4)는 「보냄」 재설정의 신원이 기한 시각으로 바뀌면 그대로 둘 수 있다 — 어느 쪽으로 구현할지는 2단계.

## Safety conclusion

- Safe edit boundary: 반환 문자열 조립 두 줄.
- High-risk impact: no.
