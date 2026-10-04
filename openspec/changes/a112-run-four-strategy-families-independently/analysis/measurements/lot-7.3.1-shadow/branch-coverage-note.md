# 분기 진입 측정 노트(a112 7.3.1 SHADOW 로트)

- 하네스: `analysis/harness/branch_coverage.py`, 사본 = HEAD 9e5f3ccf `git archive` + 이 로트 파일(병행 편집 배제).
- 엔진(`branch-coverage-engine.json`): 태그 빌드, shadow 관련 시험 29 개를 하나씩 + 패키지 합집합 1 회. 합집합 실행은 `union_passed=false` —
  실패 둘은 `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` · `TestA111FallbackSequenceRecoveryIsLazyAndPriceEvidenceUsesTheGateDuration` 이고,
  하네스가 `GOFLAGS=-trimpath` 로 돌리는데 그 둘은 소스 경로를 읽는 a111 시험이라 `-trimpath=false` 가 필요하다(이 로트와 무관, 일반 스위트에서는 PASS).
  합집합 프로필은 그대로 기록됐으므로 분기 행의 「합집합 진입」 은 유효하다.
- 투영(`branch-coverage-projection.json`): 무태그, `Shadow|Validate|Clone` 시험 7 개 + 합집합. `union=None` 은 여러 줄 조건이라 블록을 같은 줄에서
  못 찾은 경우다(진입 여부 미측정 — 행에 「측정 없음」 으로 적힘).
- 행에 「진입 0」 은 커버리지 공백이지 통과가 아니다.
