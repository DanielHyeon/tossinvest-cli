# Function Logic Map: `NewContext`

- Source: `internal/app/engine/engine.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a074-critical-events-reach-the-operator/base-commit.txt`
- 위험 등급: **High-risk** — 엔진 기동의 전 과정이다. 자동화 게이트 인터록, 계좌 해석,
  원장 open, gateway 조립, Guardian 생성이 전부 여기 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `opts.ConfigDir` | 설정 디렉터리 | 호출자 | `NewOrderPath` 실패 → 반환 |
| `cfg.Engine.AutomationGate` | 게이트 설정 | 설정 파일 | 인터록이 판정 |
| `opts.Publisher` | 알림 전송 경로 | 호출자(오늘은 항상 nil) | nil이면 critical 알림 미전달 |
| `opts.Clock` | nil이면 시스템 시계 | 호출자 | — |
| `auditLog` | 열린 audit 로그 | `openAuditLog` | 실패 → 반환 |

**불변식 1 (유지)**: 구성 순서가 engine-safety에 고정되어 있다 — 계좌 해석 → journal →
gateway → 인터록. "Each step is a precondition for the next, and the interlock runs last
precisely so that it can verify what the earlier steps produced instead of taking their
success on trust."

**불변식 2 (유지)**: audit는 **어떤 거부보다 먼저** 기록된다(B5). 그 실패는 기동 거부다.

**불변식 3 (유지)**: 실패 경로마다 `jrn.Close()`가 호출된다(B8 이후). 새 코드가 원장을
연 뒤에 실패할 수 있으면 그 경로에도 Close가 필요하다 — **a074의 새 코드는 원장을 열기
전(B4 근처)에만 실행되므로 이 부담이 없다.** 이후 a072의 projection 조립 실패(B15)도
원장을 닫고 반환한다.

**현재의 결함**: `opts.Publisher`는 프로덕션에서 항상 nil이다.
`cmd/tossctl/engine_assembly.go`가 그것을 의도적으로 비워 두고 이유를 적어 두었다 —
"configuring a transport is an operational setting with an audit trail, which is a
change of its own." 그 결과 운영 원장의 critical 알림 6건이 `attempts=0`으로 PENDING에
남아 있다.

**a074가 바꾸는 것**: 두 줄.
1. `recordGateSettings(...)` 호출에 `cfg.Engine.Notifications`와 토큰 존재 여부를 넘긴다.
2. `buildGateway`의 `publisher:` 인자를 `opts.Publisher`에서 해석된 publisher로 바꾼다.

해석 로직은 **새 파일의 함수**가 갖는다. 이 함수에는 호출만 남는다.

**a074가 바꾸지 않는 것**: 당시 존재한 분기의 조건·순서·반환, `jrn.Close()` 위치,
인터록, Guardian 생성, 반환 구조체. 현재의 B15와 strategy projection 필드는 이후 a072에서
추가됐으며 이 재기준화에서 그대로 보존한다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (413) | `NewOrderPath` 실패 | 없음 | `nil, err` | 기존 |
| B2 (420) | `openAuditLog` 실패 | 없음 | `nil, err` | 기존 |
| B3 (424) | `opts.Clock == nil` | `clk = clock.System()` | — | 기존 |
| **B4 (438)** | `opts.Publisher != nil` | 해석된 publisher를 주입값으로 교체 | — | `TestAnInjectedPublisherWins` |
| **B5 (446)** | `recordGateSettings` 실패 | 없음 | `nil, err` — 기동 거부 | **6.1–6.3** |
| B6 (461) | 계좌 해석 실패 | audit refusal | `nil, err` | 기존 |
| B7 (468) | 원장 open 실패 | audit refusal | `nil, err` | 기존 |
| B8 (477) | apply hook 바인딩 실패 | `jrn.Close()` | `nil, err` | 기존 |
| **B9 (496)** | `buildGateway` 실패 | `jrn.Close()` | `nil, err` | **5.1–5.2** |
| B10 (511) | 게이트 ON + Guardian 없음 | Guardian 생성 | — | 기존 |
| B11 (513) | factory nil | 기본 factory | — | 기존 |
| B12 (519) | Guardian 생성 실패 | `jrn.Close()` | `nil, err` | 기존 |
| B13 (539) | 인터록 실패 | `jrn.Close()` | `nil, err` | 기존 |
| B14 (543) | 미검증 | `guardian = nil` | — | 기존 |
| B15 (548) | dormant strategy projection 생성 실패 | `jrn.Close()` | `nil, err` | `TestDormantSnapshotContainsExactPairedHonestMarkets` |

