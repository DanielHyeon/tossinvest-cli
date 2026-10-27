# a176 review

## proposal-freeze (2026-10-27, 경량: validate + Manager 셀프리뷰, docs/WORKFLOW.md 「위험 등급 가중」 UI change)
- 보이스: Manager(Opus 5.5) 단독. `/verify` 는 승인 창이 있는 화면이지만 폼·확인 문자열·CSRF·승인 프롬프트는 바이트 불변이고 단계 목록 텍스트만 무손실로 나눠 접는다 — 적대 보이스는 해당 없음.
- base: `0e8ba1c7c`(a175 archive, origin/main).
- 입력: Fable 5.1 정의, HEAD 원문(`pages.go:132-188`, `templates.go:838-997`, `templates_overview.go`, `internal/verifylive/report.go:416-466`, `verifylive.go:118`), a173 시험 `TestStructuralUnmeasuredSentenceIsSaidOnce`.
- 발견과 결정(design ★):
  1. 단계 데이터는 구조체가 아니라 `WriteSteps` 문자열이다. verifylive(실계좌 검증 패키지)를 고쳐 단계별 출력 함수를 만드는 대신, 알려진 단계 ID 경계에서 무손실 분할한다(이으면 원문과 바이트 동일·조각 수 == 단계 수·실패 시 종전 `<pre>`). 태그를 콘솔에서 재계산하지 않으므로 CLI 와 어긋나지 않는다.
  2. 개요 요약(`dl.console-summary`)과 구조적 미측정 1회 처리는 a173 이 이미 넣었다. 정의본의 "같은 글 3회" 는 a173 이전 값이다 → 렌더로 다시 세고, 이미 1회면 회귀 가드로 쓴다.
  3. FLM 면제 문장을 거뒀다.
- 남은 위험: 분할은 `WriteSteps` 의 머리 줄 형식(`%-22s `)에 기대므로 형식이 바뀌면 fail-safe 로 종전 화면이 된다 — 시험이 그 순간을 FAIL 로 알린다. 개요 요약을 세 묶음으로 재배치하는 범위는 a173 이후 실측으로 다시 좁힌다.
- 결론: proposal-freeze 승인. 착수 가능.

## 구현 사전 확인 (2026-10-27, Terra)

- research-first: 저장소에서 `verifyPage`·`renderVerify`·`verifylive.WriteSteps`·기존 콘솔 시험을 검색함. `StepItems` 또는 같은 무손실 단계 분할 구현은 없고, `verifylive.WriteSteps`의 단계 ID·고정 폭 머리 줄이 유일한 정본임을 현재 HEAD에서 확인함. 기존 Go 표준 라이브러리 `fmt`·`strings`만으로 충분하므로 외부 패키지·MCP·레지스트리·GitHub 탐색은 중단함. 의존성·라이선스·CVE 변화 없음. 결정: 새 파일의 최소 메서드를 구현함.
- CodeGraph: a176 worktree는 초기화되지 않아 `codegraph status .`가 `Not initialized`으로 끝남. 인덱스 초기화는 하지 않고 현재 소스의 `rg`·원문으로 범위를 확인함. CodeGraphContext 보고서는 `/tmp/tossos-a174-close-20261026`을 가리켜 현재 worktree 근거로 쓰지 않음. GBrain은 현재 프로젝트 동기화 상태를 제공하지 않음.
- Pre-Edit Gate: change `a176`, task 2~3. 대상은 새 파일의 `verifyPage.StepItems` 및 템플릿·CSS 상수뿐임. 기존 Go 함수 본문·`internal/verifylive`·`renderVerify`·`handleVerify`·`handleOverview`·패널 함수는 편집하지 않음. Function Logic Map: not-applicable — 기존 Go 함수 본문 편집 0, 새 leaf 메서드와 새 시험 함수만 추가함. Branch Test Map: not-applicable — 기존 분기 변경 0, 새 계약 시험으로 StepItems 실패 경계만 고정함. 기존 함수 편집이 생기면 즉시 중단하고 gate ⑤ 번들을 먼저 작성함. 기존 테스트 변경 대신 새 시험 파일을 작성함. 설정·DB·journal 변경 없음. 실계좌 주문·승인·운영 토글 영향 없음.

## 구현 후 경량 리뷰 (2026-10-27, Terra)

