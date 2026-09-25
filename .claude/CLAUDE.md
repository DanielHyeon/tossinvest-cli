# TossOS — agent safety bootstrap

이 파일은 Claude/Codex 최소 안전 부트스트랩의 정본이다.
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
14. **보고는 Chisle 로 짧게 쓰되 증거는 줄이지 않는다.** 인사·반복 요약·장식만 뺀다.
    판정·수치·실패 출력·`not-applicable` 사유·남은 위험은 그대로 남기고, 커밋·PR·OpenSpec
    문서는 대상이 아니다. Chisle 의 도구 출력 압축 hook(Bash 출력의 중간을 잘라 바꿔치기)은
    쓰지 않는다 — 측정은 잘리지 않은 출력으로 한다.

11~14 의 외부 스킬은 아래 안전 부트스트랩과 SDD 순서보다 아래다. 충돌하면 스킬을 멈춘다.

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
   High-risk 기존 함수는 면제할 수 없다.
5. 함수 내부의 분기·early return·side effect를 **근거로 삼는 문서**는 그 근거를 손으로
   읽어서 만들지 않는다. proposal·design·review가 그런 주장을 담으면 대상 함수의
   `tools/logic-map` AST 산출물을 **먼저** 만들고 그 열거를 근거로 쓴다.
   산출물 없이 쓴 분기 주장은 미검증이다.
6. `make sdd-sync`, `make sdd-check`, `make gate CHANGE=<change-id>`와 독립 리뷰가
   끝나기 전에는 완료라고 보고하지 않는다.

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
8. gstack review + make sdd-check + make gate
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

## 하네스 주석 (Claude)

이 절은 SDD_SHARED 블록 **바깥**이다(`.codex/agents.md` 미러 대상이 아님).

- **⚠ `codegraph affected` 는 이 저장소에서 Go 테스트를 못 찾는다. 기본값을 믿지 말 것.**
  `affected` 는 공유 `isTestPath` 대신 CLI 안에 박힌 JS/TS 정규식 6개
  (`.test.` · `.spec.` · `/__tests__/` · `/tests?/` · `/e2e/` · `/spec/`)로 테스트를 고른다.
  Go 관례인 `foo_test.go` 는 밑줄 접미사라 **어느 패턴에도 걸리지 않는다**.
  실측(2026-09-18, v1.6.0, `cmd/tossctl/adoptionsettings.go`):

  ```text
  기본 판별식          → 의존 1,732개를 탐색하고 테스트 1건
                         (그마저 Go 가 아닌 auth-helper/tests/test_cli.py 가 /tests/ 로 걸린 것)
  --filter '*_test.go' → 797건
  ```
  797 대 1 이다. **의존 탐색은 도달하고 판별식만 거른다** — 그래프가 빈 게 아니다.
  → **우회(매번 이렇게)**: `codegraph affected <files> --filter '*_test.go'` 를 쓴다.
  기본 실행만 보고 「영향받은 테스트 없음/거의 없음」이라 결론내지 않는다.
  ※ upstream 수정은 issue #1507 → PR #1803(2026-09-08). 최신 릴리스 v1.6.0(2026-08-26)
  이후라 **아직 어느 릴리스에도 없다.** 버전을 올린 뒤 재측정해 이 항목을 지운다.
- **색인 갱신자는 `PostToolUse(Write|Edit) → codegraph sync` 훅 하나다.**
  `.mcp.json` 의 `serve --mcp --path . --no-watch` 는 의도된 선택이고(이 저장소는 `/mnt/D` 위에 있다),
  watcher 를 켜면 갱신자가 둘이 된다. 2026-09-18 이전에는 이 훅이 없어서
  **갱신자가 아예 없었다** — 설정 밖에서 뜬 watcher 데몬 하나가 그 구멍을 우연히 메우고 있었을 뿐이다.
  그 데몬을 정리하면서 훅을 넣었다. `--no-watch` 를 빼는 방식으로 되돌리지 말 것(갱신자 이중화).
- **CodeGraph 수치에는 반드시 버전을 적는다.** 2026-09-18 이 저장소의 색인은 codegraph
  **0.9.8** 로 만든 것이었는데, 그 버전은 파일 간 import 엣지를 거의 만들지 않았다
  (여기 실측: **8 / 11,329 = 0.1%**). upstream 버그이고(issue #779 / PR #708) 수정은
  **1.0.0**(2026-06-12)부터다. 1.6.0 재색인 후 파일 간 해소 36.7%(go 34.7% · python 88.8% ·
  ts 93.3%), 엣지 123,131 → 143,869. 버전 없는 신뢰도 주장은 도구가 아니라 빌드에 대한 주장이다.
