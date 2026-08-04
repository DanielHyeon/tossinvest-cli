# Function Logic Map: `consoleGateSwitchSeam`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a065-alerts-turn-on-with-one-button/base-commit.txt`
- 위험 등급: Normal

## Inputs and invariants

이 함수는 **편집되지 않았다.** `consoleNotificationSeam`을 바로 아래에 삽입했고,
`git diff --unified=0`이 만든 hunk의 anchor가 이 함수의 닫는 괄호 다음 줄이라
`check_analysis.py`의 `intersects`(count==0일 때 `start <= line <= end + 1`)가
이 함수를 "수정됨"으로 판정한다. 증거를 만드는 것이 그 판정을 부인하는 것보다 싸고,
인접 삽입이 정말로 무해한지는 이 표를 쓰면서 실제로 확인된다.

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root *rootOptions` | 임의 | 플래그 | 프로필 미해석 → nil |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (627) | `newGateSwitchSeam(root)`가 nil이 아님 | 반환값을 인터페이스로 | 구현 또는 nil | 기존 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newGateSwitchSeam` | 게이트 스위치 저장 seam | nil = 미배선 | AST |

## State mutations and fallbacks

- 상태를 만들지 않는다. 구체 포인터의 nil을 인터페이스 밖에서 판정해 typed nil을 막는 것이 전부다.
- a065는 이 함수의 바이트를 바꾸지 않았다 — `git diff`의 hunk는 이 함수 **밖**에서 시작한다.

## Safety conclusion

- Safe edit boundary: 없음 — 편집되지 않았다.
- High-risk impact: **no**.
- 인접 삽입이 이 함수의 어떤 줄도 바꾸지 않았음은 `git diff`의 hunk 좌표(`@@ -627,0 +633,10 @@`)가 증명한다: 삭제 0줄.
