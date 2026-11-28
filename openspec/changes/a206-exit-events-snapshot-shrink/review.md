# a206 review

## 1단계 경량 리뷰 r1 (2026-11-27, 1회 — 상한 2회 중 1회 사용)

- 리뷰어: Codex CLI 0.165.0(`codex exec --sandbox read-only`, 작성자 Claude 와 다른 계열). 범위: proposal · design · tasks · spec 델타 ·
  claims · evidence-reconciliation · a206_measure.py · Story. 지시: 읽기 전용, 저널 · DB 접근 금지, 새 가드 · 기능 제안 금지.
- 질문 4: (1) 닫힌 포지션 행의 스냅숏을 다시 읽는 생산 리더 누락, (2) C1~C7 의 file:line · 분기 정확성,
  (3) 제안 정비가 깨는 불변식 · 유일 인덱스 · 리더, (4) 측정 스크립트가 D1 과 같은 대상을 고르는지 · 원본에 안전한지.

| # | 등급 | 지적 | 처분 |
|---|---|---|---|
| 1 | — | 리더 누락 없음(SQL · CTE 검색) | 확인 |
| 2 | P3 | C2 는 `full()` 이 20 열 중 **필수 부분집합**만 요구함을, C6 은 Saved 가 없으면 `encode` 가 NULL 을 돌려줌(:803-804)을 빠뜨림 | 수리 — claims C2 · C6 문구 정정 |
| 3 | — | 불변식 깨짐 없음: completed 거절이 decision 중복 검사보다 앞(exit_state.go:463 · :494), 유일 인덱스는 NULL 제외(exit_snapshot.go:61), trade outcome · backfill · provenance · a166 replay 가 쓰는 열은 남음 | 확인 |
| 4 | P2 | 스크립트 대상 술어가 D1 과 다름(`exit_states.completed=1` 누락, 확정 PLACE 만 제외) | 수리 — 술어를 D1 그대로(CLOSED + completed=1 + broker_order_id 가진 attempt 전부 제외)로 바꾸고 재측정. S1 · S2 · S3 수치 불변 |
| 5 | P2 | 원본이 작업 디렉터리 안이면 시나리오 사본이 덮어씀 · 지움, `refuse_live` 가 운영 디렉터리 자체를 허용 | 수리 — 두 경우 모두 거절(측정 하네스 안의 안전 검사이며 생산 가드 아님) |
| 6 | P2 | S4 가 effective 를 먼저 비워 `lag(effective)` 가 NULL 이 됨 → saved 중복 제거가 측정에서 빠짐 | 수리 — 순서를 바꿔 재측정: 203,235,328 → **175,517,696** B |
| 7 | P2 | design 의 11월 건수(172 · 3)가 측정 출력(86 · 1)과 다름 | 수리 — 원인: 1차 ad-hoc 집계가 무장 행과 같은 intent 를 가리키는 PROPOSAL_REFUSED 훅 행까지 세어 두 배가 됨. 측정 스크립트에 11월 일별 무장 행 집계를 넣고 그 출력(10 · 64 · 12 · 1)으로 정정 |

- 재측정: `analysis/a206_measure.out.txt`(05:3x Z), 리더 확인 `analysis/a206_reader_check.out.txt` 도 수리된 술어로 재실행
  (사본 시점 CLOSED 129 · OPEN 1 — 그 사이 포지션 하나가 닫힘; S1 오류 124/129, S2 오류 0).
- 2회차는 열지 않음(P1 0, P2 넷 모두 측정 · 문구 수리로 닫힘).

## Function Logic Map

Function Logic Map: not-applicable — 1단계는 기존 Go 함수 본문을 편집하지 않음(생산 · 시험 diff 0). 분기 주장은 FLM 번들이 아니라
`analysis/ast-evidence/`(logic-map AST 11 개)를 근거로 씀. 2단계는 새 파일 · 새 함수만 목표(design D5).

## 변이 원장 · 다중 적대 리뷰

not-applicable — 1단계는 문서 · 측정 하네스뿐(비례 원칙). 2단계 구현은 원장 쓰기라 High-risk 이며 그때 정함.

## 2단계 구현 리뷰 (2026-11-27, 2회 — 상한 2회 모두 사용)

- 리뷰어: Codex CLI 0.165.0(`codex exec --sandbox read-only`, `</dev/null`; 작성자 Claude Opus 5.5 와 다른 계열). 범위: 새 파일 4
  (`internal/journal/a206_exit_evidence_compact{,_test}.go` · `cmd/tossctl/a206_exit_evidence_compact{,_test}.go`) · `engine.go` 등록 ·
  `help_convention_test.go` 표지 집합 · spec 델타 · proposal · design. 지시: 읽기 전용, 저널 · DB 접근 금지, 새 가드 · 기능 제안 금지.

### r1

