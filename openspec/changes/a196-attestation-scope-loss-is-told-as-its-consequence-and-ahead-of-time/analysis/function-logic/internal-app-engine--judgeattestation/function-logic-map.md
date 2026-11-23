# Function Logic Map: `judgeAttestation`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-23, base `064f59fd9`.** 2단계 GREEN 뒤 AST·지도를 다시 만든다.

- Source: `internal/app/engine/attestation_watch.go` (87-122)
- AST evidence: `ast.json` — branches **12** / returns **3** / calls **10**, source SHA-256 `29252ace6e7338abe7fa91bf9a7cfe894b8edde2c87657e60e9f54e1e9c7d68d`(파일 단위).
  명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/app/engine/attestation_watch.go --func judgeAttestation`.
- Risk scan: `risk-pattern-report.md`(일치 0).
- a166 census: `shapes.json`·`sites.json` 에 `attestation_watch.go` 0건(grep).
- 호출자: production `attestationWatch.prepare`(call judgeAttestation@283) 1, 시험 3(`TestA135StageBoundaries`·`TestA135ScopeLossIsCritical`·`TestA135UnusableAttestationsAreCritical`) — CodeGraph 1.6.0 + grep.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `att attest.Attestation` | `attest.Load` 결과(JSON RFC3339Nano 왕복 — 시각 손실 없음) | 인터록과 같은 파일 | `loadErr` 가 있으면 내용 무시 |
| `loadErr` | nil 또는 `ErrMissing`/`ErrMalformed`/`ErrFormatTooNew` 를 감싼 오류 | `attest.Load`(attest.go:161-179) | B1 → unreadable |
| `now` | 감시 시계 | `w.clock.Now()`(prepare call@282) | — |
| `accountRef` | 엔진 자격 증명 계좌 | `Context.AccountRef` | Verify 가 비교, 판정 결과에 싣지 않음 |
| 불변식 | 판정 술어 = 인터록 술어(`att.Verify` + `RequiredEndpoints()`) — 사본 금지 | engine-safety 정본 | — |
| 불변식(a196 이 쓸 사실) | 발급자가 쓴 파일이면 `SupervisedBy[].At ≤ IssuedAt`(acceptSupervised B6 if@332 의 `age < 0` 제외), `ExpiresAt − IssuedAt = c.Validity`(BuildAttestation :264·:265 같은 `now`·같은 `c`) → 감독 증거가 실리면 감독 증거의 기한 ≤ 만료 | `internal/soak/attest.go`(cited-ast) | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@88) | `loadErr != nil` | 없음 | return@89 `{Stage: unreadable}` — 발급·만료·ProofAt 0 | `TestA135UnusableAttestationsAreCritical`(파일 없음·깨진 파일) |
| (L91) | — | `v` 생성: `IssuedAt`·`ExpiresAt`·`ProofAt: binding(att.SupervisedBy)`(call@91) | — | `TestA135SupervisedProofAtIsTheBindingProof`(binding 단독) |
| B2 (if@93) | `att.Verify(now, accountRef, required)` 오류 | — | — | 아래 B4~B8 |
| B3 (switch@94) | 오류 종류 | — | — | — |
| B4 (case@95) | `ErrExpired` | `v.Stage = expired` | return@108 | `TestA135StageBoundaries`(만료 순간·1시간 뒤) |
| B5 (case@97) | `ErrEndpointNotAttested` | `v.Stage = scope_lost`, `v.Missing = att.MissingEndpoints(required)`(call@99, 정렬됨 — attest.go:256) | return@108 | `TestA135ScopeLossIsCritical` |
| B6 (case@100) | `ErrAccountMismatch` | `v.Stage = account_mismatch` | return@108 | `TestA135UnusableAttestationsAreCritical`(다른 계좌) |
| B7 (case@102) | `ErrIncomplete` | `v.Stage = incomplete` | return@108 | `TestA135UnusableAttestationsAreCritical`(endpoint 없음) |
| B8 (case@104) | 그 밖 | `v.Stage = unreadable` | return@108 | **없음** — 커버 0(base-cover 104.11-106.29 count 0). Verify 가 돌려주는 sentinel 은 넷뿐이라 현재 도달 불가 |
| (L112) | Verify 통과 | `v.DaysLeft = daysLeftCeil(att.ExpiresAt.Sub(now))`(call@112) — **a196 의 편집 지점**(기준 시각을 min(만료, 감독 증거의 기한)으로) | — | `TestA135StageBoundaries` |
| B9 (switch@113) | 날수 단계 | — | — | — |
| B10 (case@114) | `DaysLeft > 7` | `healthy` | return@121 | StageBoundaries「D-8」「30일 남음」 |
| B11 (case@116) | `DaysLeft > 2` | `notice` | return@121 | StageBoundaries「정확히 7일」~「2일 1초」 |
| B12 (case@118) | 그 밖 | `imminent` | return@121 | StageBoundaries「정확히 2일」~「1초」 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `binding`(call@91) | 감독 증거 중 가장 먼저 기간을 벗어날 것 | 순수. 증거 0 → zero time | AST `binding` |
| `RequiredEndpoints`(call@92) | 인터록 필수 집합(GET 6 + POST 2) | 순수 | interlock.go:219-230 |
| `att.Verify`(call@93) | 인터록 술어 — 필수 칸 → 만료(`!now.Before(ExpiresAt)`) → 계좌 → endpoint | sentinel 4종 | cited-ast `internal-attest--attestation.verify` |
| `daysLeftCeil`(call@112) | 하루 단위 올림 | 「양수로 들어옴」 전제(주석 :124). 0 이하가 들어가면 `int` 나눗셈이 0 쪽으로 잘려 −59m59s → 1(overlay 프로브 실측, review.md 1단계 절) | AST `daysleftceil` |

## State mutations and fallbacks

- 지역 `v` 만 만든다. 수신자·전역 쓰기 0, 브로커·파일 호출 0(Load 는 호출자 몫).
- fallback: 알 수 없는 Verify 오류 → unreadable(B8).
- a196 편집 예정: L112 의 기준 시각, verdict 에 기한 종류·시각 필드 추가(구조체 선언은 함수 본문 밖). Manager 결정 2026-11-23 Q4: 남은 시간 ≤ 0 은 새 단계 `proof_overdue` 로 `daysLeftCeil` 호출 **전에** 가른다 → `daysLeftCeil` 무편집. 새 단계는 `attestationVerdict.critical` 의 default(B3 case@78 → return@79 true)가 흡수 → `critical` 무편집, 번들 없음.

## Safety conclusion

- Safe edit boundary: Verify 통과 뒤(L112~121)만. B1~B8(인터록 술어와 그 분류)은 바꾸지 않는다.
- High-risk impact: no — 관측 전용 판정. 진입·청산·운영 모드·게이트에 닿지 않는다(호출자 prepare → 알림만).
