# Function Logic Map: `Journal.TransitionOperatingMode`

- Source: `internal/journal/operating_mode.go` (346-485)
- AST evidence: `ast.json` — branches 28, returns 20, calls 39, assignments 20,
  **defers 1, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**H1과 design D1이 둘 다 이 함수의 AST를 근거로 삼는다.** a092는 편집하지 않는다.

두 가지 사실만 필요하다.

1. **`direction == 0`은 announce 전에 반환한다** — 오늘의 두절 사이클이 68초가 아니라
   34초인 이유. B15 `:410` → `return :415`.
2. **announce는 커밋 *뒤*에 있다** — 호출자 기한이 원장 트랜잭션을 자른다는 주장이
   거짓인 이유. `tx.Commit()` `:468` → `AnnounceOperatingMode` `:479`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `req.AccountRef` | 비어 있지 않음 | 호출자 | B2 `:354` 오류 |
| `req.Mode` | `ValidOperatingMode` | `operating_mode.go` | B3 `:357` 오류 |
| `req.Actor` | AUTO 또는 OPERATOR | 같은 위 | B4 `:360` 오류 |
| `req.Cause` | 비어 있지 않음 | 호출자 | B5 `:364` 오류 |
| `req.Announcer` | **nil 허용** | 프로덕션은 `ectx.Notifier` | nil이면 B27이 건너뛴다 — **알림 없음** |
| `req.Auditor` | nil 허용 | — | 완화(`direction < 0`)에는 필수(B19) |
| `current.Mode` | 트랜잭션 안에서 읽는다 `:398` | `operating_modes` 테이블 | 호출자가 믿은 값이 아니다 |

## Branches and early returns

| Branch | 위치 | 조건 | 반환 | `Notify` 도달 |
|---|---|---|---|---|
| B1 | `:353` | 요청 검증 switch | — | — |
| B2 | `:354` | `account == ""` | `:355` 오류 | ❌ |
| B3 | `:357` | `!ValidOperatingMode(mode)` | `:358` 오류 | ❌ |
| B4 | `:360` | `!ValidModeActor(actor)` | `:361` 오류 | ❌ |
| B5 | `:364` | `cause == ""` | `:365` 오류 | ❌ |
| B6 | `:371` | `actor == ModeActorAuto` | — | — |
| B7 | `:372` | 자동 요청의 목표 모드 switch | — | — |
| B8 | `:373` | `ModeHaltAll` | `:374` `ErrHaltAllIsNeverAutomatic` | ❌ |
| B9 | `:375` | `ModeNormal` — 자동 완화 | `:378` 오류 | ❌ |
| B10 | `:381` | `!AutomaticTrigger(cause)` | `:382` 오류 | ❌ |
| B11 | `:392` | `BeginTx` 실패 | `:393` 오류 | ❌ |
| B12 | `:399` | `currentModeTx` 실패 | `:400` 오류 | ❌ |
| B13 | `:406` | `modeDirection` 실패 | `:407` 오류 | ❌ |
| B14 | `:409` | 방향 switch | — | — |
| **B15** | **`:410`** | **`direction == 0` — 이미 그 모드다** | **`:415` `changed=false`, err nil** | **❌ — H1** |
| B16 | `:417` | `direction < 0 && actor == Auto` — 보수 우선 | `:421` `changed=false` | ❌ |
| B17 | `:423` | `direction < 0` — 완화 | — | — |
| B18 | `:424` | `approval == ""` | `:425` 오류 | ❌ |
| B19 | `:428` | `req.Auditor == nil` | `:429` 오류 | ❌ |
| B20 | `:437` | `modeRowCount` 실패 | `:438` 오류 | ❌ |
| B21 | `:441` | `id == ""` — id 생성 | — | — |
| B22 | `:444` | `INSERT` 실패 | `:448` 오류 | ❌ |
| B23 | `:460` | `req.Auditor != nil` | — | — |
| B24 | `:461` | 감사 기록 실패 | `:463` 오류 — **커밋 전이므로 롤백** | ❌ |
| **B25** | **`:468`** | **`tx.Commit()` 실패** | `:469` 오류 | ❌ |
| B26 | `:475` | projector 있음 | — | — |
| **B27** | **`:478`** | **`req.Announcer != nil`** | — | **여기부터 ✅** |
| **B28** | **`:479`** | **`AnnounceOperatingMode` 실패** | `:480` `record, true, err` | **✅ 이미 도달함** |
| — | `:484` | — | `record, true, nil` | ✅ 이미 도달함 |

