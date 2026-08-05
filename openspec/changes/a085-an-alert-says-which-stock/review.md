# a085 · Review

## 2026-08-05 · 경량 리뷰 (Manager 셀프)

위험 등급 Normal — 알림 문구와 표시 전용 배선. 주문·손절·사이징·원장 판정 경로를
건드리지 않으므로 WORKFLOW의 경량 리뷰(validate + Manager 셀프리뷰 + 기록) 대상이다.

### 발견 1 (수용, 범위 축소) — 콘솔 절반이 a077의 승인된 요구사항과 충돌한다

구현 후 `TestAQuarantinedPositionIsNotDrawnAsProtected`가 실패했다. a077이 "격리된
행에 보호선을 그리지 않는다"를 안전 요구사항으로 세워 두었고, a085의 콘솔 절반은
그것을 뒤집는다.

`Frozen` 표시와 기존 경고를 함께 두면 안전하다는 것이 a085의 논거지만, 그것은 승인된
requirement의 **수정**이고 WORKFLOW는 별도 리뷰 게이트를 요구한다. ADDED로 새
요구사항을 쓰면 validator만 통과하고 두 요구사항이 서로를 부정한다.

콘솔 절반을 a086으로 분리했다. proposal에 근거를 남겼다. **a083·a084가 이 문제의
대부분을 이미 해결한다** — 격리가 재판정되어 풀리면 exit line은 `fresh`가 되고 다섯
칸이 정상적으로 채워진다.

### 발견 2 (수용) — 산문을 고정한 테스트가 번역에 걸린다

`TestAnExternalIncreaseAfterAdoptionIsReported`가 제목의 영어 단어 `"grew"`로 알림을
찾고 있었다. 알림은 정상 발송되는데 테스트만 실패한다.

이벤트 Key(`|grown|`)로 바꿨다. 제목은 운영자 산문이고, 산문을 고정한 테스트는 그
테스트가 다루는 것을 전혀 바꾸지 않은 변경에 실패한다.

### 발견 3 (설계 확인) — 비교 타입에 표시 필드를 넣는 것

`reconcile.Holding`은 수량 불일치를 판정하는 비교 타입이다. `Name`은 `CostBasisRaw`가
이미 만든 선례를 따른다 — 비교에 참여하지 않고, 비어도 비교가 그대로 동작한다.
`TestNamesDoNotAffectTheComparison`이 그것을 고정한다.

### 발견 4 (설계 확인) — §0.4가 변하지 않는다

이름은 이미 대금을 지불한 `GET /api/v1/holdings` 응답의 기존 필드다.
`TestNamingAStockCostsNoExtraRequest`가 주기당 holdings 읽기가 2회(안정화가 이미 쓰는
수)임을 고정한다.

### 발견 5 (설계 확인) — 기계 판독 표면은 영문·원문이다

`Fields` payload 키와 값, 구조화 로그, 이벤트 타입, 원장 cause는 그대로다. 원장 질의가
언어에 의존하게 되면 사후 조사가 알림 문구의 개정 이력에 묶인다.
`TestTheUnmanagedAlertNamesTheStockInKorean`이 payload 키의 ASCII 여부와 raw 심볼을
단언하고, §0.8로 계좌 참조가 알림에 없음도 함께 단언한다.

## 결정

**SHIP.** 콘솔 절반은 a086으로 분리. `make gate`는 a084와 같은 이유로 병행 세션의
agent-config 편집에 막혀 있다.

## 2026-08-05 · 독립 리뷰 (별도 컨텍스트)

앞의 proposal-freeze 기록은 **Manager 셀프 리뷰**였고, WORKFLOW가 요구하는
"작성자와 검증자의 분리"를 충족하지 못했다. 그 지적을 받고 gstack `/review`를 실행했다.

**보이스 구성**: 별도 컨텍스트 리뷰어 4명 — 적대적 Eng, 보안·안전 불변식, 데이터
마이그레이션, 테스팅/유지보수·성능. 각자 신선한 컨텍스트에서 diff를 읽었고 작성자의
근거를 알지 못한다.
**codex(gpt-5.6)는 여전히 사용량 한도**(2026-08-08 해제)로 교차 모델은 없다.

**결과: 셀프 리뷰가 놓친 blocking 결함을 찾았다.** `issues.md`의 `B` 항목이 정본이며,
여러 건은 실행 가능한 재현으로 확인했고 나는 원본 코드를 직접 읽어 재확인했다.

**결정: 배포 보류.** blocking 항목이 닫히기 전에는 `make gate` 통과 여부와 무관하게
a085 를 배포하지 않는다. 게이트는 테스트가 green이라고 말할 뿐, 여기서 발견된 것들은
테스트가 없어서 green이었다.

### 개정 2 반영 결과

D5 이름을 안정화 전에 배운다 · D6 §0.8 주장을 검사 범위로 좁힌다 · 취약한 문구 단언 교체.

blocking 항목은 전부 RED 테스트 선행 후 닫혔다. 설계는 `design.md` 개정 2,
요구사항은 spec delta 개정 2, 작업은 `tasks.md` 10장에 있다.

리뷰 과정에서 발견한 부수 사항은 `issues.md`에 기록했다.
