# Tasks — a114-the-console-reattaches-its-lifecycle

**a109 design 선언된 생략(P2-7)의 이행. 등록 2026-08-16, 착수 2026-09-25.**

## 0. 착수 전 조건

- [x] 0.1 a115와의 합본 여부 판단 — **합본하지 않는다**(Manager 결정 2026-09-25): Story↔change 1:1 을
  유지하고, 같은 `runConsole` 이지만 다른 블록(lifecycle B33–B37 / 전략 B38–B42)·다른 dial 이며 같은
  Teammate 가 a114 → a115 순차 구현(a115 base 는 a114 착지 뒤)이라 충돌이 없다. design 0.1.
- [x] 0.2 base 재고정 — `634cf3c5`(a113 착지 뒤).
- [x] 0.3 콘솔 부팅 경로(lifecycle client dial)와 그 소비 화면의 Function Logic Map 을 proposal 갱신 전에
  만든다 — `runConsole` · `quarantineClient` · `handlePositionManagement` · `decoratePositionRows`.
- [x] 0.4 design.md + proposal-freeze 독립 적대 리뷰(review.md §0) + 판결 반영.

## 1. 구현

- [x] 1.1 RED: `runConsole` 이 lifecycle 을 직접 dial 하지 않음(AST) + 엔진이 콘솔보다 늦게 뜨는 순서에서
  재시작 없이 붙음·가동 중 재시작 재부착·기전 테스트(새 심볼 — 컴파일 RED).
- [x] 1.2 GREEN: 새 파일 `console_lifecycle_attach.go`(wrapper + `consolePositionPolicyCommanderFor`) +
  runConsole 블록 한 줄.
- [x] 1.3 뮤테이션: a109 T2 원장의 wrapper 뮤테이션 중 콘솔 적용판 + 콘솔 고유(원장 `mutation-ledger.md`).
- [x] 1.4 FLM 구현 후 재최신화 + `check_analysis.py` rc 0.
- [x] 1.5 검증: cmd/tossctl `-race` 대상 테스트 · `make vet` · `make lint` · `make test`.
- [x] 1.6 콘솔 실측(규칙 13): 격리 config·엔진 없이 콘솔을 띄워 정책 화면 본문 확인(버튼 누르지 않음).
- [x] 1.7 구현 후 리뷰 → review.md §1.
- [x] 1.8 착지 기록 커밋.

## 2. 게이트 (Manager)

- [x] 2.1 `make gate CHANGE=a114-the-console-reattaches-its-lifecycle` 후 archive.