편집은 B5의 **인자**와 B9 직전 `gatewayInputs` 리터럴의 **필드 값** 두 곳이다. 어떤
분기 조건도, 어떤 `jrn.Close()`도, 어떤 반환도 바뀌지 않는다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `NewOrderPath` | 설정·경로·official 조립 | 실패 → 반환 | AST |
| `openAuditLog` | audit 로그 | 실패 → 반환 | AST |
| `newAutomationStatus` | 게이트 상태 | — | AST |
| `recordGateSettings` | §0.5 audit | 실패 → 기동 거부 | AST |
| `resolveAccountRef` | 계좌 확인 | 브로커 read 1회 | AST |
| `openEngineJournal` | 원장 | 실패 → 반환 | AST |
| `bindApplyHooks` | 체결 projection | 실패 → Close + 반환 | AST |
| `buildGateway` | gateway·notifier·retrier | 실패 → Close + 반환 | AST |
| `runInterlock` | 기동 인터록 | 실패 → Close + 반환 | AST |
| `resolveNotificationPublisher` (신규) | 설정 ⊕ 환경 → publisher | **error를 반환하지 않는다** | 신규 |

**신규 호출이 error를 반환하지 않는 이유** (design D7): 알림 전송은 보호 경로가 아니다.
오타 하나로 엔진이 뜨지 않으면 그때부터 손절 평가도 멈춘다 — 알림이 없는 상태보다
나쁘다. 해석할 수 없는 설정은 publisher 없이(= 오늘과 같이) 기동하고 사유를 audit과
로그에 남긴다. 그래도 안전 방향은 유지된다: publisher가 없으면 critical 알림은 outbox에
남고 entry gate가 잠기며 지속되면 ENTRY_BLOCKED로 강화된다.

**`opts.Publisher`가 여전히 우선한다**: 테스트가 주입한 publisher를 설정이 덮어쓰면
패키지 테스트가 프로세스 환경에 의존하게 된다.

## State mutations and fallbacks

- 파일 I/O: audit append, 원장 open. 네트워크: 계좌 해석 read 1회.
- 신규 코드는 **읽기만** 한다 — `cfg`(이미 메모리에 있음)와 환경변수.
- 신규 코드는 원장을 열기 전에 실행되므로 새 `jrn.Close()` 경로를 만들지 않는다.
- fallback은 하나뿐이고 명시적이다: 해석 실패 → nil publisher = 오늘 동작 (§0.2).

## Safety conclusion

- Safe edit boundary: `recordGateSettings` 인자 1개 추가, `gatewayInputs.publisher`
  값 1개 교체. 두 줄 모두 B4 이전/직후이고 분기 구조 밖이다.
- High-risk impact: **yes** — 다만 편집은 조립 인자 두 개이며 인터록·게이트·Guardian·
  계좌·원장 로직에 닿지 않는다.
- §0.2: 알림 설정이 없거나 off면 `resolveNotificationPublisher`가 nil을 반환하고
  `buildGateway`가 받는 값은 오늘과 같다.
- §0.7: 이 함수는 알림을 켜지 않는다. 설정 파일이 켠다.
- §0.8: 토큰은 이 함수에 값으로 들어오지 않는다 — `resolveNotificationPublisher`가
  환경에서 읽어 `obs.Ntfy`에 직접 넣고, 이 함수는 존재 여부(bool)만 audit로 넘긴다.
