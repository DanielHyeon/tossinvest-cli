# a105 — 보호 감독이 Wired 로 간다 (protection-supervision-goes-wired)

## Why

a100 의 원안 범위에서 proposal-freeze 리뷰(2026-08-11)가 잘라낸 「`Wired` 생산」 부분의
수용처다. 그 목록은 a100 tasks.md 「a105로 이관」 절에 **사라진 것이 아니라 옮겨진 것**으로
보존돼 왔고, 2026-10-04 Manager 판정(a100 R0 (4), review.md 「R0 — 2026-10-04」)이 예약
번호 a105 를 실제 change 로 등록하기로 했다. 등록 시점의 구현은 0 이다 — 이 문서는 범위와
선행 조건을 정본에 고정하는 것이 목적이다.

주의(낡은 참조 정정): 옛 a100 문서의 「a105 = 레인 활성화」 표기는 낡았다. 레인(전략군)
활성화 권위는 a112(서명 4-가족 활성화, 2026-10-05 아카이브)로 갔다. a105 의 대상은
**보호(protection) 감독의 Wired 생산 전환**이다.

## What Changes

a100 이관 목록 전체(원문: a100 tasks.md 「a105로 이관」):

1. `internal/protectionsupervisor` 신규 패키지와 시장별 `Wired` 판정(a100 원안 D2).
2. `productionProtectionAssemblies` 의 `wired` 파라미터화와 identity 문자열 교체(원안 3.3·3.4).
3. `ProductionProvider.initialize` 의 refusal 분화(원안 3.5) — **FLM 선행 대상**.
4. manifest digest 불일치의 진단 가능한 refusal 과 배포 절차의 재서명(원안 D3·6.2·6.3).
5. **서명 도구 신설.** 저장소에 서명자가 없다 — `internal/attest/protection_signature.go` 는
   검증 전용, 서명 함수는 `_test.go` 에만, `supervisor_digest` 를 발행하는 `cmd/` 경로 없음.
   a105 는 이것 없이 `Wired` 를 켤 수 없다.
6. 포지션 단위 coverage latch 와 entry supervisor 소비(원안 D5·4.8).
7. `engine.runInterlock` B3 도달(원안 1.3.1) — `WIRED` 가 생산 가능해야 도달한다.
8. `a071_security_review_test.go` 의 무조건 단언 반전(원안 3.6).
9. `guardian_test.go:131-132`·`interlock_entry_test.go:70-71` 의 `EntryPermitted` 전제 변경(원안 4.9).
10. **flat 포지션 상주 주문 창 닫기.** a100 은 다음 수렴 주기까지의 창을 운영 문서로
    처리한다(a100 6.4.3). 자동 매수가 생기면 그 창은 「방금 산 주식에 남의 손절이 걸리는」
    경로가 되므로 — **a105 가 진입을 열기 전에 닫는 것이 면제 불가 선행 조건이다.**

## Impact

- 착수 조건: **a100 land**(a107 과 같은 결속 — a100 의 lifecycle core·수렴 워커·원장 스키마가
  입력이다). a100 은 현재 M-A 사람 실측 대기.
- High-risk: 보호·Guardian·진입 관문 경로 전부 해당. Full SDD(freeze → FLM → RED/GREEN →
  변이 → 독립 리뷰 → gate) 그대로.
- 안전 불변식: `Wired` ON 은 운영 토글 flip 에 준한다 — 사람 승인 없이 켜지 않는다.
- 이 등록 커밋 자체의 코드 변경: 0.
