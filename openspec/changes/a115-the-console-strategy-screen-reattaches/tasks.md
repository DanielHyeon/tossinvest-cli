# Tasks — a115-the-console-strategy-screen-reattaches

**이 change는 a109 review §2 A2 P1-1의 이행으로 등록만 먼저 했다(2026-08-16).**

## 0. 착수 전 조건

- [ ] 0.1 a114와의 합본 여부 판단 — 같은 `cmd/tossctl/console.go` 부팅 경로를 편집한다.
- [ ] 0.2 base 재고정(`tools/sdd/capture_change_base.py --change a115-…`).
- [ ] 0.3 콘솔 전략 projection dial과 접힘 지점, 소비 page의 Function Logic Map을
  proposal 갱신 전에 만든다.

## 1. 구현 (착수 시 상세화)

- [ ] 1.1 RED: 전략 runtime이 구성된 채 엔진만 내려간 상태에서 콘솔 화면이
  NOT_CONFIGURED가 아니라 도달 불가 상태를 표시하는 테스트.
- [ ] 1.2 GREEN: nil 접힘 제거 + 재부착 이식.
- [ ] 1.3 뮤테이션: 접힘 재도입 시 실패하는 핀.
