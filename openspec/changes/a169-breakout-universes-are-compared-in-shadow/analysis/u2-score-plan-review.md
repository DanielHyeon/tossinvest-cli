# a169 U2 점수 로트 독립 설계 검토

대상 `analysis/u2-score-plan.md` 2026-10-31, base 2d67c226ed175a4bd53500f8bc668f46f5909f9b. 결론: **bounded READY**, 한 후보의 미연결 순수 score 범위에서 P0/P1 발견 없음. source/문서 손읽기만; 코드·시험·운영 없음. Manager 구현 승인 및 구현 독립 검증은 별도다.

- `design.md:141-145,158-169`의 최신 20봉 S20·결측/0 한계·ratio floor와 increase의 두 번째 floor, 자정 경계·0Time·진단을 계획 :9–17과 대조했다. 산식은 서로 합쳐 한 번만 floor하지 않는다. open/close/selected 검증과 정규장 닫힘 시각 허용은 수학적 지원 경계이며 실제 U2 deadline 충족 판단은 caller 후속이다. selected_at max(receipt), D1 target과 ready deadline을 계산 helper의 입력 검증과 분리했다.
- 2^53−1, 전체 prior 날짜 중복 거절, int64 변환 overflow, 최대25봉은 기존 입력 전체에 대한 정본 보장으로 제시하지 않고 보수적 fail-closed 지원 제한으로 기술됐다. 특히 중복 검사는 선택20 밖에도 적용돼 가용성을 낮출 수 있으며 plan:13–15에 명시돼 있다. float64 내 값도 이미 변환된 원문 정확성을 증명하지 않는다. history_lt_20·baseline_invalid·baseline_zero 및 유효 V=0 구분을 계획한다.
- 입력 V의 canonical 출처·두 랭킹 간 필터 충돌·tie/TopN·선정 receipt 결속·네트워크·마감/late·원장/보고는 범위 밖이라고 명시했다. 오류 사유를 성공 점수 0과 구분하고 입력 봉을 복사하는 검증도 지정돼 있다. a169 전체 2.2·freeze·실측·archive 완료 주장은 없다.
- 권고: 구현 검토에서 중복 및 기간 오류와 다른 배제 사유가 동시에 있을 때 오류 우선순위를 명시·fixture로 고정하고, int64 이외의 중간 계산은 반드시 big.Int로 유지할 것. 이는 한정된 함수의 결정성 검토이며 새 정책 결정을 뜻하지 않는다.
