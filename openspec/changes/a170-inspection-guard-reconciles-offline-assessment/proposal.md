# a170 — 검사 패키지 가드가 a162 오프라인 평가를 이름으로 받고, 그 평가의 런타임 소비자를 막는다

## Why

공유 브랜치의 `go test ./...` 가 빨갛다(a140 완료 gate 2 차, 2026-10-23 — a140 review §6.2):

```
--- FAIL: TestInspectionPackageHasNoRuntimeConsumers (0.28s)
    authority_guard_test.go:80: unexpected inspection import: internal/offlinecontinuation/assessment.go
FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/campaignstateevidence
```

원인 조사(2026-10-23):

- 가드(`internal/campaignstateevidence/authority_guard_test.go`)는 a161(`cd5f27219`, 2026-10-19)이 마지막으로 고쳤다. 의도는 a161 design Non-Goals
  「engine/CLI/API/UI 연결」 금지 — 캡처한 계획 · 위험 증거(`campaignstateevidence`)와 lane 별 `CaptureOfflineEvidence` 는 검사 전용이고 런타임 소비자가
  없어야 한다. 허용 목록은 lane 의 `offline_evidence.go` 셋뿐이다.
- a162(`3f8e34632`, 2026-10-20, `openspec/changes/archive/2026-10-20-a162-assess-follow-on-risk-offline/`)는 이 사용을 **설계로 의도했다**:
  design 「Input and authority boundary」 — 「동일 호출의 동일 native plan/risk에 기존 lane별 `CaptureOfflineEvidence`를 적용해 내부 조회 projection을 얻는다
  … projection을 외부로 반환하지 않는다」. 결과는 `inspection_only=true` · `execution_authority=false`(같은 design 「Calculation and result」)이고,
  `internal/offlinecontinuation/dependency_test.go` 가 journal · execgw · broker · engine · trading · gateway · net/http · os/exec import 를 막는다.
- a162 가 이 시험을 깨고도 archive 된 이유: a162 는 전체 suite · gate 를 사용자 명시 제한으로 돌리지 않았다(a162 review.md:34 · :118, tasks 3.3).
  가드는 a162 뒤에 바뀌지 않았다(`git log` — 가드 최종 커밋 `cd5f27219` < `3f8e34632`).
- `internal/offlinecontinuation` 을 import 하는 비시험 파일은 저장소에 0 개다(`internal` · `cmd` · `tools`, 2026-10-23 grep).

판정: **의도된 검사 전용 · 권한 없는 사용**이다. 가드의 허용 목록에 근거와 함께 그 파일 하나를 이름으로 더한다. 그런데 이름으로 받기만 하면
`offlinecontinuation` 을 import 하는 런타임 파일이 캡처 경로를 간접으로 얻는 문이 열린다. 그래서 같은 가드가 **`offlinecontinuation` 의 비시험
import 를 막는다** — 허용한 소비자 자신이 다시 런타임 소비자를 갖지 못하게.

## What Changes

- `authority_guard_test.go`(시험 파일만):
  1. `internal/offlinecontinuation/assessment.go` 를 evidence import 와 lane `CaptureOfflineEvidence` 사용 둘 다에서 이름으로 허용(같은 패키지의 다른 파일은 불허).
  2. 비시험 파일의 `internal/offlinecontinuation` import 를 거절.
  3. 시험: 허용 파일 원문은 통과 · 같은 원문을 다른 경로로 두면 거절 · 엔진 파일이 `offlinecontinuation` 을 import 하면 거절(이스케이프 표기 포함).
- 생산 코드 변경 0.

## Non-goals

- `internal/offlinecontinuation` · lane · `campaignstateevidence` 의 생산 코드 수정. a162 의 계산 · 경계 재설계.
- 가드의 다른 규칙(lane 점 import · 기존 허용 셋) 변경.

## Impact

- 위험도: Small(시험 전용). High-risk 경로 무접촉. Function Logic Map: not-applicable — 생산 함수 편집 0, 시험 파일의 가드 함수만 바꾼다.
- PM: `STORY-TOS-a170` → FEAT-TOS-013(a161 · a162 와 같은 Feature).
- 이 change 가 닫히면 a140 완료 gate 7 단계의 막힘이 풀린다(다른 이유가 없다면).
