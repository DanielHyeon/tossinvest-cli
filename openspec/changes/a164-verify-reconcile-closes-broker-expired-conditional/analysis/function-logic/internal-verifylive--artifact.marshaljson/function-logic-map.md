# Function Logic Map: `Artifact.MarshalJSON`

- Source: `internal/verifylive/reconcile_record.go` (110-128)
- Qualified function: `Artifact.MarshalJSON`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` 01f19dd45c00351b…)
- AST evidence: `ast.json` — AST branches 1, 반환 2(줄 113, 115), 호출 이름 2
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: 대사된 artifact 의 확정형 직렬화. a164 는 넷째 종결도 확정형으로(보유·취소 필드 없이) 씀.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 수신자 `a Artifact` | — | `record.go` | json 오류만 |

## Branches and early returns

AST 열거 순서 그대로다. 「Required test」 칸은 손으로 고르지 않았다 — 시험 함수 하나씩 `-coverprofile` 을 떠서 분기 본문
블록이 실행된 시험만 적었다(측정 도구 `analysis/measure/` — `pertest.sh` · `btm_cov.py` · `gen_flm.py`).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 줄 112: `if !a.ReconciledAbsent && !a.ReconciledExpired {` | 없음 | `return json.Marshal(plainArtifact(a))` | `TestADeleteNotFoundFromEveryListReadDoesNotReconcile` · `TestAFilledFieldSurvivesTheRecordRoundTrip` · `TestAFullRunProducesAReportWithNoIdempotencyGaps` · `TestAnOlderBinaryStillSeesTheExpiredArtifactAsOutstanding` 외 92 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `json.Marshal` | 직렬화 | 오류 반환 | AST |

AST 호출 이름 전체: `json.Marshal`, `plainArtifact`.

## State mutations and fallbacks

- 상태 변경 없음. 대사되지 않은 artifact 는 별칭 타입으로 base 와 바이트 동일(`TestUnreconciledArtifactsSerialiseExactlyAsTheBase`).

## Safety conclusion

- Safe edit boundary: 조건을 `ReconciledAbsent || ReconciledExpired` 로 넓히고 `reconciled_absent` 를 omitempty 로, `reconciled_expired` 를 omitempty 로 더함 — 부재 줄 바이트 불변.
- High-risk impact: 인접 — 확정형이 HeldUntil 을 실으면 구 바이너리 재보유 모양이 바뀜.

## a164 GREEN 재추출

- 편집 전(구현 base `aec9bde67`) 분기 열 길이 1 에서 편집 뒤 1 로. `source_sha256` 84baeb1173cc… → 01f19dd45c00….
- 옛/새 분기 열(kind + 원문 줄)을 difflib 으로 정렬한 결과: replace 옛 B1..B1(1) → 새 B1..B1(1).
- 위 표의 B 번호는 **새** AST 기준이고, 시험 칸은 GREEN 코드에서 시험별 프로필을 다시 떠 측정한 값이다(손 재번호 없음).
