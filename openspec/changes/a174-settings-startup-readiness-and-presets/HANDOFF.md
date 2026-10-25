# HANDOFF — 콘솔 UX 프로그램 (a174 → a175 → a176)

작성 2026-10-25 (Manager: Fable 5.1). 이 파일은 다음 세션이 a173 **이후** 작업을 이어받기 위한 것이다. a173 종결(gate·archive·push)은 작성 시점의 세션이 Opus 에게 맡겨 진행 중이며 이 핸드오프 범위가 아니다.

## 1. 지금 상태 (사실)
- 프로그램 정의: `openspec/changes/archive/2026-10-25-a173-console-ux-common-cleanup/proposal.md` 「프로그램 개요」(a173 archive 뒤 경로; archive 전이면 `openspec/changes/a173-…/proposal.md`). S1=a173, S2=a174, S3=a175, S4=a176, S5=a177.
- 끝난 것: **a177**(성과 조회 창 표기 단일화) archive·push(`abe803db4`). **a173**(공통 정리) 구현·Manager 검증 완료, main `2036fd41b` 에 ff 병합됨, 종결(FLM 번들 7·격리 gate·archive·push) 진행 중 — 시작 전에 `git log --oneline -5` 와 `ls openspec/changes/archive | grep a173` 으로 끝났는지 확인하라. 끝나지 않았으면 **a174 를 시작하지 말고** a173 종결 상태(`openspec/changes/a173-…/tasks.md` 5.1 영수증, `/tmp/claude-1000/…/scratchpad/a173-close/gate.log`)부터 본다.
- 정의만 된 것: a174·a175·a176 — 각 change 디렉터리에 `proposal.md`·`design.md`·`tasks.draft.md`·`specs/operator-console/spec.md`(ADDED 델타)·`.openspec.yaml`·`base-commit.txt`(= `b27a1ea87`, **착수 시 a173 착지 커밋으로 갱신**). Story YAML `STORY-TOS-a17[4-6]` 등록, tracker 상태 `designed`.
- 동시 착수 상한 4(기억 `tossos-sdd-workflow`). 순서: a174 → a175(a174 와 병행 가능, 파일 안 겹침) → a176(a174·a175 와 병행 가능).

## 2. 사람 결정 (확정)
1. 설정 탭 축은 **가역성·빈도 유지**(spec `operator-console` 851 SHALL). 5묶음 안은 철회.
2. 엔진 시작/정지는 **검증 콘솔 유지**(spec 634·662). 기동 준비 체크리스트(읽기 전용)를 그 위에 둔다(a174).
3. 개발자 식별자는 **접힌 기술 정보**(`details.tech`, a173 define)로. 안전 상태 단어는 접지 않는다.
4. 확인 문구·추가 확인 단계 금지(사용자 지시, 기억 `no-typed-confirmation-friction`). 주문·토글·승인·엔진 버튼 클릭 수 불변.

## 3. 사람 결정 (미확정 — 다음 세션이 먼저 물을 것)
1. **구현 worktree 위치**: 저장소 안 `worktree/<id>` 는 저장소 전수 census 시험을 오염시켜 main 체크아웃 `make test` 가 빨갛고 CodeGraph 색인이 부푼다(a177 실측). 권장: `/mnt/D/Axipient/workspace/TossOS.worktrees/<id>`(형제 디렉터리). 사용자는 앞서 `./worktree` 를 지시했었다 → 확인 뒤 진행.
2. **`docs/WORKFLOW.md` 비례 원칙 문구 vs gate ⑤**: 기존 Go 함수를 한 줄이라도 고치면 `check_analysis.py` 가 FLM 번들을 요구하고 `not-applicable` 표지는 required 0 일 때만 통한다(기억 `flm-exemption-only-when-no-function-modified`). 권장: WORKFLOW 문구를 게이트에 맞춘다(docs 커밋 1건).
3. a177 의 ai-memory durable page(교훈) 작성 여부 — 규칙 18 은 change 종결 시 검증된 교훈을 허용, 스킬은 명시 요청 요구. 사용자에게 한 번 물을 것.

