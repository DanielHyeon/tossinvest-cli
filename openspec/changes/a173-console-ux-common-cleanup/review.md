# a173 review

## proposal-freeze (2026-10-25, 경량: validate + Manager 셀프리뷰, docs/WORKFLOW.md 「위험 등급 가중」 UI change)

- 보이스: Manager(Fable 5.1) 단독. 적대적 Eng 보이스는 해당 없음(주문·위험관리·원장·reconciliation 미접촉).
- 입력: 2026-10-25 UX 전수 검사(22개 화면 데스크톱 전수·모바일 2장), 정찰 보고(Sonnet, 읽기 전용, 46 tool uses), `operator-console` spec 574~612·632·706·851~880·1374~1442.
- 발견과 결정:
  1. 검토 보고서의 "설정 5묶음"·"엔진 버튼 이동" 권장은 spec 851·634·662 와 충돌 → 철회, 사람에게 보고(proposal 「사람 결정 기록」). a173 범위에는 영향 없음.
  2. `console_prose_test.go:29` 가 `<details>` 안 `.notice` 를 금지 → 접힌 경계·기술 정보 블록에는 상자를 두지 않는 설계로 수정(design §2.2·§2.3).
  3. `TestEveryScreenIsCalledOneThing` 이 `<h1>` 텍스트를 비교 → 경계 배지는 `<h1>` 밖(design §2.3).
  4. 스펙 1374 가 사유를 "열거해 셀 수 있어야" 한다고 요구 → 원인 1회 설계는 자유 문장을 줄이되 사유 코드는 모든 칸에 유지(design §2.1). 스펙 1442 의 "안내 1회" 와 일치.
  5. 측정 하네스는 change 디렉터리에 커밋(기억 `harness-cited-but-never-committed`), 채운 fixture 로도 센다(기억 `universal-check-passes-on-an-empty-sample`).
- `openspec validate a173-console-ux-common-cleanup --strict`: 아래 「validate 기록」.
- 결론: proposal-freeze 승인. 구현 착수 가능(worktree/a173, Sonnet 5.5).

## validate 기록
- 2026-10-25: 델타 없음으로 FAIL → `specs/operator-console/spec.md` ADDED 5건 추가 후 재실행(결과는 구현자가 1.1 전에 다시 확인해 여기 적는다).

## 구현 후 (2026-10-25, 구현자 Sonnet 5.5)
- `openspec validate a173-console-ux-common-cleanup --strict` 는 이 구현자가 다시 돌리지 않았다(Manager 확인 필요).
- 기존 함수 편집 7건(번들 필요): tasks.md 3.1 목록. FLM/BTM 면제 불가 — 종결 단계에서 Opus 가 번들 작성.
- 고정 리터럴 시험 수정 1건: tasks.md 「고정 리터럴 시험 수정 기록」.

## 구현 리뷰 (gstack `review` 경량판, 2026-10-25 — diff 직접 읽기 + Manager 판정)
스킬의 Review Army·adversarial·Greptile·텔레메트리는 돌리지 않았다. 차단급(P0/P1) 없음. 범위 점검: DRIFT 없음.
1. 배너 등급(P2) — **수용·수정**: `causes` 를 읽기 실패(`journal_unreadable`·`broker_read_failed`·`config_unreadable`·`journal_value_unparsable`·`discovery_unreadable`)=`.danger.cause-banner` 와 상태 설명(`seam_unwired`·`never_fetched`·`verify_suspended`)=`.aside.cause-banner` 로 분리. 시험 `TestCauseBannerSplitsBlockingFromStateNotes`.
2. `causesOf` 포인터 순회(P2) — **수용·위험 기록**: `causes` 는 값 스냅숏 뷰 모델(overview·orders·signals `Snap`)에서만 호출. `/verify*` 의 `*runView` 가 복사본인지는 확인하지 않았다 — 그래서 그 화면에는 걸지 않는다. `docs/console-design.md` 에 규칙 한 줄.
3. 복합 `:not()`(P2) — **수용·수정**: `main p, main li, main details > :not(summary)` + `td p, th p, .console-status p { max-inline-size: none }` 두 규칙. 재측정 가장 긴 `<p>` 672px 유지.
4. advisory `title` 한계(P3) — **수용**(스펙 574~612 는 advisory 표기만 요구, 접근성 한계는 review-after.md 기록 유지).
5. `/orders` 영문 원문 2곳(P3) — **수용·수정**(원인 1회 규칙은 a173 범위): `orderbroker` 두 곳은 원문을 빼고 `title` + 위 배너로. 기존 시험은 그 자리 원문을 단언하지 않아 고정 리터럴 수정 0.
6~8. 시험 서버 다수 생성·문장 쪼개짐 시 FAIL 방향·`handleReport` 문자열 치환 의존(P3) — **수용**(회귀는 `/report` 본문 경로 0 시험이 잡는다).
9. `<nav>` 안 `<details>`(P3) — **a175 이월**.
- 추가 확인: `/verify-console` 의 `<details class="scope">` 안에 경계 문단("이 화면은 로컬 파일만 읽는다…")이 실제로 들어 있음(렌더 덤프로 확인).