- 적용성: UI 변경이므로 impeccable·fixing-accessibility를 적용함. `impeccable detect --json --target internal/console/templates.go` 결과는 `[]`임. `design-critic` 실행 파일은 PATH에 없어 수동 렌더 8장과 정적 접근성 검토로 대체함.
- gstack review: `gstack` 실행 파일과 프로젝트 `.claude/skills/review/checklist.md`가 없어 자동 워크플로는 실행할 수 없었음. 설치된 gstack `review/checklist.md`의 CRITICAL(SQL·동시성·신뢰 경계·셸 주입·enum) 및 INFORMATIONAL(완결성·시간 창·형 강제·뷰·배포) 항목을 현재 diff와 새 파일에 한정해 수동 적용함. 발견: 0건. 새 enum·상태·외부 호출·비동기·DB·셸·배포 변경 없음. `git diff --check 4b868a18d`도 통과함.
- 안전 계약: `internal/verifylive`, `pages.go`, `renderVerify`·`handleVerify`·`handleOverview`·패널 함수 본문을 편집하지 않음. `/verify` 시작·승인·중단 폼, CSRF, 승인 프롬프트는 `a176_verify_forms.golden.html` 렌더 fixture로 고정함.
- 변이: `templates.go`의 단계 본문 `<pre>{{.Body}}</pre>`를 한 번 제거하면 `TestVerifyStepsAreFoldableAndListTextIsPreserved`가 14개 단계 모두에 대해 FAIL함(`analysis/mutation-output.txt`). 원복 뒤 심볼 수 `1`, 같은 시험 PASS를 확인함.
- 판정: 구현 범위 안에서 blocker 없음. `/verify` 접힌 전체 높이 2047px는 목표 1800px을 넘음. 설계가 요구한 머리말 `<pre>`와 44px 단계 표식을 압축·은닉하지 않아 남은 측정 위험으로 보류함.

## Manager 수정 라운드 1 (2026-10-27, Terra)

- P2 검증 보류: 요약의 `미체결 · 살아 있는 주문` 행에 기존 조건 `{{if .Snap.Open.VerifyHeld}}`와 문구 `검증 중 — 갱신 보류`를 복원함. 같은 행만 추출하는 새 시험은 다른 프로세스의 활성 run-lock fixture에서 해당 문구를 확인함.
- P2 오늘 소진분: 기존 요약은 `.Structural`을 `미측정`이라고만 쓰고, 계산 가능한 축도 `측정`이라는 글자만 썼음. 구조적 세 축에는 `reading`·열거 사유 코드가 없으므로 코드처럼 꾸미지 않았음. 대신 이미 계산된 비구조 축의 `guardianAxisRow.Consumed`를 `cell` 템플릿으로 그대로 렌더하고 상세 링크를 둠. 값이 미측정이면 기존 `unmeasured` define을 통과하므로 a173의 `class="muted unmeasured"` 표본·사유 코드 검사에 포함됨. 기존 a173 시험이 놓친 이유는 옛 행이 `muted`만 사용해 표본 정의 자체에 들어오지 않았기 때문임.
- P2 원인 우선순위: `causes`·`journalstate`를 개요 요약 앞의 원래 위치로 되돌림. 정상 fake-console fixture는 원인 배너 0개와 데스크톱 요약 하단 644px을 기록함. 실제 빈 config-dir 격리 콘솔 장애 상태는 배너 1개와 1036px(390px 1581px)으로 900px을 136px 넘음. 초과를 줄이려고 여백·글꼴·원인 문구를 압축하지 않았으며, 정상·장애의 모바일·다크 수치는 측정 JSON에 남김.
- P3 Guardian: 요약은 현재 `Currency`와 시장별 잔여물만 갖고, `Limits`는 의미가 서로 다른 다섯 값의 배열이라 단일 한도 값으로 옮길 수 없었음. `/verify-console` 링크도 이번 수정 범위에서 설정 링크로 바꾸지 않았음. 설계의 "값 + 설정 링크" 충족 여부는 Manager 4.5 판단으로 남김.
- P3 `/verify`: 빈 config-dir 격리 콘솔에서 접힌 높이 2047px(390px 2368px, 목표 1800px 초과)을 재확인함. 단계 머리말 `<pre>`와 44px summary를 유지했으며 높이만 맞추기 위한 축소는 하지 않았음.
- 안전·검증: `internal/verifylive`, 기존 핸들러·패널 함수, `/verify` 시작·승인·중단 폼과 fixture는 변경하지 않았음. 임시 fake-console 측정 서버는 test 종료와 함께 포트 38327·38328을 해제함. Playwright는 GET만 수행했고 버튼 클릭 0회.
