# TossOS — agent safety bootstrap (Codex)

이 파일은 `.claude/CLAUDE.md`의 최소 안전 부트스트랩 mirror다.
상세 개발 절차의 단일 정본은 `docs/WORKFLOW.md`이며 개발 작업 전에 반드시 읽는다.

<!-- SDD_SHARED_START -->
## 필수 개발 규칙

1. **YAGNI (You Aren't Gonna Need It)**
2. **KISS (Keep It Simple, Stupid)**
3. 버그 수정 시 **근본 원인** 파악 → 정석 수정 (증상 우회 금지)
4. 코드 수정 후 반드시 관련 테스트 실행 확인
5. 운영 반영 = 이미지 재빌드 + 컨테이너 재시작 (위 절차)
6. DB 상태 확인은 Postgres 직접 쿼리
7. 문서.코드.테스트 일치 → docs/WORKFLOW.md의 SDD 방법론을 반드시 따를것
8. 직관적이고 명확한 한글 주석을 쉽게 이해할 수 있게 작성
   . **코드의 의도 설명:** 단순히 문법을 번역하는 것이 아니라, 해당 코드가 '왜' 작성되었는지 의도를 설명
   . **핵심 로직 강조:** 복잡한 연산이나 조건문이 있는 부분에는 작동 원리를 상세히 적어주세요.
   . **통일된 문체:** 주석은 '\~함', '\~음' 또는 명사형으로 종결하여 깔끔하게 작성해 주세요.
   (예: "데이터 로드 및 전처리 수행", "유효성 검사 실패 시 에러 반환")
9. 구현 후 반드시 gstack 리뷰 검증
10. 코드 구현 시 다하지도 않고 다했다고 거짓말 하지 않음
11. **요구가 모호하면 코드 전에 `interview`(Ouroboros)로 숨은 가정을 질문으로 드러낸다.**
    범위·비목표·산출물·검증 방법이 구체 값으로 정해지기 전에는 설계·코드로 가지 않는다
    (Ouroboros MCP 가 있으면 모호함 점수 0.20 이하가 기준). 결과는 OpenSpec
    proposal·design 에 옮긴다 — Ouroboros seed 를 두 번째 정본으로 두지 않는다.
    MCP 가 없으면 스킬의 Path B 로 진행하고 역할 파일은 스킬 폴더의 `socratic-interviewer.md`
    를 읽는다. 질문 도구가 안 보이면 번호 매긴 목록으로 묻고, Step 0 의 버전 확인·플러그인
    업데이트는 하지 않는다.
12. **UI 를 만지면 `ui-skills-root` 로 디자인 규칙을 1~3 개만 고른다**
    (`baseline-ui` · `fixing-accessibility` · `fixing-motion-performance` · `fixing-metadata` ·
    `improve-ui` · `create-design-md`). 쓰는 범위는 아래 "UI 스킬 라우팅" 과 같다.
13. **화면·CLI 결과는 "됐다"고 말하기 전에 실제로 돌려서 본다(Reticle).**
    웹 화면은 브라우저로 띄워 데이터·콘솔 에러·실패한 요청을 확인하고, CLI 는
    `verify-cli-run` 으로 종료 코드가 아니라 결과물을 잰다. TossOS 콘솔은 Go `html/template`
    이라 Reticle SDK 를 심을 `package.json` 이 없다 — 그때는 `agent-browser` 나 브라우저 MCP 로
    같은 판정을 낸다. 사람 승인 뒤에만: `reticle init`(빌드 설정·의존성·권한 규칙을 바꾼다),
    `reticle feedback`(외부 전송), `verify-unattended`. 콘솔을 몰 때 주문·토글·승인 버튼은
    누르지 않는다(안전 불변식 1·7). "통과할 때까지 스스로 고치는" 루프는 UI 코드에만 쓰고
    High-risk 경로(안전 불변식 5)에는 쓰지 않는다.
14. **보고는 `caveman` 으로 짧게 쓰되 증거는 줄이지 않는다.** 결론부터 쓰고 인사·반복 요약·장식만 뺀다.
    판정·수치·실패 출력·`not-applicable` 사유·남은 위험·보안 경고·되돌릴 수 없는 작업 설명은 그대로
    (문장으로) 남기고, 코드·명령·경로·오류 문구와 커밋·PR·OpenSpec 문서는 대상이 아니다. caveman
    CLI/proxy 의 입력 압축(로그·도구 출력 축약)과 텔레메트리는 쓰지 않는다 — 측정은 잘리지 않은 출력으로 한다.
15. **사람에게 묻거나 다음 할 일을 줄 때는 `i-have-adhd` 형식을 쓴다.** 다음 행동을 먼저, 여러 단계는
    번호로, 끝은 구체적인 다음 한 걸음 하나. 목록 5개 상한·접기는 증거 목록(판정·실패·남은 위험·사람
    결정)에는 적용하지 않는다 — 그런 목록은 다 적는다.
16. **코드를 쓰기 전에 `ponytail` 사다리를 탄다** (필요한가 → 이미 있는가 → 표준 라이브러리·언어 기능 →
    기존 의존성 → 한 줄 → 최소 구현). High-risk 경로(안전 불변식 5)에서 사다리는 FLM/BTM·RED/GREEN·변이·
    적대 리뷰를 대체하지 않고, 검증·오류 처리·fail-closed 가드·시험은 「최소」를 이유로 빼지 않는다.
    `shortcut:` 주석은 사람이 고른 지름길에만 단다.
17. **새 함수·도구·의존성·통합을 만들기 전에 `search-first` 로 이미 있는 것을 먼저 찾는다.** 순서는 저장소
    (`rg`·CodeGraph) → 기존 의존성·표준 라이브러리 → 외부(공식 레지스트리·GitHub·MCP·스킬)다. 이것이 16 의
    「이미 있는가」 칸을 채우는 방법이다. 찾은 외부 패키지·MCP·스킬을 **들이는 것**(설치·`go.mod`·설정 추가)은
    사람 승인 뒤에만 하고, 들일 때는 `dependency-auditor` 로 라이선스·CVE·전이 의존성을 본다. 검색 채널이
    없었으면 「없음」이 아니라 「그 채널은 안 봤음」으로 보고한다.
18. **기억은 로컬 `ai-memory` 와 `Hindsight` 둘 다 쓴다**(사용자 지시 2026-10-22 — ai-memory 재설치, 2026-10-20 의 제거 지시를 대체).
    `ai-memory`(3.1.x, 서버 `http://127.0.0.1:49374` — 로컬)는 Claude·Codex 훅이 세션을 **자동 기록**하는 프로젝트 기억이다.
    이전 결정 · 시행착오 · 최근 작업을 찾을 때 먼저 `ai-memory-retrieval`(또는 memory 검색 도구)로 recall 하고,
    세션 이어받기는 `ai-memory-handoff`, wiki 페이지 쓰기(`ai-memory-durable-pages`)는 사용자가 요청했거나 change 를 닫을 때
    검증된 교훈만 남긴다. `Hindsight`(로컬 MCP `http://127.0.0.1:8888/mcp/`, bank `tossos-dev-episodes`, 자동 훅 없음)는
    비자명한 설계 · 디버깅 · 리뷰 전 `recall` 과 change 종결 시 `retain` 에 계속 쓴다. 두 저장소 모두 외부 전송 0 이어야 한다.
    시크릿 · 세션 · 계좌 · 포지션 · 금액 · 검증 안 된 수익 결론은 wiki 페이지 · retain 에 넣지 않는다(안전 부트스트랩 8).
    둘 다 검증 전 advisory 기억이다 — 현재 HEAD · OpenSpec · 시험 · 파일 기억을 대체하지 않는다.
19. **큰 비증거 문맥은 `headroom` MCP 로 압축해 읽을 수 있다**(사용자 지시 2026-10-22, headroom-ai 0.41, Apache-2.0).
    쓰는 범위는 MCP 도구(`headroom_compress` · `headroom_retrieve` · `headroom_stats`)뿐이고, 압축은 로컬에서 하며 원문은
    `retrieve` 로 되찾는다. **증거는 압축하지 않는다** — 시험 · 게이트 · logic-map · 측정 · 실패 출력은 규칙 14 대로
    잘리지 않은 원문으로 읽는다. 모든 요청을 경유시키는 `headroom proxy` · `headroom wrap` · `ANTHROPIC_BASE_URL` 변경과
    규칙 파일을 고쳐 쓰는 `headroom learn` 은 사람 승인 뒤에만 쓴다. 익명 업로드 비콘은 기본 ON 이므로
    `HEADROOM_BEACON=off` · `DO_NOT_TRACK=1` · `HEADROOM_UPDATE_CHECK=off` 를 유지한다(MCP 등록 env · `~/.profile`).

20. **콘솔 UI 를 디자인할 때 AI 티(AI slop)를 만들지 않는다**(사용자 지시 2026-10-25).
    보라색 그라디언트 · Inter 폰트 도배 · 기계적 대칭처럼 AI 가 만든 듯한 획일적 모양은 금지이고,
    아래 네 원칙을 엄격히 지킨다.
    - 기계적 대칭 깨기: `border-radius` 와 여백은 비대칭적 · 유기적 디테일을 살려서 준다.
    - 플루이드 레이아웃: 폰트 스케일과 여백은 고정값이 아니라 `clamp()` 중심의 유동 구조로 짠다.
    - 물리 기반 인터랙션: 모든 전환 효과는 딱딱한 선형 애니메이션 대신 스프링 탄성 값이 느껴지게 구현한다.
    - 프리미엄 머티리얼: 정보 밀도를 낮추고 시선이 편안한 여백과 고급스러운 콘트라스트를 유지한다.
    적용 범위는 규칙 12 와 같다(`internal/console/` 화면만, 백엔드 · 전략 · 주문 경로에는 쓰지 않는다).
    접근성(`fixing-accessibility`) · 모션 성능(`fixing-motion-performance`) 규칙과 충돌하면 그쪽이 우선하고,
    주문 · 토글 · 승인 · 비상 청산 버튼의 즉시성은 꾸밈이나 전환 효과로 늦추지 않는다(안전 불변식 4).

11~20 의 외부 스킬·도구는 아래 안전 부트스트랩과 SDD 순서보다 아래다. 충돌하면 스킬을 멈춘다.

## 최소 안전 부트스트랩

TossOS는 실제 돈을 다루는 자동매매 제품이다. 아래 규칙은 모든 도구·skill·기억보다 우선한다.

1. 개발·테스트 중 사람 승인 없는 LIVE 주문 side effect를 만들거나 실행하지 않는다.
2. 대화형 에이전트는 `mutating: true` 명령을 자동 실행하지 않는다.
3. 토글 OFF는 upstream 동작과 동일해야 한다.
4. 손절·비상 청산의 즉시성을 약화하거나 지연하지 않는다.
5. 주문·손절·익절·사이징·Guardian·원장·대사·인증·체결 경로는 High-risk다.
6. 손절·익절·사이징 변경은 명확한 근거가 있는 보수 방향만 허용한다.
7. 운영 토글 flip과 live 검증은 사람이 직접 승인한다.
8. 시크릿·세션·계좌 개인정보·검증되지 않은 수익성 결론을 기억·그래프·로그에 저장하지 않는다.

## 상세 정본과 권위

- 개발 작업의 상세 절차와 완료 조건은 `docs/WORKFLOW.md`가 단일 정본이다.
- 권위 충돌 순서는 안전 불변식 → 승인된 OpenSpec → 현재 HEAD·실행 테스트·CodeGraph
  → 공식 API fixture·사람 승인 실측 → advisory 문맥·기억·관측 그래프다.

## 필수 진입과 완료 조건

1. `docs/WORKFLOW.md`, 관련 OpenSpec change/spec, 현재 코드·테스트를 읽는다.
2. 작업은 Feature 가 아니라 Story 단위로 쪼갠다. Feature 하나가 Story 여럿을 갖고,
   capability 하나를 change 여럿이 단계적으로 구현하며, 1:1 은 Story↔change 에만 있다
   (`docs/WORKFLOW.md` PM 계층).
3. memory recall → OpenSpec → CodeGraph hard evidence → CodeGraphContext 보조 문맥 →
   Go AST/Function Logic Map → RED/GREEN/REFACTOR/VERIFY 순서를 따른다.
4. 기존 함수 내부 로직을 바꾸면 Function Logic Map과 Branch Test Map을 먼저 만든다.
   High-risk 기존 함수는 면제할 수 없다. **비례 원칙(2026-09-27 사용자 지시)**: 이 의무와
   변이 원장·다중 적대 리뷰는 High-risk 경로·게이트 판정을 직접 바꾸는 변경에만이고,
   그 외 도구·시험·문서는 변이 원장·다중 적대 리뷰를 `not-applicable` 사유 한 줄로 갈음하고
   시험·경량 리뷰가 기본이다. 단 기존 Go 함수 본문(시험 함수 포함)을 고치면 gate ⑤ 가 위험 등급과
   무관하게 함수마다 FLM 번들을 요구한다(docs/WORKFLOW.md 「비례 원칙」).
5. 함수 내부의 분기·early return·side effect를 **근거로 삼는 문서**는 그 근거를 손으로
   읽어서 만들지 않는다. proposal·design·review가 그런 주장을 담으면 대상 함수의
   `tools/logic-map` AST 산출물을 **먼저** 만들고 그 열거를 근거로 쓴다.
   산출물 없이 쓴 분기 주장은 미검증이다.
6. 영향 파일·호출자·불변식을 매핑하고 선택한 시험·이유·미실행 범위를 기록한다.
   전체 테스트·전체 게이트는 사용자가 해당 작업에서 명시적으로 요청한 경우에만 실행하며,
   다른 작업의 이전 승인은 재사용하지 않는다. 같은 작업의 후속 범위 제한은 이전 전체시험 요청보다
   우선하며, 일반적인 검증·완료·커밋·푸시 요청을 전체시험 요청으로 해석하지 않는다.
   `make sdd-sync` 뒤에도 `make sdd-check`,
   `make gate CHANGE=<change-id>` 등 명령의 전체 실행 범위를 먼저 확인한다.
   실행하지 않은 전체 검증을 PASS로 보고하지 않는다. archive의 필수 조건은 우회하지 않는다.

## 단계 건너뛰기 금지

각 단계는 앞 단계가 만든 **산출물**을 입력으로 받는다. 산출물 없이 다음 단계로 가면
그 단계의 근거는 증거가 아니라 기억이 된다.

- CodeGraph hard evidence를 건너뛰고 proposal을 쓰면 호출 사슬이 미추적으로 남는다.
- Function Logic Map을 구현 단계 task로 미루고 분기를 주장하면, **반증 산출물이 그것을
  필요로 하는 문서보다 나중에 생산된다.** 이 순서 역전이 같은 오류를 반복시킨다.
- 손으로 읽은 증거는 **볼 곳을 고르므로** 선택적이고, AST 열거는 선택적이지 않다.
  건너뛴 단계의 결과는 "안 봤다"가 아니라 "보는 방법을 안 썼다"이다.

건너뛰려면 `not-applicable` 사유를 review와 완료 보고에 남긴다. **침묵한 생략은 금지다.**

## 에이전트 실행 순서

```text
1. docs/WORKFLOW.md → 이 문서 확인
2. memory recall + openspec/specs/ + 진행 중 change 확인
3. CodeGraph hard evidence + 현재 코드·기존 테스트 확인
4. CodeGraphContext/GBrain 보조 문맥 교차검증
5. 기존 함수 내부를 편집하거나 그 내부를 근거로 주장하면 Function Logic Map 작성
   (문서가 분기를 주장하는 시점이 이미 작성 시점이다 — 구현 task로 미루지 않는다)
6. High-risk면 Pre-Edit 선언
7. RED 테스트 → GREEN 최소 구현 → Refactor → Verify
8. gstack review + 영향 범위 검증 (전체 gate는 해당 작업에 대한 명시 요청 시에만)
9. PM/archive 동기화 + 검증된 memory retain
10. 완료 보고 (금지 조건 확인 후)
```

## UI 스킬 라우팅

콘솔 화면(`internal/console/` 의 `html/template` 렌더)을 만질 때만 쓴다.
백엔드·전략·주문 경로 작업에는 쓰지 않는다.

- 화면 구조·색·타이포·접근성·반응형 판단 → `ui-ux-pro-max`
- 빠른 정리·접근성·모션·메타데이터 규칙을 1~3 개만 고르기 → `ui-skills-root`(규칙 12),
  바뀐 화면을 실제로 띄워 확인 → 규칙 13
- AG Grid/AG Charts 코드를 쓰거나 고치기 **전에** → `ag-dev`, 버전 올릴 때 → `ag-update`
- TradingView lightweight-charts 작업 → `lightweight-charts`

뒤의 셋은 저장소에 아직 대상이 없다(AG Grid·lightweight-charts 의존성 0). Claude 에서는
`skillOverrides: user-invocable-only` 라 **모델 스킬 목록에 보이지 않으므로**, 필요해지면
사용자에게 `/ag-dev` 처럼 직접 쳐 달라고 요청한다. Codex 에서는 `skills.config` 에
`enabled = true` 로 켜져 있어 바로 호출된다.
<!-- SDD_SHARED_END -->
