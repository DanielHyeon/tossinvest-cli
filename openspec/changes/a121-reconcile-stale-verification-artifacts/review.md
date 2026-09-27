# Proposal-freeze review

## Status: BLOCKED before implementation

The independent adversarial review at
[`analysis/proposal-freeze-adversarial-review.md`](analysis/proposal-freeze-adversarial-review.md)
blocked this proposal on 2026-09-06. No RED work, implementation, account read,
or live command is authorized.

The current official OPEN/CLOSED pagination has no proven consistent snapshot
cut; therefore absence of a parent conditional cannot safely exclude a
replacement or triggered child. The contract also lacks a versioned non-raw
identity digest and executable profile/account/market equality checks. These
must be specified, including fail-closed behavior when the official API cannot
prove a consistent read, before an adversarial re-review and subsequent gstack
review can run.

`make sdd-check` remains blocked by stale/missing CodeGraph hard evidence even
after `make sdd-sync`; it is an implementation blocker. PM generation and its
check now pass. a063 remains operationally blocked and unarchived.

## Revision 1 (2026-09-27) — 설계 갭 해소 초안, 재리뷰 전

적대 리뷰의 차단 갭 셋을 design.md 「Revision 1」 이 코드 영수증으로 다룬다. 분기 근거는
`analysis/ast-evidence/`(HEAD `cb378a63`)의 AST 로 인용했다.

| 갭 | 초안의 처리 | 남은 사람 결정 |
|---|---|---|
| G1 부재 증명(컷 토큰 없음 · 후속/child 모호) | 컷 토큰 부재를 코드로 확인(`conditional_reads.go:44-48`). id 부재 대신 "심볼의 OPEN 0건 + CLOSED 에 id 없음 + 두 번 읽기 동일" 을 요구, 그 밖은 전부 거절 | Q1(발동 후 소멸 배제 — 미측정), Q3(경과 한도), Q4(같은 심볼의 다른 조건주문) |
| G2 신원 표현 | 새 줄 종류 `reconcile` + `Artifact.terminal()` 셋째 종결. 키는 기록이 이미 쓰는 `Kind\x00ID`, 근거는 버전·도메인 태그 지문. 구 바이너리는 여전히 outstanding 으로 읽는다(안전 쪽) | Q2(spec 의 원문 id 경계 문구 수정 vs 태그 표현) |
| G3 계좌·프로필·시장 | `AccountRef` 전 줄 일치 + 현재 참조 일치, seq 0 거절, `--record` override 거절, `--market` 필수 + 행 시장 일치 | Q5(계좌 여럿) |

상태는 그대로 **BLOCKED** 다 — 이 초안은 RED·구현·계좌 읽기·라이브 명령을 허가하지 않는다. 적대 재리뷰는
Manager 지시 뒤에 돈다. `make sdd-check` 차단(위 절)도 그대로 구현 차단 조건이다.
