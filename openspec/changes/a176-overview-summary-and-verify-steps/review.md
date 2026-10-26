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
