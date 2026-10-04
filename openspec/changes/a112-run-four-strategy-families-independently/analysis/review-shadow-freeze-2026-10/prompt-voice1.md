# 보이스 ① — 적대 MUST NOT / 노출 (a112 7.3.1 SHADOW freeze)

공통 브리프: 같은 디렉터리의 `brief.md` 를 **먼저 전부** 읽고 그 안전 규칙을 지켜라. 네 배역은 아래 하나다.

**배역: SHADOW 가 ON · dispatch · 노출 · 재시작 생존에 닿는 길을 찾는다(브리프 질문 1 · 4 · 7 중심).**
- 설계 §3 의 둘째 문 · 불투명 타입 · 봉투 없는 결과로도 남는 길: 레인 런타임(`strategy_lane_runtime.go` evaluate — 관측을 무엇으로 채우는가), 투영(`strategy_lane_projection.go` —
  desired/effective 를 어디서 읽는가), 조정자(`coordinateMarketProposals` · 관문 admit), handoff · dispatch · lease · 게이트웨이, 원장 쓰기(잠금 · 복구 세대), 활성화 파일.
- 핀이 막지 못하는 모양: 같은 패키지 비공개 경로, 함수 값 · 인터페이스 · 제네릭으로 봉투를 운반, shadow 적재기가 FamilyActivation 을 만들 수 있는 길, 관측 메모리를 통해 다음
  파도의 활성화 판정에 새는 길, 재시작 뒤 첫 파도 전 투영.
- 가능하면 c7219640 사본에서 설계의 의도된 모양을 최소 스케치로 만들어(사본 안 실험 파일만) 길이 열리는지 실측하라.
