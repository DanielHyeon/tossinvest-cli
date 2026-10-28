# Function Logic Map: `writeReconciled`

- Source: `internal/verifylive/report.go` (299-319)
- Qualified function: `writeReconciled`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` 57cd90ed8528a35e…)
- AST evidence: `ast.json` — AST branches 5, 반환 1(줄 301), 호출 이름 2
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: report·status 텍스트. a164 편집: artifact 마다 「reconciled absent」/「reconciled expired」 를 가름(B3).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `arts []Artifact` | `ReconciledArtifacts` 결과 | report.go | — |

## Branches and early returns

AST 열거 순서 그대로다. 「Required test」 칸은 손으로 고르지 않았다 — 시험 함수 하나씩 `-coverprofile` 을 떠서 분기 본문
블록이 실행된 시험만 적었다(측정 도구 `analysis/measure/` — `pertest.sh` · `btm_cov.py` · `gen_flm.py`).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 줄 300: `if len(arts) == 0 {` | 없음 | `return` | `TestProgressPointsAtTheRestartWhenOneIsPending` · `TestReportTextNamesTheUnverifiedProperties` |
| B2 | range 줄 305: `for _, a := range arts {` | 텍스트 한 줄씩 씀 | `if a.ReconciledExpired {` | `TestExpiryReconcileTextLabelsReconciledExpired` · `TestReconcileAbsenceKeepsItsA121Wording` · `TestReconcileTextLabelsReconciledAbsentInReportAndStatus` |
| B3 | if 줄 306: `if a.ReconciledExpired {` | 만료 라벨 | `basis = "공식 목록·상세 읽기 근거"` | `TestExpiryReconcileTextLabelsReconciledExpired` |
| B4 | range 줄 311: `for _, a := range arts {` | 없음 | `// a164: 부재(목록 두 번 읽기)와 만료(by-id 상세 두 번이 EXPIRED 로 이름)를 갈라 보임.` | `TestExpiryReconcileTextLabelsReconciledExpired` · `TestReconcileAbsenceKeepsItsA121Wording` · `TestReconcileTextLabelsReconciledAbsentInReportAndStatus` |
| B5 | if 줄 314: `if a.ReconciledExpired {` | 없음 | `ending = "reconciled expired"` | `TestExpiryReconcileTextLabelsReconciledExpired` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fmt.Fprintf` | 텍스트 | 오류 무시(기존 관례) | AST |

AST 호출 이름 전체: `fmt.Fprintf`, `len`.

## State mutations and fallbacks

- 쓰기만.

## Safety conclusion

- Safe edit boundary: 부재 artifact 의 줄 문구는 그대로.
- High-risk impact: 낮음.

## a164 GREEN 재추출

- 편집 전(구현 base `aec9bde67`) 분기 열 길이 2 에서 편집 뒤 5 로. `source_sha256` e96b2103cb0c… → 57cd90ed8528….
- 옛/새 분기 열(kind + 원문 줄)을 difflib 으로 정렬한 결과: equal 옛 B1..B2 = 새 B1..B2 · insert 옛 B3..B2(0) → 새 B3..B5(3).
- 위 표의 B 번호는 **새** AST 기준이고, 시험 칸은 GREEN 코드에서 시험별 프로필을 다시 떠 측정한 값이다(손 재번호 없음).
