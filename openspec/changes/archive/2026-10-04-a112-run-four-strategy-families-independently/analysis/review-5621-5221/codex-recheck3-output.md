**BLOCK — 소스 변경 검출은 성립하지만, 시험 파일·digest를 고치지 않고 동결 시험 실행 자체를 건너뛰는 경로가 남습니다.**

이번 재확인에서 `~/.codex` 접근 없음. 현재 digest는 직접 재계산해 일치를 확인했습니다. Go 시험·변이는 재실행하지 않았으며, 아래 무력화 반례는 정적 판정입니다.

`H` = `internal/strategyhandoff/`, `C` = `openspec/changes/a112-run-four-strategy-families-independently/`.

| 검토 항목 | 판단 · 파일:줄 근거 |
|---|---|
| 생산 `.go`·파일 이름·정규화 | **충족.** `H/source_freeze_test.go:49–72`는 정렬된 `*.go`에서 `_test.go`만 제외하고 파일 이름·NUL·gofmt 정본·NUL을 해싱한다. 현재 대상은 `handoff.go`, 재계산 결과는 고정값 `sha256:1ad10787…4584b22`와 일치(`:45`). |
| 빌드 태그 파일 | **포함.** 빌드 조건을 평가하지 않고 파일을 읽으므로 비활성 태그·다른 OS용 `.go`도 digest가 바뀐다(`:49–64`). |
| 비`.go`·cgo | **비`.go`는 동결 밖.** 현재 소스에는 비Go 파일을 읽는 embed나 cgo 연결이 없다. `import "C"` 추가는 Go digest와 직접 import 허용 목록 모두에 걸린다(`H/dependency_closure_test.go:25–61`). 다만 이 검사는 `.s`·`.syso` 등 모든 빌드 입력의 부재를 보증하지 않는다. |
| 다른 패키지에서 주조 | **일반적인 import 경유 우회는 별도 검사 대상.** `H/dependency_closure_test.go:179–235`가 생산·시험 의존 폐포에서 engine 이외의 importer를 거절한다. 이는 모든 `unsafe`·링커 경로를 보증하는 검사는 아니다. |
| digest 동시 갱신의 가시성 | **논증 타당, 절차적 조건부.** 고정값 변경은 명시적 재고정이며 이번 과녁 밖이라는 구분은 타당하다(`H/source_freeze_test.go:11–17`). 다만 digest 일치 자체가 독립 리뷰 수행을 증명하지는 않는다. |
| 기존 실증 반례 | **동결 시험이 실행되면 검출.** `C/analysis/measurements/lot-5.6.2-5.2.2/mutation-5.2.2.1-fix4.tsv:3–6`에 H09~H12를 동결 시험만 검출했다고 기록. 현재 코드의 해시 비교(`H/source_freeze_test.go:77–86`)와 부합한다. |
| 시험 파일 편집 없는 무음 무력화 | **미충족.** 생산 패키지의 `init()`에서 시험 바이너리만 종료 코드 0으로 끝내면 digest·import 폐쇄 시험 모두 실행되지 않는다. 아래 반례. |

새 생산 파일에 기존 주조 반례와 함께 다음을 추가할 수 있습니다. 시험 파일과 `frozenSourceDigest`는 그대로입니다.

```go
package strategyhandoff

import (
    "os"
    "strings"
)

func init() {
    if strings.HasSuffix(os.Args[0], ".test") {
        os.Exit(0)
    }
}
```

이 파일은 **해시 대상이지만 해시 계산 전에 실행됩니다.** 일반 생산 바이너리에서는 종료하지 않으므로 함께 추가한 둘째 주조 문은 생산에서 남습니다. 금지된 `os` import도 폐쇄 시험이 실행되지 않아 검출되지 않습니다.

현재 설치된 Go 소스로 확인한 근거입니다.

- `os.Exit(0)`의 시험 중 panic 보호는 `testing.M.before()`에서 활성화됩니다: `/usr/local/go/src/testing/testing.go:2434,2677–2678`. 패키지 초기화는 그보다 먼저입니다.
- 보호가 꺼져 있으면 `os.Exit`은 그대로 종료합니다: `/usr/local/go/src/os/proc.go:62–77`.
- Go 실행기는 프로세스 성공 종료를 `ok`로 보고하며, 이 지점에서 개별 시험 실행을 요구하지 않습니다: `/usr/local/go/src/cmd/go/internal/test/test.go:1712–1732`.
- 변이 하네스도 종료 성공을 GREEN으로 분류합니다: `C/analysis/harness/a112_lot_mutate.py:238–257`.

따라서 정확한 보증은 **“동결 시험이 실제로 실행되면 정본 소스 변경을 검출한다”**입니다. “생산 소스 편집만으로는 무음 우회가 불가능하다”는 더 강한 주장은 성립하지 않습니다. 현재 트리에 위 종료 코드가 있다는 뜻은 아닙니다.

Manager 지시에 따라 이 발견과 비Go 빌드 입력·외부 주조 경로의 한계는 **5.2.2.2 / 6.2로 이월**합니다. 동결 방식 자체는 기존 모양 census보다 명확한 검출 장치입니다.

Recommendation: BLOCK — 동결 시험의 실제 실행을 외부에서 확인한다는 조건 없이 종결을 선언할 수 없음. 초기화 조기 종료 반례와 실행 증거 검증을 5.2.2.2 / 6.2로 이월.
