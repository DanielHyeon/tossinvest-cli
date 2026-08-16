# Tasks — a114-the-console-reattaches-its-lifecycle

**이 change는 a109 design 선언된 생략(P2-7)의 이행으로 등록만 먼저 했다(2026-08-16).**

## 0. 착수 전 조건

- [ ] 0.1 a115와의 합본 여부 판단 — 같은 `cmd/tossctl/console.go` 부팅 경로를 편집한다.
  합치면 한쪽을 `openspec archive --skip-specs`로 정리하고 사유를 남긴다.
- [ ] 0.2 base 재고정(`tools/sdd/capture_change_base.py --change a114-…`).
- [ ] 0.3 콘솔 부팅 경로(lifecycle client dial)와 그 소비 화면의 Function Logic Map을
  proposal 갱신 전에 만든다.

## 1. 구현 (착수 시 상세화)

- [ ] 1.1 RED: 엔진이 콘솔보다 늦게 뜨는 순서에서 콘솔이 재시작 없이 붙는 것을
  요구하는 테스트.
- [ ] 1.2 GREEN: httpapi 재부착 계약 이식(백그라운드 single-flight·요청 경로 dial 금지).
- [ ] 1.3 뮤테이션: a109 T2 원장의 wrapper 뮤테이션 중 콘솔에 적용 가능한 판 이식.
