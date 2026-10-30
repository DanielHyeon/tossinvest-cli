# 순수 계산 이후 남은 a169 계약

2026-10-31 Manager. U0/U1·유형 게시 `54f1dc59b`, U2 단일 점수 게시 `789ed34fb`. U2 합집합 `9ebb80d62`는 독립 검증과 Manager 게시 승인을 받았고 이 기록과 함께 게시한다. 모두 기존 runtime caller에 연결되지 않은 순수 계산이다. 전체 design freeze1.3, tasks2.2, 리허설·측정·archive는 미완이다.

## 확정한 범위

U0/U1 정상30·승인된 별도20판, 숫자6자리·KRW·가격 경계, 정확 enum과 raw 보존. U2는 각 랭킹별 필터의 ANY-pass 합집합, raw volume 존재 시 V 우선, 두 수신시각 max, 정확20일 분모·두 번 floor·ratio와 raw rank 정렬을 사용한다. float64 안전 정수·중복 날짜·int64 결과 제한은 지원 제한이며 실제 자료 품질 보장이 아니다. 순수 결과의 Amount/Volume filter view는 U2 시점 입력에 대한 중간 계산이고 장초 U0/U1의 실제 선정 원장을 대신하지 않는다.

## 시작·전송 차단 설계에서 남은 결정

현재 Snapshot API만으로 전송 직전의 shared event fence는 구현되지 않았다. source 조사 `prewire-fence-research.md`는 wire lease Acquire 뒤 자기 Begin 전에 lease 소유를 검증해 event 상태를 읽는 최소안을 제시한다. 기존 Snapshot은 pending 상태에서 wire lock을 다시 검사하므로 같은 lease 아래 그대로 재사용하지 않는다. 새로운 내부 읽기와 transport 결속은 기존 함수 변경이므로 별도 FLM/BTM·독립 설계·실제 local TLS token/401 반례 검증을 거쳐야 한다. 일반 callback framework를 만들지 않는다.

**사람 결정 완료(2026-10-31):** 「현재 안정된 시점부터 시작하고 과거 gap은 제외」. immutable 시작 cursor 이전 이력은 제외하되 현재 pending/dirty/손상/읽기 오류는 거절한다. `prewire-fence-plan.md`로 구현 계약을 검토한다. gap 삭제·새 instance 자동 전환은 하지 않는다. 이 결정만으로 actual measurement fence와 실행 모드가 완료되는 것은 아니다.

별도 미완: 감시 실패/취소와 first Write의 선형화·공유 abort latch 수명, 실제 활성 엔진·프로필·동일 예산 identity 결속, 같은 날 abort의 영속성/동시 writer/재시작. 이미 전송을 시작한 요청을 없던 것으로 주장하거나 broker 도착 시각을 보장하지 않는다.

## 원장·보고·운영

U2 양 랭킹 가격이 다를 때 단일 last_price projection은 아직 정하지 않았다. 순수 결과는 두 원문/사유를 보존하며 원장 parsing/schema를 바꾸지 않는다. 선정 ready deadline·cache·일봉 provenance·late 전체제외, 귀속의 무돌파/신호 없음/identity mismatch, lot별 보고와 중복 집계는 후속이다. `attribution-report-research.md`는 연구이며 승인된 구현 명세가 아니다. Q8은 n과 평균 순R을 함께 사용하고, 기억에 수익성 결론을 남기지 않는 규칙과 보고 계산을 혼동하지 않는다.

a166의 실제 namespace·UID·filesystem·모든 참여자 확인 및 운영 전환은 별도 사람 단계다. 별도 키는 토큰 격리이며 3req/s 예산 분리가 아니다. a087 production admission nil과 사람5.4/활성화5.5는 유지한다. 시크릿·실계좌·브로커·실측·서비스/설정 변경은 수행하지 않았다. 전체 tests/gate도 미실행이며 로트별 지정 시험 결과만 인정한다.