| # | 등급 | 지적 | 처분 |
|---|---|---|---|
| 1 | — | 대상 SQL 이 D1 과 같고 D8 SQL 이 문자 그대로임. 앵커 없는 `max()` 의 NULL 은 `NOT IN` 전에 걸러짐. JOIN 중복 없음(PK) | 확인 |
| 2 | P2 | `wal_checkpoint(TRUNCATE)` 가 열린 읽기 연결 때문에 `busy=1` 을 돌려줘도 성공(`outcome=ok`)으로 보고 | 수리 — busy≠0 이면 「증거는 커밋됨, 파일은 다음 checkpoint 에 줄어듦 — 콘솔을 멈추고 다시 실행」 오류 |
| 3 | P2 | 잘못 지정한 디렉터리에서 `journal.Open` 이 새 원장을 만들고 「정비 성공」 을 보고 | 수리 — `--apply` 는 `journal.Open` 전에 파일 존재를 확인하고 없으면 거절(`TestA206CompactApplyRefusesAMissingJournal`). 옛 스키마를 `journal.Open` 이 마이그레이션하는 것은 유지 — 표준 writer 경로(엔진 · reconcile-resolve)이고 자기 pre-migration 백업을 만듦 |
| 4 | P2 | root `PersistentPreRunE` 가 승인 검사보다 먼저 돎(세션 · 설정 읽기, 알림 캐시 쓰기) | **기각** — 그 훅이 쓰는 것은 update-check 캐시뿐이고 원장 · 주문 · 토글에 닿지 않음. 같은 가족 명령(attempt-resolve · mode-release)도 RunE 에서 같은 방식으로 검사함. r2 가 기각을 타당하다고 확인 |
| 5 | P3 | design D4 「실패하면 롤백」 이 커밋 뒤 실패(VACUUM)에는 거짓 | 수리 — 커밋 전 / 커밋 뒤 실패를 나눠 적음 |
| 6 | 시험 공백 | 앵커 후보가 하나뿐 · 같은 intent 에 연결 + 미연결 attempt 공존 사례 없음 · root 훅 미검출 | 앞 둘 추가(c0 · c1 두 후보 — `max→min` 변이 CAUGHT; l1 에 브로커 번호 없는 attempt 추가). 셋째는 4 기각과 같음 |

### r2

- 지적 0(P1 · P2 · P3 없음). 수리 2 · 3 · 5 와 추가 시험이 맞고, 새 결함 없음, 4 의 기각에 원장 · 안전 영향 근거 없음. 한계로 checkpoint busy 갈래의 직접 시험이 없다고 적음(아래 「남은 위험」).

## 2단계 RED → GREEN · 변이

- RED: (1) 빌드 실패(새 심볼 없음, `undefined: a206EvidenceColumns` 등), (2) 빈 stub(쓰기 0)에서 행동 RED — `plan = {Targets:0 …}, want 3 targets` ·
  「emptied row N reads as … Recomputed:{Snapshot:0x…}, want legacy_event」, (3) 명령: 「`tossctl engine exit-evidence-compact` is not attached」 ·
  「unexpected mutating=true」(표지 집합 추가 전).
- GREEN: 아래 「선택 시험」.
- 변이(구현 파일 사본에서 한 번에 하나, 끝에 `cmp` 로 원복 확인): CAUGHT — keep 제외 삭제 · 연결 제외 무력화 · `state='CLOSED'` 삭제 ·
  `completed=1` 삭제 · NULL 거름 삭제 · 「이미 전부 NULL」 조건 삭제 · 20 열 중 `arm_suppressed_reason` 누락 · UPDATE 무력화 · 앵커 `max→min` ·
  `os.Stat` 거절 무력화. SURVIVED 2: 앵커 술어의 `effective_source <> 'recomputed_repeat'` 삭제(동등 — a207 압축 행은 늘 `recomputed_snapshot_json`
  NULL 이라 앞 항이 같은 행을 거름, D8 문자 일치를 위해 남김) · 무결성 검사 무력화(건강한 픽스처에서는 실패할 수 없음 — 손상 주입은 하지 않음).
  처음 판에서 `state='CLOSED'` · `completed=1` · `arm_suppressed_reason` 셋이 살아남아 픽스처에 반쪽 닫힘 포지션 둘(K · Q)과 무장 보류 판정 행을 더함.

## 선택 시험 · 미실행 범위

- `go test -count=1 -timeout 30m ./internal/journal` · `go test -count=1 -timeout 30m ./cmd/tossctl`(패키지 전체 — 결과는 tasks 2.5), `make lint`,
  `openspec validate a206-exit-events-snapshot-shrink --strict`, `check_analysis.py --change a206-exit-events-snapshot-shrink`.
- 결과(최종 코드): `./internal/journal` ok 596.2 s rc 0 · `./cmd/tossctl` ok 74.8 s rc 0 · `make lint` rc 0(gofmt · vet · 태그 vet) · strict validate valid ·
  check_analysis rc 0. 첫 패키지 판에서 `TestTheReadOnlyHandleHasNoWriteMethods`(ReadOnly 허용 메서드 census)가 FAIL → 근거 주석과 함께 집합에 더하고 재실행 PASS.
