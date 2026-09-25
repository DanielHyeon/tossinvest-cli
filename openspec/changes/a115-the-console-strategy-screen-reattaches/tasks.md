# Tasks — a115-the-console-strategy-screen-reattaches

**a109 review §2 A2 P1-1(선언된 생략)의 이행. 등록 2026-08-16, 착수 2026-09-25~26.**

## 0. 착수 전 조건

- [x] 0.1 a114와의 합본 여부 판단 — **합본하지 않는다**(Manager 결정 2026-09-25): a114 design 0.1 과 같은
      근거(Story↔change 1:1, 같은 `runConsole` 의 다른 블록·다른 dial, 순차 구현). a115 base 는 a114
      착지·기록 뒤의 `8688f74f`. design 0.1.
- [x] 0.2 base 재고정 — `8688f74f`(a114 착지 뒤, `base-commit.txt`).
- [x] 0.3 콘솔 전략 dial 과 접힘 지점, 소비 page 의 Function Logic Map 을 proposal 갱신 전에 만든다 —
      `runConsole` · `resolveStrategyRuntimeReader` · `strategyRuntimeAttachment.StrategyRuntimeConfigured` ·
      `buildMultiMarketStrategyRuntimePage` · `strategyRuntimeSummary`(`analysis/function-logic/` 5벌) +
      CodeGraph 증거 조정(`analysis/code-context/` 3파일, codegraph 1.6.0).
- [x] 0.4 design.md + proposal-freeze 리뷰(review.md §0) + 판결 반영. 경량 등급(콘솔 화면·주문 경로 무관)
      + 독립 적대 보이스 1(별도 컨텍스트, 읽기 전용).

## 1. 구현 (Teammate)

- [x] 1.1 RED — internal/console: 부재 신호 false 인 non-nil reader → dormant(Read 0회) · 신호 true + Read
      실패 → 도달 불가(NOT_CONFIGURED 아님) · 요약도 같은 판정 · 신호 없는 reader 는 오늘처럼 wired.
      cmd/tossctl: 진짜 `strategyprojectionrpc.Start` 로 ① 구성·엔진 다운(죽은 descriptor+socket) → 도달
      불가 → 엔진 기동 후 회복 ② 미구성 → dormant → 늦은 엔진 → 회복 ③ 렌더 없이 펌프로 부착 ④ 재시작
      재부착 ⑤ `runConsole` 이 `strategyprojectionrpc.Dial` 을 직접 부르지 않음(AST — base 에서 RED).
      추가(freeze 리뷰): 판정 동치(strategyprojection 판정 vs httpapi 위임, 세 상태 + 신호 없는 reader) ·
      콘솔 해석과 httpapi `resolveStrategyRuntimeReader` 의 판정 동치 · 재시도 경고 discard · 새 부팅 문구가
      dormant 를 말하지 않음 · 펌프 interval≤0 가드 · 펌프 goroutine 정리(cancel + inFlight 대기).
- [x] 1.2 GREEN — 새 파일 `cmd/tossctl/console_strategy_attach.go`(`resolveConsoleStrategyRuntime` +
      `consoleStrategyRuntimeReaderFor`, wrapper 재사용 + 무조건 wake 펌프·콘솔 전용 간격 변수) + runConsole
      블록 교체. 부재 판정·presence 를 `internal/strategyprojection` 으로 이동(httpapi 는 alias·위임 — 판정
      한 벌, design D2) + `internal/console` 두 소비자의 nil 판정 교체 + `cmd/tossctl` 컴파일 결속 `var _`.
- [x] 1.3 뮤테이션(사본 · 무변이 대조군 선행): nil 접힘 재도입 · 화면 nil 판정 복귀(두 자리 각각) ·
      부재 신호 무시 · 펌프 제거 · 부팅 dial 재도입 · sentinel 대신 nil → 각각 빨강. 원장 `mutation-ledger.md`.
- [x] 1.4 FLM 구현 후 재최신화(`revision: current` 재추출, Branch Test Map 재번호) + `check_analysis.py` rc 0.
- [x] 1.5 검증: cmd/tossctl·internal/console `-race` 대상 테스트 · `make vet` · `make lint` · `make test`.
- [x] 1.6 콘솔 실측(규칙 13): 엔진 없이 두 절반 — ① 깨끗한 config → dormant 표기 ② 잔재 descriptor +
      죽은 socket → 도달 불가 표기. 버튼 누르지 않음. issues R2.
- [ ] 1.7 구현 후 리뷰(gstack/독립 적대) → review.md §1.
- [ ] 1.8 착지 기록 커밋(`--record-landing`).

## 2. 게이트 (Manager)

- [ ] 2.1 Manager 독립 검증(diff·테스트 재실행) 후 `make gate CHANGE=a115-the-console-strategy-screen-reattaches`,
      archive · Story 경로 · PM `--check`.
