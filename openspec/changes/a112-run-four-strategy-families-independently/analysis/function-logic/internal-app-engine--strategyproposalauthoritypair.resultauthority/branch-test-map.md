# Branch Test Map: `strategyProposalAuthorityPair.ResultAuthority`

- Source: `internal/app/engine/strategy_proposal_authority.go` (144-152); file SHA-256 `b7a565a767a7ef790ff1390a5db4c5e83cac19897d408aae167d7243466b3d38`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.

- Measurement regime: Go coverage profiles, count mode.
- 모든 실행은 `systemd-run --user --scope -p MemoryMax=… -p MemorySwapMax=0` cgroup 안에서 돌렸다.
  묶지 않고 돌린 이 패키지가 커널 OOM 으로 데스크톱을 세 번 죽였기 때문이다(`engine.test`, anon-rss 약 36GB).
  원인은 이 lot 이 고친 조정자 용량이며, 측정 방법이 아니라 측정 대상이 문제였다.
- engine tagged suite: `go test -c -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine ./internal/app/engine/`
  뒤에 그 바이너리를 `-test.coverprofile` 로 실행. 스위트 전체 PASS, 73.8% of statements.
- engine untagged suite: 같은 명령에서 `-tags tossos_testseams` 만 뺀 것. 스위트 전체 PASS, 63.6% of statements.
- Per-test attribution set: 같은 태그 바이너리를 `-test.run '^<Test>$'` 로 하나씩 돌린 열 개의 프로파일.
- **귀속 완전성은 주장이 아니라 측정이다.** 아래 모든 분기에서 테스트별 진입 수의 합이 스위트 전체 진입 수와
  정확히 같다. 이 집합 밖의 테스트가 어느 arm 이든 들어갔다면 그 등식이 깨진다. 깨진 행은
  `ATTRIBUTION MISMATCH` 로 표시되며 아래에는 하나도 없다.

이 표는 5.4.1 판본의 옛 형식(다른 `-coverpkg`, 일곱 테스트 귀속 집합)을 그대로 두지 않고
5.4.2 의 측정으로 다시 만든 것이다. 함수 본문은 이 lot 에서 바뀌지 않았지만 파일이 바뀌어
줄 번호와 파일 해시가 움직였고, 옛 표가 인용하던 분기 위치는 더 이상 이 파일의 어떤 위치도 아니었다.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 146:3 | arm entered 1x (engine tagged suite); arm not entered (engine untagged suite); `TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
