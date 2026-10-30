# a169 귀속·보고 순수 계약 조사

작업 범위: 정본 design §5/§9 및 현재 a140 source 손읽기. 코드·시험·운영·설계 승인 없음. 기존 사용자 결정 Q1/Q5/Q8/Q9/Q10은 `proposal.md:95-104`, 구체 규칙은 `design.md:218-271,491-519`이다.

## 최소 귀속 경계
평가기는 종목·세션마다 1회이고 cohort마다 재평가하지 않는다(`design:222-230`). 순수 판정은 명시적 cohort membership, 실제 선정 시각 T, 정규장 open, 장중 최초 BREAKOUT_CLOSED transition의 봉 수 N 및 setup identity, 종료 후 재계산의 돌파 봉 i/setup identity·재계산 성공 여부를 입력으로 받는 것이 최소다. 데이터 획득·봉 prefix 반복 평가·신호/원장 파싱은 다른 경계다. 결과는 귀속, 선정 전 제외, 미정, 불일치 등 서로 다른 상태를 유지한다.

- 돌파가 확인되면 봉 닫힘 `open+(i+1)분`이 T 이하일 때만 선정 전 제외, T보다 늦으면 귀속한다(`design:228-233`). 동시각은 제외다. 돌파가 없다는 것은 평가 성공과 신호/재계산 증거로 확인된 경우이며, 신호 행이 아예 없는 경우 정본은 U2 미정으로 처리한다(`:262-265`). U0/U1까지 같은 미정 규칙을 임의 확장하지 말고 후속 설계에서 입력 결손 처분을 명시한다.
- 장중 transition이 있고 재계산 setup_id가 다르면 불일치다. `N=i+1` 일치, `N<i+1` 불일치, `N>i+1`은 i와 N−1의 닫힘이 각 cohort의 T 양쪽에서 같은 편일 때만 지연 일치, 다르면 불일치다. 장중 transition 없음+재계산 돌파도 불일치(`design:244-258`). 재계산 실패·결손을 '돌파 없음'이나 0으로 조용히 바꾸지 않는다. 유효 시간·N/i·identity 누락은 명시 오류/미정 처분이 필요하며 원장 결합 전에 고정한다.
- U0/U1의 `@T_U2` 대조는 별도 기준 시각으로 위 필터를 다시 평가한 보조 열이며 판정이 아니다(`design:234-236`). 늦은 선정의 실제 시각과 cohort별 mismatch를 독립 적용한다.

## 보고 단위와 정본 판정
`design:493-508`: 같은 lot 내 cohort U0/U1/U2 × STOCK/ETF/ETN 아홉 칸. 기타·미분류는 개수만. session/cohort 멤버십은 중복을 허용해 여러 cohort에 같은 종목을 귀속하고, 겹침과 해당 cohort 고유 종목을 따로 센다. 늦은 cohort-session, 불일치 종목-session, abort session은 판 합계에서 제외하고 별도 계수한다. 판 합계는 정시 전체와 short cohort-session 제외 두 형태다. 미정·제외를 성공 라벨로 합치지 않는다.

각 칸에서 실제 라벨 수 n, outcome별 분포와 순 `r_ppm`을 사용해 평균 순 R=Σr_ppm/n/10^6, 표본 표준편차(n−1), 표준오차를 산출한다(`design:494-498`). n=0/1의 평균·표준편차 undefined 처분을 typed 결과로 명시할 필요가 있다. **Q8 정본은 raw 라벨 개수만으로 결론을 내리지 않는다**: n<30 보류; n≥30일 때 평균 순 R>0 후속 검토 후보, ≤0 중단 근거 유지다(`:499-505`). 통계 판정이 아닌 사람 판단 입력이며 9칸 동시 비교 표기를 유지한다. 비용은 이미 r_ppm에 반영되므로 두 번 빼지 않는다. 서로 다른 lot의 합산이나 U0@T_U2 보조 열의 판정은 금지다. 수치/합·분산 overflow 방지에 math/big를 검토하고 원문 정수와 반올림 정책을 명시한다.

## 기존 재사용과 공백
- `ledger.go:31-47,49-60,87-91`의 `ledgerRow`는 session/lot/symbol, signal SetupID/Transitions/Bars 및 label Outcome/RPPM을 제공한다. 현재 a169 cohort·breakout·type·abort 행은 없어 현재 타입만으로 완성된 a169 입력이라고 주장할 수 없다. 원장 파서·identity 연결 검증은 순수 필터 밖이다.
- `evaluate.go:51-66`의 `contiguousPrefix`·`evaluateSetup`, `label.go:8-18`의 `labelResult`·`labelProposal`은 a140 평가/라벨 생성 선례이며 실제 series·quote·sizing 의존이 있다. 귀속 필터에서 이 평가기를 새로 구현하거나 HTTP를 호출할 필요는 없다. breakout i를 얻는 prefix 재평가 자체는 별도 후속 로트(`design:238-246`).
- `report.go:12-15,17-32,39-100,126-148`의 minVerdictSamples=30, phaseRank, 기존 lot/session label/outcome/rSum은 참고 가능하지만 `renderReport`는 원장 행·a140 형식에 결속되며 a169 cohort/identity/mismatch/표준편차가 없어 그대로 완성 보고서가 아니다. 특히 phaseRank에 terminal INVALIDATED/TIMED_OUT이 없으므로 BREAKOUT_CLOSED 여부를 phase만으로 추정하지 않는다(`design:262-264`).
- 새 함수는 독립 typed 입력의 결정적 변환과 조건별 결과만 담당하고 신뢰받는 ledger/provenance·재시작·fence/abort 연결은 후속으로 남긴다. 부정확한 입력을 실제 무돌파/성공으로 보정하지 않는다.
