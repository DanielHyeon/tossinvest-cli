# Function Logic Map: `Console.handlePositionManagement`

- Source: `internal/console/position_policy.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a063-operator-can-lift-a-quarantine/base-commit.txt`
- 위험 등급: Normal (읽기 전용 렌더). mutation 없음. Pre-Edit 선언은 `review.md`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.opts.PositionPolicies` | nil 가능 | 콘솔 배선 | nil이면 `Wired=false`로 조회 전용 렌더 후 return |
| `c.opts.Settings` | nil 가능 | 콘솔 배선 | nil이면 `DesiredErr` 문구 |
| `PositionPolicies.Runtime` | 오류 가능 | 엔진 RPC | `RuntimeErr`에 담고 계속 렌더 |
| `PositionPolicies.List` | 오류 가능 | 엔진 RPC | `LoadErr` 렌더 후 return |
| `c.holdings.peek(now)` | 항상 사용 가능 | 콘솔 캐시 | 이름 없으면 빈 문자열 (a061) |

**라우트**: `/position-management` GET, `c.session0(c.readOnly(...))` (console.go:779).
`readOnly` 아래이므로 이 핸들러는 **아무것도 mutate하지 않는다**.

**a063이 추가하는 것**: 격리 capability가 발견되면 활성 격리 목록을 한 번 읽어
`policyRowView.Quarantine` badge와 해제 action token을 채운다. 실패는 **행을 그리는
것을 막지 않는다** — 격리 조회가 실패했다는 사실만 표시하고 나머지 화면은 그대로다.
읽기 전용 성질은 유지된다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (208) | `c.opts.Settings == nil` | `page.DesiredErr` | 계속 | 기존 `position_policy_test.go` |
| B2 (210) | 위의 else | — | 계속 | 기존 |
| B3 (210) | `Settings.Load()` 오류 | `page.DesiredErr` | 계속 | 기존 |
| B4 (212) | Load 성공 | `page.Desired`, `DesiredVerdict` | 계속 | 기존 |
| B5 (216) | `PositionPolicies == nil` | 없음 | 렌더 후 **return** | `TestAnUnwiredCommanderShowsNoReleaseAction` |
| B6 (221) | `Runtime()` 오류 | `page.RuntimeErr` | 계속 | 기존 |
| B7 (223) | Runtime 성공 | `Effective`, `BlockSource` | 계속 | 기존 |
| B8 (226) | `runtime.Blocks` 순회 | `page.Blocks` 누적 | 계속 | 기존 |
| B9 (231) | `List()` 오류 | `page.LoadErr` | 렌더 후 **return** | 기존 |
| B10 (237) | `states` 순회 | 행 조립 | 계속 | `TestTheConsoleOffersReleaseOnlyForAQuarantinedRow` |
| B11 (246) | `management.Block != nil` | `row.Block` | 계속 | 기존 |
| B12 (249) | `Status == MANAGED` | 정책 action 추가 | 계속 | 기존 |
| B13 (258) | B12의 else | — | 계속 | 기존 |
| B14 (251) | 등록된 공통 정책 순회 | override action 누적 | 계속 | 기존 |
| B15 (255) | `ExternalLifecycleEligible()` | 「자동관리 해제」 추가 | 계속 | 기존 |
| B16 (258) | `RELEASED && ExternalLifecycleEligible()` | 「새 generation 재편입」 추가 | 계속 | 기존 |

**a063이 추가하는 분기**: 격리 capability 탐지 1개, 격리 조회 오류 1개, 행별
"이 포지션에 활성 격리가 있는가" 1개. 최종 번호는 구현 후 재생성한 AST로 갱신한다.

**중요**: 새 분기는 B15/B16(lifecycle action)과 **독립**이다. 격리 해제는
`ExternalLifecycleEligible()`을 요구하지 않는다 — 격리는 편입 출처와 무관하게
어떤 관리 포지션에나 생길 수 있고, 해제는 lifecycle을 건드리지 않기 때문이다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.opts.Settings.Load` | desired adoption 설정 | 오류는 문구로 표시, 렌더 계속 | AST |
| `PositionPolicies.Runtime` | 엔진 실효 설정 | 오류는 `RuntimeErr`, 렌더 계속 | AST |
| `PositionPolicies.List` | 정책 행 목록 | 오류는 렌더 후 return | AST |
| `positionpolicy.ProjectManagement` | 표시용 투영 | 순수 | AST |
| `c.holdingNames(c.now())` | 종목명 (a061) | broker 호출 0 — 캐시 peek | AST |
| `c.policyAction` | 서명된 1회 select token | 순수 + HMAC | AST |
| `c.render` | 템플릿 렌더 | — | AST |
| **(a063)** 격리 목록 조회 | badge/action 재료 | 오류는 문구로 표시, 렌더 계속 | 신규 |

## State mutations and fallbacks

- 원장·파일·설정에 아무것도 쓰지 않는다. `readOnly` 미들웨어 아래다.
- fallback 방향: 어떤 보조 읽기가 실패해도 화면 자체는 뜬다. 정책 목록 실패만
  조기 return이다. a063의 격리 읽기는 **보조**이므로 실패해도 조기 return하지 않는다.

## Safety conclusion

- Safe edit boundary: 행 조립 구간과 page 구조체 필드.
- High-risk impact: **no** — 읽기 전용이다. 해제 mutation은 별도의 POST 핸들러가
  담당하고 그것은 신규 파일이다.
- §0.4 rate budget: 신규 broker 호출 없음. 격리 조회는 엔진 loopback RPC이며
  공식 API가 아니다.