- 이유: 새 코드는 journal(정비 함수)과 cmd/tossctl(명령 · 등록)에만 있음. 기존 리더 · 쓰기 함수 무편집이라 다른 패키지 호출자 없음.
- 미실행: `make test` 전체 · `make test-seams` · `make test-race` · `make gate`(사용자 요청 없음). 운영 저널 대상 실행(사람 결정 3 — 사람이 함).

## Function Logic Map

- 수정된 기존 함수 3(gate ⑤ required 3): `cmd/tossctl/engine.go:newEngineCmd`(등록 한 줄), `cmd/tossctl/help_convention_test.go:TestMutatingAnnotationOnTradeCommands`
  (표지 집합 한 줄), `internal/journal/readonly_test.go:TestTheReadOnlyHandleHasNoWriteMethods`(허용 메서드 집합 한 줄 — 패키지 전체 시험에서
  「unlisted method "ExitEvidenceCompactionPlan"」 FAIL 을 보고 근거 주석과 함께 더함). 번들 `analysis/function-logic/` 셋(편집 뒤 AST — 사후 작성).
  새 함수는 전부 새 파일(design D5). 기존 리더 · 쓰기 함수 편집 0. `check_analysis.py` rc 0 「evidence complete」.
- a166 census(`internal/accountbudgetruntime/testdata/a166-wiring/shapes.json`)에 `engine.go` · `newEngineCmd` 없음 — 확인함.

## 변이 원장 · 다중 적대 리뷰

- 원장 쓰기라 High-risk. 변이는 위 표(12 판), 적대 리뷰는 다른 계열 1 보이스 × 2 회(Manager 지시 상한). 다중 보이스는 Manager 지시 범위 밖이라 열지 않음.

## 남은 위험

- checkpoint busy 갈래(콘솔이 열려 있을 때)는 직접 시험 없음 — 한 줄 분기이고 r2 가 코드로 확인함. 실행 창에서 콘솔도 멈추면 닿지 않음.
- 엔진 정지 = 손절 없음. 실행은 두 시장이 닫힌 창에서 사람만(3.1).
- 백업(≈290 MB)은 사람이 검증 뒤 지움 — 그 전까지 디스크가 오히려 늘어남.
- 재발: R-c(브로커 확정 거절)가 길게 돌면 다음 정비까지 다시 쌓임. a207 압축이 반복 무장 행의 JSON 을 줄이므로 성장 속도는 a207 배포 뒤 다시 잴 것.

## 운영 실행 (3.1, 2026-11-28)

- 실행: 2026-11-28(토) 아침 두 시장이 닫힌 창에서 Manager 가 엔진 · 콘솔 정지 → `tossctl engine exit-evidence-compact --apply` → 엔진 · 콘솔 재기동
  (사용자 「2 승인」 2026-11-28). 재기동 뒤 정상은 Manager 보고이며, 팀메이트는 컨테이너를 보지 않았음 — 간접 증거로 실행 뒤 journal.db
  mtime 19:50 KST · `-wal` mtime 20:12 KST(실행 10:01 KST 뒤에도 writer 가 씀).
- 팀메이트 재측정(2026-11-28 20:1x KST, 읽기 전용 — `stat` 과 audit 줄 파싱뿐, journal.db 를 sqlite 로 열지 않음):
  - `stat -c '%n %s' ~/.config/tossctl/journal.db ~/.config/tossctl/journal.db.pre-a206-20261128T010155Z.bak`
    → journal.db **68,947,968 B**, 백업 **290,603,008 B**(mtime 10:01:56 KST). `-wal` 은 4,120,032 B(재기동 뒤 엔진이 쓰는 중).
  - `~/.config/tossctl/audit.log` 에서 `action == "journal.exit_evidence_compact"` 줄만 골라 `at` · `action` · `new` · detail 의 `outcome=` 만 출력
    (operator · approval · note 는 출력하지 않음) → 1 줄:
    `2026-11-28T01:01:58.617798634Z journal.exit_evidence_compact rows=54883 rows_after=54883 targets=28025 emptied=28025 remaining=0
    kept_repeat_rows=0 snapshot_bytes=181264442 file_bytes=297348840 file_bytes_after=68947968 outcome=ok`
- 전달 수치와 대조: 비운 행 28,025 · 297 MB → 69 MB · 백업 291 MB 모두 일치. 표기 차이 둘 — (1) 297 MB 는 정비 전 journal.db **+ WAL** 합
  (`journalFileBytes`), (2) 백업이 그보다 작은 것(290.6 MB)은 `VACUUM INTO` 사본이라 WAL · 빈 페이지가 빠지기 때문. 사본 실측(2.5)
  293,183,488 → 68,947,968 B 와 끝 크기가 같음. 행 수 불변(54,883 → 54,883 — 행을 지우지 않고 증거 열만 비움, design D1 · D2).
- 남은 것: 3.2 백업 처분(사용자 결정 「월요일 실측 뒤 삭제」 — 아직 안 함), 3.3 main 병합 · archive.
