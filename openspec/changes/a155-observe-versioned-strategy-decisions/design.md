## Context

우선순위 8/9: 버전 전략 관측 전용 연결. 선행 change: a149, a150, a153, a154.
이 문서는 목표 동작을 명세한다. 소스 구현·배포·실거래 검증 완료를 뜻하지 않는다.
공통 기준 증거: `../archive/2026-10-18-a148-clarify-parker-risk-reward-contract/analysis/code-context/`; 구현 전에는 현 HEAD의 호출 사슬·설정 binding과 수정 함수 AST를 새로 확인한다.

## Goals / Non-Goals

**Goals:** 오프라인 계산이 런타임 입력에서도 재현되는지 관측할 필요가 있다. 관측 결과를 주문·예약·운영 제어에서 분리한다.

**Non-Goals:** 기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장과 위험 통제는 보존한다. 이 문서는 운영 설정 변경·실주문·재시작·추가 세션 수집 승인이 아니다. Manager(Astra)가 설계를 작성·검토하고, Terra Teammate가 코드·테스트를 담당한다. 구현 writer와 독립 reviewer는 분리한다. gstack 리뷰, 해당 change 전체 gate, Manager 최종 확인 후에만 Terra가 범위 한정 커밋·푸시·정식 archive/PM sync를 수행한다.

## Decisions

관측은 order-capable strategy entry path의 paper/shadow mode가 아니다. 주문 권한 없는 별도 consumer로 typed diagnostic records만 만든다. broker client, reservation writer, position mutation을 dependency로 받지 않으며 runtime coordinator에 EntryDecision을 제출하지 않는다. 과거 정본의 LIVE-only 실행 경로를 변경하지 않는다.
연구 record는 parent evidence ID, policy/calculation version, source/as_of, ready/refusal, entry/stop/target geometry, planned risk, optional outcome linkage를 가진다. append-only output을 사용한다. 기본 queue=256 records, encoded record max=64KiB, queue-full/oversize는 drop-newest와 이유별 counter로 보고한다. 생산자 enqueue는 nonblocking이며 observation disk I/O·retry를 trading/exit thread에서 수행하지 않는다. record ID는 입력 digest와 policy version으로 결정한다. 단일 writer가 full newline record append와 fsync 완료 후 checkpoint를 atomic replace한다. crash 후 마지막 불완전 record는 신규 파일/오류 상태로 격리하고 기존 evidence를 truncate하지 않는다. checkpoint보다 앞서 durable append된 동일 ID는 재생 시 duplicate skip하며, checkpoint가 durable record를 앞서면 복구를 중단해 명시적 오류를 반환한다. 기록 위치는 기존 증거 경로와 분리한다. 기록 실패/부하는 기존 protective exit를 지연시키지 않는다.
이번 범위에서 구현·검증은 fixtures와 offline replay까지다. 실제 관측 시작은 관측 market/session/window, 수집 API budget, 저장 위치·보존 기간을 적은 별도 사용자 승인 및 frozen manifest가 필요하다. 현재 승인으로 collector를 시작하거나 새 scheduler/컨테이너를 재시작하지 않는다. a128/r1/a147 경로에 쓰지 않는다.

## Risks / Trade-offs

- 과적합·표본 부족 → 수익성 주장과 기능 정확성 판정을 분리하고 unknown/inconclusive를 보존한다.
- 기존 계약 회귀 → frozen fixtures, OFF 동작, 기존 포지션 snapshot을 보존한다.
- 구현 표면 확대 → Separate observation adapter consuming immutable market inputs; no gateway/reservation writer capability. 범위부터 조사하고 기존 기능을 재사용한다. 별도 strategy framework를 만들지 않는다.
- 위험도: High isolation boundary; observation must have no order authority. high-risk 변경은 FLM/BTM·Pre-Edit Gate와 적대적 리뷰가 선행된다.

## Migration Plan

설계→선행 change 검증→Terra 구현/시험→독립·gstack 리뷰→Manager 검증→전체 gate→범위 한정 커밋/푸시/정식 archive. DB 변경이 없으면 migration은 not-applicable로 기록한다. 변경이 필요하면 기존 reader 호환성과 additive-nullable rollback을 먼저 리뷰한다. 운영 반영은 별도 승인 전 실행하지 않는다.

## Open Questions

선행 산출물의 실제 가용성, 기존 구현 재사용 범위와 검증 자료의 충분성은 구현 착수 시 확인한다. 자료가 없으면 승인·측정을 추측하지 않고 해당 기능을 unsupported/not_ready로 남긴다. 후속 정책 성능이나 live 승인을 이 설계로 대체하지 않는다.
