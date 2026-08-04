# Function Logic Map: `mergeEngine`

- Source: `internal/config/engine.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a074-critical-events-reach-the-operator/base-commit.txt`
- 위험 등급: Normal — 순수 파싱이다. 다만 이 함수가 **자동화 게이트의 값을 결정**하므로
  기존 분기를 흔들면 게이트 의미가 바뀐다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `raw *rawEngine` | nil이면 아무것도 안 함 | 설정 파일 | 기본값(전부 off) 유지 |
| `raw.Autostart` | `*bool`, nil이면 미변경 | 파일 | — |
| `raw.Adoption` | nil이면 미변경 | 파일 | 거부 시 블록 전체 0 |
| `raw.ExitPolicy` | nil이면 미변경 | 파일 | 거부 사유를 `Rejected`에 |
| `raw.AutomationGate` | nil이면 **조기 반환** | 파일 | — |

**불변식 (유지)**: OFF가 기본이다. 파일 헤더가 명시한다 — "The zero value of every field
is therefore the safe one … There is no 'unset means on' anywhere in this file, and no
default that opens anything."

**불변식 (유지)**: 거부된 블록은 **통째로 0**이다. `mergeAdoption`의 주석이 이유를
적는다 — "a block that kept `enabled` while dropping the fraction would be adoption
running on a stop nobody chose."

**a074가 바꾸는 것**: `rawEngine`에 `Notifications *rawNotifications`를 더하고,
`mergeEngine`에 그것을 병합하는 분기 하나를 더한다. 새 분기는 `AutomationGate`의
조기 반환(B4) **앞**에 온다 — 뒤에 두면 automation_gate가 없는 설정 파일에서 알림
블록이 조용히 무시된다.

**a074가 바꾸지 않는 것**: 기존 다섯 분기의 조건·순서·대입. AutomationGate가 마지막에
남는다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (305) | `raw == nil` | 없음 | 조기 반환 | 기존 |
| B2 (308) | `raw.Autostart != nil` | `cfg.Autostart` | — | 기존 |
| B3 (312) | `raw.ExitPolicy != nil` | `cfg.ExitPolicy` (+ 거부 사유) | — | 기존 |
| **신규** | `raw.Notifications != nil` | `cfg.Notifications` (+ 거부 시 0) | — | **4.1–4.4** |
| B4 (317) | `raw.AutomationGate == nil` | 없음 | 조기 반환 | 기존 |
| B5 (321) | `gate.Enabled != nil` | `cfg.AutomationGate.Enabled` | — | 기존 |

`mergeAdoption` 호출(311행)은 분기가 아니라 문이므로 AST 분기표에 없다. 그 호출 순서도
바뀌지 않는다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `mergeAdoption` | adoption 블록 병합 | 거부 시 블록 0 | AST |
| `ExitPolicy.validate` | 정책 id 확인 | 사유 문자열 | AST |
| `strings.TrimSpace` | 정규화 | — | AST |
| `mergeNotifications` (신규) | 알림 블록 병합 | 거부 시 블록 0 | 신규 |

**환경변수를 읽지 않는다** (design D4). `internal/config`는 파일만 해석한다. 토큰과
topic 오버라이드의 해석은 조립 지점의 일이다 — 그래야 이 패키지의 테스트가 프로세스
환경에 의존하지 않는다.

## State mutations and fallbacks

- `cfg` 포인터의 필드만 쓴다. 파일 I/O도 네트워크도 없다.
- 새 필드의 zero value는 `Notifications{}`이고 그것은 `Enabled=false`다 → publisher nil
  → 오늘 동작 (§0.2).
- 알림 블록이 없는 기존 설정 파일은 `raw.Notifications == nil`이므로 새 분기를 지나가지
  않는다. 파싱 결과가 편집 전과 **바이트 단위로 같다.**

## Safety conclusion

- Safe edit boundary: `rawEngine` 구조체에 필드 하나, `mergeEngine`에 B4 앞 분기 하나,
  새 함수 `mergeNotifications`.
- High-risk impact: **no** — 주문·손절·사이징·원장 경로에 닿지 않는다. 자동화 게이트
  분기는 편집하지 않는다.
- §0.2: `notifications` 키가 없는 모든 기존 설정이 편집 전과 동일하게 로드된다.
- §0.8: 토큰 필드가 **없다.** 파일에 토큰을 담을 자리를 만들지 않는 것이 이 설계의 일부다.
