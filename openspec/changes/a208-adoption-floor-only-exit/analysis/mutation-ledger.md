# a208 변이 원장(design 8.4)

- 하네스: `analysis/a208_mutation.py`(이 디렉터리, 커밋됨). 사본 = 연결 worktree `/home/daniel/tossos-work/a208-mut`(detached `a547fa67f`) — 작업 worktree 무접촉.
- 절차: 시작 sha 단언 · 사본 깨끗함 단언 → 무변이 대조군(engine · config · console `-run TestA208`) GREEN → 변이 하나씩 적용 → 대상 시험 → `git checkout` 원복 → 원본 바이트 동일 · `status --porcelain` 빈 값 단언. 한 번에 한 판, `GOFLAGS=-trimpath`.
- 결과(2026-11-28): **12 / 12 CAUGHT**, 대조군 GREEN.

```text
control GREEN: engine · config · console TestA208
M1 CAUGHT internal/app/engine/a208_adoption_floor_only.go :: FAIL
M2 CAUGHT internal/app/engine/a208_adoption_floor_only.go :: FAIL
M3 CAUGHT internal/app/engine/a208_adoption_floor_only.go :: FAIL
M4 CAUGHT internal/app/engine/a208_adoption_floor_only.go :: FAIL
M5 CAUGHT internal/config/engine.go :: FAIL
M6 CAUGHT internal/config/adoption_io.go :: FAIL
M7 CAUGHT internal/config/adoption_io.go :: FAIL
M8 CAUGHT internal/config/engine.go :: FAIL
M9 CAUGHT internal/config/engine.go :: FAIL
M10 CAUGHT internal/console/settings.go :: internal/console/settings.go:244:2: declared and not used: current
M11 CAUGHT internal/app/engine/interlock.go :: FAIL
M12 CAUGHT internal/app/engine/interlock.go :: FAIL
SUMMARY 12 / 12
```

| M | 변이 | 잡은 시험 |
|---|---|---|
| M1 | 정책 선택이 토글 무시 | TestA208FloorOnlyAdoptionRecordsTheRunnerCopy · TestA208FloorOnlyAdoptionReplacesRatchet |
| M2 | OFF 에서도 RUNNER | TestA208OffAdoptionIsUnchanged |
| M3 | 알림 문구 항상 ON | TestA208OffAdoptionIsUnchanged |
| M4 | 알림 문구 항상 OFF | TestA208FloorOnlyAlertSaysNoPartialTakeProfit |
| M5 | mergeAdoption 이월 누락 | TestA208FloorOnlyExitRoundTrips |
| M6 | LoadRawEngineAdoption 이월 누락 | TestA208 (config) |
| M7 | SaveEngineAdoption 이월 누락 | TestA208FloorOnlyExitRoundTrips |
| M8 | omitempty 제거 | TestA208FloorOnlyExitRoundTrips |
| M9 | 거부 블록이 토글 유지 | TestA208ARefusedBlockClearsFloorOnly |
| M10 | handleSettingsSave 이월 누락 | 컴파일 실패(`current` 미사용)로 잡힘 — 「잡힘」이지만 시험이 아니라 컴파일러가 잡음. 대체 변이 M10b(`FloorOnlyExit: false,`)는 아래 |
| M11 | 기동 audit 행 누락 | TestA208FloorOnlyIsAuditedAtStartup |
| M12 | audit 값 상수 false | TestA208FloorOnlyIsAuditedAtStartup |

## 리뷰 1회차 뒤 재확인

- T7 「키 없음」 하위 시험을 문자열 부재 → **파일 바이트 동일**로 강화(codex P3). M8 을 강화된 시험으로 다시 돌림: CAUGHT(`an off toggle changed the saved bytes`), 원복 뒤 `git diff --quiet` 확인.
- M10b(`FloorOnlyExit: current.FloorOnlyExit && false,` — `current` 는 쓰이되 값은 버림): **CAUGHT** — `TestA208ConsoleSaveKeepsFloorOnly`(`floor_only_exit was true before the save and false after`). 원복 뒤 `git diff --quiet` 확인.
