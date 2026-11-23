# Function Logic Map: `attestationWatch.prepare`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-23, base `064f59fd9`.** 2단계 GREEN 뒤 AST·지도를 다시 만든다.

- Source: `internal/app/engine/attestation_watch.go` (269-299)
- AST evidence: `ast.json` — branches **5** / returns **3** / calls **12**, source SHA-256 `29252ace6e7338abe7fa91bf9a7cfe894b8edde2c87657e60e9f54e1e9c7d68d`(파일 단위).
  명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/app/engine/attestation_watch.go --func attestationWatch.prepare`.
- Risk scan: `risk-pattern-report.md`(일치 0).
- a166 census: 0건. 호출자: `attestationWatch.checkOnce`(call w.prepare@232) 1 — **CodeGraph 1.6.0 `callers prepare` 는 이 호출을 못 찾았다**(같은 이름의 다른 `prepare` 8건만), grep 으로 확정.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `w.load(w.path)` | `attest.Load`(production)·시험 주입 | `newAttestationWatch` | 오류는 judge 가 unreadable 로 |
| `w.clock.Now()` | 감시 시계 | `Context.clock()`/주입 | — |
| `w.identity string` | 마지막 신원 `IssuedAt|ExpiresAt`(UTC RFC3339Nano) | 이 함수만 씀 | 초기값 "" |
| `w.sent map[string]bool` | 이 신원에서 이미 알린 단계 키 | 이 함수가 만들고 `checkOnce`(L241·L257)가 씀 | **초기값 nil**(`newAttestationWatch`·시험 `a135Watch` 모두 안 만듦) → 첫 호출 B3 가 만든다 |
| 불변식 | 상태는 Run 의 goroutine 하나만 만진다(잠금 없음, 주석 :162) | — | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@271) | defer 안 `recover() != nil` | 이름 있는 반환 `ok = false` | (패닉 회수) | `TestA135CheckPanicDoesNotStopTheWatch` |
| B2 (if@273) | `w.log != nil` | `w.log.Warn(EventAttestationWatchFailed, check_panicked)`(call@274) | — | 없음 — 시험 하네스의 log 는 nil(커버 273.20-277.5 count 0) |
| (L281-286) | — | 읽기·판정·**신원 계산 `identity := v.IssuedAt…+"|"+v.ExpiresAt…`(Format@286)** | — | — |
| B3 (if@287) | `identity != w.identity \|\| w.sent == nil` | `w.identity = identity`, `w.sent = map{}` — **신원이 바뀌면 모든 단계의 「보냄」을 잊는다**(거절 단계 포함). a196 D4·D5 의 편집 지점 | — | `TestA135RenewalResetsAndScopeLossIsImmediate`·`TestA135SameStageOnANewAttestationIsAnnouncedAgain` |
| B4 (if@291) | `v.Stage == healthy` | 없음 | return@292 `ok=false` | `TestA135WatchAnnouncesEachStageOnce`(D-8) |
| (L294) | — | `key = stageKey(v, now)`(call@294) | — | — |
| B5 (if@295) | `w.sent[key]` | 없음 | return@296 `ok=false` | `TestA135WatchAnnouncesEachStageOnce`(같은 D-7 재판정) |
| (끝) | — | `w.eventFor(v, key)`(call@298) — eventFor 가 `w.identity` 를 outbox Key 에 넣는다(eventFor L370) | return@298 `ok=true` | 다수 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `w.load`(call@281) | 파일 읽기 | 오류값만(패닉은 B1) | AST |
| `judgeAttestation`(call@283) | 판정 | 순수 | AST `judgeattestation` |
| `stageKey`(call@294) | 중복 방지 키 | 순수 | AST `stagekey` |
| `w.eventFor`(call@298) | 문구·필드·outbox Key | 순수(수신자 읽기만) | AST `attestationwatch.eventfor` |
| 알림기 | **이 함수 안에서 부르지 않는다** — 회수 경계 밖(checkOnce) | — | 주석 :268, AST calls 12 에 Notify 없음 |

## State mutations and fallbacks

- 수신자 쓰기: `w.identity`·`w.sent`(B3 안) 둘뿐. `w.sent[key] = true` 는 checkOnce 가 기록 성공(또는 알림기 없음) 뒤에 한다 — 기록 실패 시 「보냄」이 안 남아 다음 판정이 다시 보낸다(`TestA135FailedRecordIsRetried`).
- 상태는 프로세스 메모리뿐. 재시작하면 비어 있고, critical 은 알림기 outbox 가 같은 Key 를 `RemindAfter`(기본 1시간, obs/notifier.go:60) 안에서 `ClaimSettled` 로 삼킨다(notifier.go:326). normal 은 outbox 를 안 거쳐 재시작 뒤 다시 나간다(기존 동작).

## Safety conclusion

- Safe edit boundary: 신원 계산(L286)과 B3 의 재설정 규칙. 회수 경계(B1·defer)와 「알림기를 부르지 않는다」는 바꾸지 않는다.
- High-risk impact: no — 관측 전용.
