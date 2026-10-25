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
- gstack review 미실행(tasks 4.5 미체크).
