# 목소리 B — 증거 · 하네스 · census

공통 브리프를 먼저 읽는다. 너의 축은 **주장된 증거가 주장을 증명하는가**다.

필수 판정:
1. **RED 진정성.** `red-6.2-seal.log` 의 실패가 기능 부재인가(조립 실패 아님). 편집 전에도 PASS 한 시험들(identity 대조 셋 · 개수 관문 앞 선택)의 판별력이 변이로 보장되는가 —
   각 시험을 **유일하게** 빨갛게 하는 변이가 원장(`mutation-6.2-seal.tsv`)에 있는가, 없으면 네가 만들어 재라.
2. **변이 판별력.** S01~S09 · F01~F05 가 주장한 성질을 겨누는가(예: S02 의 재정의 판이 「identity 로 고르기」를 정말 모사하는가, S06 이 편집 전 모양인가). 원장 밖에서
   살아남는 변이를 찾아라 — 특히 `authorityForOwnerScope` 의 정규화 · 유일성 · 오류 처리, `collectStrategyFirstLegAuthority` 의 분기 순서(B2~B5 재배열), A-lite 식.
3. **`verify_named_tests.py` 자체의 우회.** 하위 시험 이름 처리, `-run` 정규식(이름 접두 충돌), skip · 캐시(`-count=1`), 패키지 빌드 실패, 출력이 `rtk` 로 요약되는 경우,
   `go test -json` 이 pass 사건을 내지만 시험 몸이 아무것도 안 한 경우(t.Skip 없이 조기 return) — 하네스가 증명하는 것과 못 하는 것을 정확히 적어라. 양성 대조(init 조기 종료)를
   사본에서 재현해 보라.
4. **strategyflow census 완전성.** `seal_census_test.go` ①②③ 을 우회해 기본 빌드에서 봉인된 Result 를 만드는 모양(예: 포인터 · 필드 경유 쓰기 `(&r).proposalSeal = …`, 메서드 수신자 경유,
   `unsafe`/`reflect` 가 import 가능한지, 태그 파일이 기본 빌드에 섞이는 조건, 봉인 해시 입력을 바꾸는 다른 함수 편집(동결 셋 밖의 `proposalResultSeal` 입력 필드 정의)).
   review 결속(③)이 문자열 포함만 보는 것의 한계. golden 재생성 환경변수의 위험.