## 4. 작업 방식 (이 프로그램에서 쓴 것, 그대로 반복)
- 역할: Manager = OpenSpec 작성·검토·최종 검증만. 구현(코드·시험·측정) = Sonnet 5.5 Teammate. 종결(FLM 번들·tracker·sdd-sync·격리 gate·archive·push·retain) = Opus.
- 착수 절차: ① `tasks.draft.md` → `tasks.md` 로 이름 변경, `base-commit.txt` 를 현재 main 으로 갱신, `review.md` 에 proposal-freeze 경량 리뷰(validate + Manager 셀프리뷰) 기록 → ② `openspec validate --all --strict --no-interactive` 0 failed 확인 → ③ worktree 생성(브랜치 `feat/<change-id>`) → ④ Sonnet 브리프(아래 5) → ⑤ 구현 보고 뒤 Manager 가 diff 읽기·시험 재실행·크롬 실측(데스크톱 1280·모바일 390 × 라이트·다크, `chrome-devtools` `emulate`) → ⑥ 수정 라운드 → ⑦ Opus 종결.
- 구현 브리프 필수 항목: 작업 디렉터리 한정 · 운영 콘솔 37085·`~/.config/tossctl`·docker 금지 · 격리 콘솔은 빈 설정 폴더 + 새 포트(383xx) + PID 종료 증명 · 버튼 0 · TDD(RED 출력 기록) · **기존 Go 함수 편집 0 목표(새 함수·새 파일·템플릿 함수)** · `check_analysis.py --change <id>` 의 required N 보고 · 고정 리터럴 시험 수정은 tasks.md 에 전수 기록(약화 금지) · 측정 하네스 `openspec/changes/archive/2026-10-25-a173-…/measure/{shoot.cjs,outline.py,README.md}` 재사용 · gstack `review` 경량 → review.md · 커밋 `[<id> <task>]` + `Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>` · main 미접촉·푸시 금지.
- 종결 브리프 필수 항목: a177·a173 선례 그대로(rebase → ff → tasks 종결 체크 → FLM 번들 → tracker → `/tmp/tossos-<id>-close-<date>` 격리 worktree 에서 `make sdd-infra` → `sdd_sync.py --no-gbrain` → `make sdd-check` → `make gate` → `openspec archive <id> --yes` → Story YAML 경로 갱신 → validate → commit → push → worktree/branch 제거 → Hindsight retain). gate ⑥ CodeGraph 15초 probe 는 `codegraph status .` 워밍업 후 재시도, Pending 이 있으면 `codegraph sync .`.

## 5. 각 change 의 핵심 (상세는 design.md)
- **a174**: 기동 준비 체크리스트(새 `readiness.go`, `GatePreflight()` 결과를 옮김, 상시·당일 머리 + 검증 콘솔 엔진 버튼 위, 컨트롤 0) · Guardian 프리셋 통화별 비교표(열마다 기존 POST 폼 1) · 거래 정책 4토글 한 줄 형식 · 편입 설명 중복 정리. 깨질 시험: `settings_cadence_test.go:279`(컨트롤 단일 위치 — 체크리스트는 폼 0 이라 통과해야 함), `settings_limits_test.go` 마커 슬라이스, `settings_operating_test.go:151,217-309`, `engineproc_test.go`. 목표 `/settings/daily` 1280 높이 ≤ 2500px(a173 after 5237).
- **a175**: 최적화 내비 하나(카드 그리드 삭제, KPI 개요 한정) · 중계 3개 요약+링크 · 익절 프리셋 `단계×프리셋` 표(`.tier-card` 제거 시 `TestAmbientMotionContract` 제외 목록 확장 기록) · 후보 필터 `사유×시장` 행렬 · runtime KR/US 두 열 + 계약·레지스트리 기술 정보 · 신호 공통 설명 1회·그림자 밴드 접기. 목표 후보 필터 ≤1200px, 중계 ≤700px, 잘림 0.
- **a176**: 개요 요약 세 묶음(여섯 사실·시장별·여덟 사유·링크 유지, 상세 섹션 제거 금지, 1280 첫 화면 안) · `/verify` 단계 `details` 구조화(`--list` 텍스트 DOM 보존, 폼·CSRF·확인 문자열 바이트 불변).

## 6. 함정 (이번 세션 실측)
- gate ⑤: 기존 함수 편집 1줄 = 번들 1개(시험 파일 fake 도 셈). 사후 작성이면 review.md 에 그렇게 적는다.
- gate ⑥·⑦ 은 main 체크아웃에서 실패한다(저장소 안 worktree 오염) → 격리 worktree 필수.
- `console_prose_test.go:29`: `<details>` 안에 `.notice`/`.danger` 금지. `TestEveryScreenIsCalledOneThing`: 배지는 `<h1>` 밖.
- 모바일 블록 `button, summary, … { display: inline-flex }` 가 summary 의 ▶ 를 지운다 — a173 이 scope/tech summary 를 `list-item` 으로 되돌림. 새 `details` 를 추가하면 같은 규칙에 넣을 것.
- `causes` 템플릿 함수는 값 스냅숏 뷰 모델에만 건다(공유 가변 포인터 페이지 금지, `docs/console-design.md`).
- 스펙 1442 "원장 미판독 안내 1회" · 706 "미측정 + 사유" · 851 "컨트롤 단일 위치, 읽기 전용 표시+링크는 요구됨".
- 구현 중 무거운 `go test ./...` 와 gate 를 동시에 돌리지 말 것(타임아웃 유발).
- `AskUserQuestion` 도구는 이 환경에서 조용히 실패 — 선택지는 번호 매긴 평문으로.

## 7. 다음 한 걸음
1. a173 종결 완료 확인(archive 디렉터리 존재 + `origin/main` 에 푸시).
2. 사용자에게 §3-1·3-2 를 물어 답을 받는다.
3. a174 착수: `tasks.draft.md` → `tasks.md`, `base-commit.txt` 갱신, review.md proposal-freeze, worktree 생성, Sonnet 브리프.
