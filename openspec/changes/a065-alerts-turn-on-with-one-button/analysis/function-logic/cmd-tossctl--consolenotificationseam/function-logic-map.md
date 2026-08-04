# Function Logic Map: `consoleNotificationSeam`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a065-alerts-turn-on-with-one-button/base-commit.txt`
- 위험 등급: Normal — 능력 하나를 화면에 건네거나 건네지 않는다.

## Inputs and invariants

새 함수다. 기존 파일 안에 있으므로 `check_analysis.py`가 증거를 요구한다 — 그리고
요구하는 것이 맞다: 이 여섯 줄이 콘솔이 알림 설정을 만질 수 있는지 없는지를 정한다.

같은 파일의 다른 seam 어댑터 열 몇 개와 **정확히 같은 모양**으로 썼다. 그 모양의
이유는 파일 주석에 이미 있다.

- 이 파일만 `internal/console`을 import할 수 있으므로, 인터페이스를 이름으로 부르는
  어댑터는 전부 여기 산다. 구현(`notificationsettings.go`)은 소비자를 모른다.
- nil 판정은 **구체 포인터** 위에서 한다. 구현을 그대로 반환하면 프로필이 없을 때
  인터페이스 안에 typed nil이 들어가고, 화면은 그것을 배선된 것으로 읽어 첫 클릭에서
  깨진다.

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root *rootOptions` | 임의 | 플래그 | 프로필 미해석 → `nil` 인터페이스 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (628) | `newNotificationSeam(root) != nil` | 없음 | 구현 반환, 아니면 `nil` | 6.6, 7.10 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newNotificationSeam` | 프로필의 config service로 seam 조립 | nil = 프로필 미해석 | AST |

네트워크도, 파일 I/O도, 난수도 여기서 일어나지 않는다. 채널 생성은 `Enable`이
눌렸을 때 seam 안에서 일어나고, 전송은 `Test`에서 일어난다.

## State mutations and fallbacks

- 상태를 만들지 않는다. 조립만 한다.
- nil을 반환하면 알림 카드가 "미배선" 사유를 렌더하고 버튼을 렌더하지 않는다
  (`TestAnUnwiredAlertSeamSaysWhyInsteadOfOfferingAButton`).

## Safety conclusion

- Safe edit boundary: 새 함수 여섯 줄.
- High-risk impact: **no** — 주문·손절·사이징·Guardian·원장·인증·체결 경로에 닿지 않는다.
- §0.2: 프로필을 해석하지 못하는 실행에서 이 함수는 nil을 주고, 화면은 편집 전과 같다.