### H1 — 오늘의 두절 사이클이 34초인 이유

B15가 `:415`에서 반환하고, `AnnounceOperatingMode`는 `:479`다. **64줄 앞이다.**
`ModeTriggerExitObservationOutage`의 목표는 `ENTRY_BLOCKED`이고(`TargetModeForTrigger`
`:537-549`), 계정은 2026-07-31부터 그 모드다(live journal `operating_modes` 1행:
`ENTRY_BLOCKED / AUTO / BROKER_AUTH_REJECTED / 2026-07-31T09:55:49Z`, read-only 확인).
`direction == 0`이므로 announce는 **일어나지 않는다.**

로그 전체에서 `AnnounceOperatingMode`가 쓴 모양의 줄(`from_state` 필드 보유)은
**line 372 하나**다(`analysis/delivery-latency.md` §0). **두절 사이클의 오늘 비용은
34초 1회이고, 68초는 계정이 NORMAL일 때의 상한이다.**

### design D1 — 호출자 기한이 원장을 자르지 못하는 이유

`BeginTx` `:391` → `tx.Commit()` `:468` → `AnnounceOperatingMode` `:479`.
**announce는 트랜잭션 밖이다.** 그러므로 "호출자에 짧은 기한을 씌우면 announce가
트랜잭션을 붙잡는다"는 서술은 틀렸다. 실제 문제는 다르다 — 짧은 기한은
**`BeginTx`(`:391`) 이전에 만료된다.** 그러면 트랜잭션이 아예 시작되지 않아
운영 모드 승격 자체가 사라진다. 예산은 `Notify` 아래 transport 계층이 가져야
두 경로에 다 걸리고 원장에는 안 걸린다.

## Calls and live bindings

| Callee | 위치 | 계약 | Evidence |
|---|---|---|---|
| `j.db.BeginTx` | `:391` | `BEGIN IMMEDIATE` — 여기부터 트랜잭션 | AST calls |
| `tx.Rollback` | `:396` (defer) | 커밋됐으면 무해 | AST defers 1 |
| `currentModeTx` | `:398` | 트랜잭션 안에서 현재 모드 | AST calls |
| `req.Auditor.RecordAction` | `:461` | **커밋 전** — 실패는 전이를 중단시킨다 | AST calls |
| `tx.Commit` | `:468` | 여기서 내구성 확보 | AST calls |
| `p.ProjectOperatingMode` | `:476` | 커밋 후 투영 | AST calls |
| `req.Announcer.AnnounceOperatingMode` | `:479` | **커밋 후·동기.** `obs/mode.go:57` `Notify` → critical 예산 | `analysis/notify-reach.md` P1a/P1b |

## State mutations and fallbacks

| 대상 | 자리 | 성격 |
|---|---|---|
| `operating_modes` 행 | `:444` | 트랜잭션 안. B15·B16은 여기 도달하지 않는다 |
| 감사 로그 | `:461` | 커밋 전 — 실패하면 전이가 없다 |
| projection | `:476` | 커밋 후 |

- **announce 실패는 전이를 되돌리지 않는다.** B28은 `record, true, err`를 반환한다 —
  모드는 바뀌었고 알림만 실패한 상태다. 호출자는 `changed=true`와 오류를 동시에 본다.
- goroutine 없음(AST `go_statements: 0`).

## Safety conclusion

- **Safe edit boundary**: a092는 편집하지 않는다.
- **High-risk impact**: **yes** — §0.5(원장·인증 경로).
- **a092가 여기서 얻는 사실 두 가지**: (1) 오늘 헤드라인은 34초다. (2) 예산은
  호출자 기한이 아니라 transport 계층에 있어야 한다.
