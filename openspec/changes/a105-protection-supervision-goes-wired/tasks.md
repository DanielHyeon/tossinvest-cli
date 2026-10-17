# a105 Tasks

착수 조건: a100 land. 그 전에는 0.x 를 포함해 어떤 구현 로트도 시작하지 않는다.
등록 시점(2026-10-05) 구현 0 — 아래는 범위 고정용 골격이고, freeze 가 세분한다.

## 0. 선행

- [ ] 0.1 a100 land 확인(원장 스키마·lifecycle core·수렴 워커가 정본에 있다).
- [ ] 0.2 base 고정 + freeze(proposal·design·spec 델타) — 독립 적대 리뷰 포함.
- [ ] 0.3 **flat 포지션 상주 주문 창 닫기 설계** — 진입 개방보다 먼저 착지해야 하는
      면제 불가 선행(proposal What 10). 창의 존재 증명과 닫은 뒤의 거부 모양을 함께.

## 1. Wired 판정과 감독

- [ ] 1.1 `internal/protectionsupervisor` 신규 — 시장별 `Wired` 판정(원안 D2).
- [ ] 1.2 `productionProtectionAssemblies` `wired` 파라미터화·identity 교체(원안 3.3·3.4).
- [ ] 1.3 `ProductionProvider.initialize` refusal 분화(원안 3.5) — FLM 선행.

## 2. 서명과 배포

- [ ] 2.1 `supervisor_digest` 서명 도구(cmd 경로) 신설 — 검증 전용 현황을 서명 가능으로.
- [ ] 2.2 manifest digest 불일치의 진단 refusal + 배포 재서명 절차(원안 D3·6.2·6.3).

## 3. 진입 결속

- [ ] 3.1 포지션 단위 coverage latch + entry supervisor 소비(원안 D5·4.8).
- [ ] 3.2 `engine.runInterlock` B3 도달 시험(원안 1.3.1).
- [ ] 3.3 `a071_security_review_test.go` 무조건 단언 반전(원안 3.6)·`EntryPermitted` 전제
      변경(원안 4.9 — guardian_test·interlock_entry_test).

## 4. 종결

- [ ] 4.1 변이·race·독립 리뷰·gate — docs/WORKFLOW.md 그대로.
- [ ] 4.2 `Wired` ON 은 사람 승인 항목으로 등재(운영 토글 flip 준용).
