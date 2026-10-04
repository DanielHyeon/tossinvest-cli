# 목소리 B — 증거의 실효성(시험 · 변이 · RED · FLM)

공통 브리프(`prompt-common.md`)를 먼저 읽는다. 너의 축은 **주장된 증거가 실제로 그 주장을 증명하는가**다.

필수 판정:

1. **RED 를 거치지 않은 시험들의 변이 대체 실효성.** review 5.2.2.1 절은 GREEN 뒤에 더한 시험(반복 헬퍼 둘 · 구조 못 `TestTheProductionCycleDeliversEveryOwnerScopeHandoff` ·
   표기 정규화 두 경우 · 활성화됐지만 닫힌 시장 · 오늘-동등성 핀의 대조)이 RED 를 거치지 않았고 변이 F01~F15 가 판별력을 대신 보인다고 한다.
   (a) 각 시험마다 그것을 **유일하게** 잡는 변이가 원장(`analysis/measurements/lot-5.6.2-5.2.2/mutation-5.2.2.1.tsv`)에 있는가 — 없으면 그 시험의 판별력은 미증명이다.
   (b) 사본에서 그 시험들을 **편집 전 소스(`36ade9b2` 의 해당 파일)**에 대고 돌려 실제로 빨간지 재라(= 사후 RED). 초록이면 그 시험은 편집을 증명하지 않는다.
   (c) 원장에 없는 변이를 **네가** 만들어라 — 특히 구조 못 우회(헬퍼를 한 다리 건너 부르기, `dispatchHandoffs` 를 변수에 담아 넘기기, 시장 단위 handoff 를
   다른 이름의 함수로 감싸기), `ownerScopeOf` 의 market 축 제거, `deliverEachStrategyHandoff` 가 거절된 handoff 에서 멈추기, 반복을 역순으로.
   살아남는 변이는 P1(뚫린 핀).
2. **변이 하네스의 유효성.** `a112_lot_mutate.py`: 사본이 HEAD archive + 로트 파일인가, 무변이 대조군이 GREEN 인가, 변이가 실제로 **적용**됐는가(NOT-APPLIED 없음),
   원복이 되는가, 시험 명령의 `-run` 필터가 변이를 잡을 시험을 전부 포함하는가(필터 밖 시험이 잡았어야 하는 경우). 5.6.2.1 세트(E01~E11)도 같은 기준.
3. **RED 로그의 진정성.** `red-5.2.2.1.log` · `red-5.6.2.1.log` 의 실패 이유가 「기능 부재」 이고 조립 실패 · 컴파일 오류가 아닌가.
4. **FLM/BTM.** `runProductionStrategyMarketCycle` 번들의 좌표 · 호출 표 · return 목록이 `00e1b9bd` 의 소스와 일치하는가(`go run ./tools/logic-map --file … --func Context.runProductionStrategyMarketCycle` 로 대조).
   같은 파일의 이동 번들 열하나(`shift_same_file_bundles.py`)가 옛 좌표만 옮겼고 다른 함수 · 측정 블록 인용을 잘못 옮기지 않았는가(표본 셋 이상 대조).
   이 번들의 「진입 0」 서술이 여전히 참인가.
