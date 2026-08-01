## Context

`exitpolicy`는 ratchet과 ladder를 decimal로 평가하고 `exitloop`가 proposal ratio를 실제 수량으로 내림한다. 현재 1주는 중간 ratio가 0주가 되어 결과적으로 주문이 생기지 않지만 이 동작이 명시적 계약과 독립 테스트로 고정돼 있지 않다.

## Goals / Non-Goals

**Goals:** 정책 판정·설명·주문 수량이 공유하는 immutable exit snapshot을 정의하고 1주 중간 익절 생략을 명시한다.

**Non-Goals:** snapshot DB 저장, UI 자체 구현, 브로커 조건주문, 운영 설정 저장. 정책 descriptor는 transport-neutral `internal/settingmeta` 계약으로 제공하고 a050이 이를 조합한다.

## Decisions

1. 순수 계산 결과 `ExitLineSnapshot`이 entry, initial stop, current protection, next target/protection, high-water, rung, action, ratio와 projected quantity를 가진다.
2. 부분익절 projected quantity가 1 미만이면 주문을 만들지 않고 state-only promotion으로 처리한다. 0수량 intent를 만든 뒤 downstream에서 버리지 않는다.
3. 최종 take-full과 protection breach는 ratio 1로 남은 전량을 사용한다. 따라서 1주는 정확히 1주를 청산한다.
4. current protection과 high-water는 이전값보다 낮아질 수 없다. 동일 관측에서 승격과 breach가 겹치면 더 높은 보호선을 먼저 계산하고 전량 보호를 우선한다.
5. 등록 정책 descriptor는 `BALANCED`, `RUNNER`, `HYBRID 50`의 현재 decimal rung 값을 기본 preset으로 노출하고 각 값의 의미·단위·1주 projection을 포함한다. 미승인 상태의 effective default는 기존 RATCHET 유지이고 추천 preset은 HYBRID 50으로 구분한다.
6. 정책은 stable ID만으로 식별하지 않는다. `policy ID + semantic version + canonical digest`를 immutable identity로 사용하며, 기존 identity 아래에서 rung 의미나 수치를 바꾸지 않는다.
7. snapshot은 policy identity, position generation, observation identity와 canonical input digest에서 결정적으로 만든 `snapshot ID`와 `decision ID`를 가진다. a042 영속성과 a043 order linkage는 이 ID를 사용하고 symbol/time 근사 join을 금지한다.
8. a041은 UI나 transport를 모르는 `internal/settingmeta`의 최소 계약(field key/type, control kind, finite stable option ID, default/effective state, apply timing, safety direction, provenance)을 정의한다. 각 domain change가 값을 소유하고 a050은 category composition만 소유한다.
9. 관측 ID는 account/market/symbol/position generation/관측시각/canonical price의 length-prefixed SHA-256으로 만들고 원문 account를 노출하지 않는다. quote `FetchedAt`이 있으면 그것을 사용하고, 없으면 `ObserveOnce` 시작 시 한 번 캡처한 cycle instant/sequence를 모든 판정에 재사용한다. pre-a042 ID-only 정책 row는 고정된 legacy digest와 정확히 일치할 때만 해석하며, 그 밖의 null/unknown 의미는 fail-closed다.

## Risks / Trade-offs

- [기존 우연한 0수량 동작과 상태 전이가 다름] → 기존 분기를 Function Logic Map으로 고정하고 RED 테스트에서 rung 승격·pending 상태를 비교한다.
- [설명 snapshot과 실행 판단 drift] → 별도 재계산을 금지하고 exitloop가 동일 snapshot만 소비한다.
- [descriptor가 UI 또는 HTTP 구현에 결합] → `internal/settingmeta`는 HTML/JSON 타입을 포함하지 않고 finite option과 provenance만 표현한다.

## Migration Plan

새 순수 타입과 테스트를 추가한 뒤 기존 evaluator/exitloop를 단계적으로 snapshot 소비로 전환한다. 정책·설정·원장 schema는 바꾸지 않는다. 대신 `ExitStateSeed.PolicyIdentity`, `ExitDecisionProvenance`, `PlaceRequest.ExitProvenance`를 a042가 소비할 typed seam으로 남긴다. a042 전까지 journal read의 zero identity는 임의 기본값이 아니라 unknown이며, engine만 고정 legacy identity와 정확히 일치하는 기존 row를 호환한다. rollback은 이전 호출 경로 복원이다.

## Open Questions

소수점 보유 주문 지원은 별도 Story로 남긴다. 이 change는 현 whole-share 주문 계약만 다룬다.
