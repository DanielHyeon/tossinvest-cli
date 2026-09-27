# Function Logic Map: `strategyFamilyGate.installed`

- Source: `internal/app/engine/strategy_family_activation.go` (57-59)
- Function: `strategyFamilyGate.installed` in package `engine`
- File SHA-256: `4a1ac0a98c8a7c598c5850a3c006e81821c56bb9e11aad113d9b943b4ec59022`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file (pre-edit).
- AST evidence: `ast.json` — AST branches 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

**편집 전 기록 (태스크 8.7.2, 2026-09-27).** 이 번들은 편집 **전** 소스의 AST 다. 8.7.1 이 만든 신규 함수라 번들이 없었고, 8.7.2 가 내부를 바꾸므로 여기서 처음 만든다.

입력은 수신자 하나(`gate`)다. 반환식은 `gate.activation.Verified() && len(gate.lanes) != 0` 한 줄이고 AST 분기는 0 이다.
불변식(편집 전): 관문이 판정 주체인지는 **검증된 활성화 + 레인** 두 사실의 곱이다. 어느 하나라도 없으면 false →
`admit` B1(72:2)이 묻지 않고 통과시킨다.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode. arm = 분기 좌표 **뒤에서 처음 시작하는** 커버리지 블록(`if`/`range` 몸통)이며, "arm entered Nx" 는 그 몸통이 N 번 실행됐다는 뜻이다.
- engine tagged suite: `go test -c -tags tossos_testseams -covermode=count -coverpkg=./internal/app/engine,./internal/strategyrouter ./internal/app/engine/` 바이너리를 `systemd-run --user --scope -p MemoryMax=16G -p MemorySwapMax=0` 안에서 실행(-trimpath 없이 — 소스를 읽는 시험 둘이 깨진다). 스위트 전체 PASS.
- Per-test attribution set: 같은 바이너리의 **전체 시험 목록**을 `-test.run '^<Test>$'` 로 하나씩 돈 프로파일(하네스 `analysis/harness/a872_pertest_cover.sh`, 표 생성 `analysis/harness/a872_attribute.py`).
- **귀속 완전성은 등식이다**: 모든 행에서 시험별 진입 수의 합 == 스위트 진입 수. 깨진 행은 `ATTRIBUTION MISMATCH` 로 찍히며 아래에는 하나도 없다.
- 이 regime 은 **몸통 진입**을 센다. 같은 change 의 옛 번들 일부(예: dispatch 의 5.x 표)는 조건 평가를 센 값이라 수가 다르다 — 섞어 읽지 말 것.

Exact AST return positions: 58:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
_(AST 분기 0 — 반환식 하나. 행동은 아래 Inputs 의 불리언 식이 전부다.)_



## Calls and live bindings



## State mutations and fallbacks

없음. 읽기 전용 술어다.

## Safety conclusion

- **편집 전 사실(AST로 확인):** 활성화가 검증되지 않으면 이 함수는 false 이고, 그러면 `admit` 은 `(빈 봉투, "", true)` 로 통과시킨다(admit R 73:3). 즉 활성화의 부재·만료·폐기는 모두 "관문 없음"이고, 그 시장은 기존 시장 단위 경로로 돈다.
- **8.7.2 편집 계획:** 배포가 핀으로 활성화를 선언했는데 검증된 활성화가 없는 상태(rolled back)에서는 관문이 **서야** 한다. 그래서 반환식에 `gate.rolledBack ||` 를 더한다. rolled back 관문은 레인이 없어도 선다 — 레인이 없으면 `admit` 이 주인 없는 제안으로 REFUSED 를 내고, 그것이 이 상태에서 맞는 값이다(통과가 아님).
- High-risk impact: yes (진입 관문).
