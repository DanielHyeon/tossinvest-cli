# 8.5 적대 리뷰 합본 (codex r2 · 보이스 1 · 2 · 3) — 2026-10-04

대상: 2.3(65444341) · 8.8.4 로트 B(0b441270). 판정은 Manager.

**Manager 최종 판정(4판 합본): HOLD → 아래 응답 로트 완료 · 재검증 시 SHIP.** P0 은 네 판 모두 0.

| 출처 | 판정 | 저장소 무변경 | ~/.codex |
|---|---|---|---|
| codex r1 | **무효**(read-only `/tmp` 로 Go 빌드 불가 · `~/.codex` 규칙 위반) — 기록 보존 | — | 위반(Manager 보고) |
| codex r2 | HOLD — P1 1 · P2 3 | 전후 status 동일(`codex-run-r2.txt`) | 준수 |
| 보이스 1 (fail-open) | SHIP — P0/P1 0 · P2 3 | HEAD 556c3c1f · untracked 3 동일 | 준수(첫 줄 명기) |
| 보이스 2 (동등성) | HOLD — P1 1 · P2 3 | HEAD 556c3c1f · status 동일 | 준수(첫 줄 명기) |
| 보이스 3 (증거 품질) | HOLD — P1 2 · P2 5, 변이 72 | HEAD 556c3c1f · untracked 3 동일 · `git diff` 빈 출력 | 준수(첫 줄 명기; 사본의 저장소-추적 `.codex/` 는 이름만 봄) |

## 발견 전행 · 처분 · 핀

| ID | 등급 | 출처 | 내용 | 처분 | 핀 |
|---|---|---|---|---|---|
| F1 | P1 | 보이스 2 P1 · 보이스 1 P2-1(같은 기전) | 관문 조기 계산으로 관문 스냅숏이 제안 적재 시간만큼 낡음 — 적재 중 취소 · 철회가 판정에 안 보임(옛 FAMILY_GATE_CLOSED → 새 READY). 「값 동등」 주장 거짓 | 응답 로트 ①: **Y**(판정 자리 재계산 복원 + 조기 carry 유지) | 브리프 §6 핀 1~5 · 주장 정정 3곳 · 잔여 비대칭 기록 |
| F1-x | 후보 | 보이스 1 P2-1 제안 | `FinalAuthorityCheck`(strategy_dispatch_cycle.go:212)가 만료만 재확인 — 핀 · 철회 재확인은 양판 공통 경합 자체를 닫음 | ROADMAP: (d) 로트 후보 | — |
| F2 | P1 | codex r2 P1 | 조기 활성화 읽기가 취소된 `collect` 를 붙잡음 — 결함은 적재기 넷 공통 기존 패턴, 이동 증분은 실패 주기 +1 읽기 | 이동 유지 + **(d) 별도 로트**(적재기 넷 동시, 「핀 선언 전 착지」 면제 불가 선행). (d) 는 F1 을 못 닫음 | 적재기별 stalled-reader seam |
| F3 | P1 | 보이스 3 P1-2 | sentinel **배타성** 미핀 — 거절 ~17곳 중 Undeclared 를 몰래 같이 감는 `%w%.0w` 변이 13곳 생존(대표 L556 은 router · 엔진 full · tossctl 전부 생존). 살면 선언된 시장의 몸통 결속 불일치가 기존 경로로 열림(:132 가 Undeclared 하나로 가름) | 응답 로트 ⑤ **최우선**: error-fields 표 40 모양 전부에 `errors.Is(err, Undeclared) == (want == Undeclared)` | L556 재실행 격추 확인 |
| F4 | P1 | 보이스 3 P1-1 · 보이스 1 P2-2(m602) | 서술자 `effective` 유효성 항(T602 = R4 = m602) 미핀 — `Effective:"MAYBE", Desired:ON` 이 verified=true 로 수락됨(보이스 3 실측; 보이스 1 은 lookup 결과 DORMANT 라 fail-safe 로 봄) | 응답 로트 ⑥: 실패 모양 추가(거절 + 필드명) | R4 · m602 재실행 CAUGHT |
| P2-a | P2 | codex r2 | nil getenv → INTERNAL_FAILURE | 응답 로트 ②: `loadFamilyActivation` nil-안전(→ Unavailable) | 다중 실패 조합 행동 시험 1 |
| P2-b | P2 | codex r2 | descriptor 오류에 `lane_id` 원문 · 개행 | 응답 로트 ③: `descriptors[%d]` + 필드명, 원문 배제 | 개행 주입 시험 |
| P2-c | P2 | codex r2 · 보이스 2 P2-3 · 보이스 3 P2-4 | 골든 `rvol_counterfactual_ppm[0]` ↔ 리터럴 미결속. 보이스 3: 입장 경로 자리는 입장이 RVOL ≥ 1.5M 을 함의해 1.5M 이하 전 구간이 동등 → 지킬 행동 없음 | 응답 로트 ⑩: **B7 자리만** 골든 행동 결속; 입장 경로는 생산 단순화 없이 FLM 에 동등성 기록 | 골든 값 · 1 ppm 아래 경계(B7) |
| P2-d | P2 | 보이스 2 P2-1 · 보이스 1 관측 | 읽기 결함이 둘째 `%w` 로 `ErrProductionRouteUnavailable` 도 만족 | 응답 로트 ④: 안쪽 `%v` | `errors.Is(err, ErrProductionRouteUnavailable)==false` |
| P2-e | P2 | 보이스 2 P2-2 | config `market` 항이 판정을 못 바꿈(편집 전부터) | 응답 로트 ⑫: FLM 기록 — 가림 가드 둘(디렉터리 읽기 실패 · 몸통 `validMarket`) 명명 + 닫아 두는 등식 | 메시지 시험이 유일 핀임을 명기 |
| P2-f | P2 | 보이스 1 P2-2(m569) · 보이스 3 R3=T569 | `issued == expires` 항 — 판정 불변, sentinel Unavailable→Expired 이동 | 응답 로트 ⑥: 실패 모양 추가(필드명 + sentinel) | m569 재실행 CAUGHT(태그 엔진 포함) |
| P2-g | P2 | 보이스 1 P2-3 | 닫힌 시장 ON 레인 REFUSED / ARBITRATION_SEAL_MISMATCH 표기 — 진단 전용 | ROADMAP: 운영자 표면 행에 「no input this wave」 detail 후보 | — |
| P2-h | P2 | 보이스 3 P2-1 | B10(FAMILY_GATE_CLOSED) · B12(큐 넘침) 닫힘의 carriage 미단언(E9 · E7 생존). E5/E6(B11 · B14)은 도달 불가 공간 → census-only 타당 지지 | 응답 로트 ⑦: 두 시험에 `familyActivation().Verified()` 단언 + BTM 에 「선택자-대입 변이는 census 도 행동도 못 본다」 한계 기록 | E7 · E9 재실행 CAUGHT |
| P2-i | P2 | 보이스 3 P2-2 | 수명 항 셋(T565 · T566 · T569)이 파생 거절에 가려 생존 — 거절 유지, 종류 · 필드명 표류 | 응답 로트 ⑧: 모양 셋(issued non-canonical · expires non-canonical · expires ≤ issued) | T565 · T566 · T569 재실행 CAUGHT |
| P2-j | P2 | 보이스 3 P2-3 | 2.3 반사실 픽스처 맹점: buffer 가 1틱으로 접힘(A4) · 첫 봉만 변이(A5) | 응답 로트 ⑨: ATR=50 경우 1 + 비첫봉 경우 1 | A4 · A5 재실행 CAUGHT |
| P2-k | P2 | 보이스 3 P2-5 | Load B10(:488) BTM 행 「도달」 인용 부정확 — 전수 커버리지 count=0 | 응답 로트 ⑪: 「도달 불가 — census/검토로 닫음」 으로 정정 | — |

## 응답 로트 변이 원장에 남길 4판 생존 변이 재판 표(요구)

L556(F3) · R4/m602(F4) · m569/T569(P2-f · P2-i) · T565 · T566(P2-i) · E7 · E9(P2-h) · A4 · A5(P2-j) — 각 행: 출처 판 · 원 판정(SURVIVED) · 응답 로트 후 판정 · 잡은 시험. 동등 변이로 종결하는 것(A1 · A2 · A7 · E5 · E6 · T450/R8b · R17 · L438)은 사유와 함께 별행.

## 각주

- `brief.md:66` 「값 동등」 은 리뷰 입력이라 기록 보존 — F1 로 거짓 판명, Y 하에서 판정 자리 동등성은 구성상 성립.
- 보이스 1 의 아카이브 사본 무태그 엔진 2 실패(a111 둘)는 사본 아티팩트(무변이 대조군도 같은 둘).
- 보이스 2 의 breakoutlane 「root」 5 실패는 `.git` 없는 사본 아티팩트(양판 동일). 보이스 3 은 `-trimpath` 가 runtime.Caller 기반 시험을 깨는 것을 실측하고 GOFLAGS 를 비움.
- 보이스 1 관측: 다른 세션의 태그 엔진 `go test` 동시 실행(무관여).
- **codex r2 산출 디렉터리 이름 변경:** `codex-r2/` → `_codex-r2/`(2026-10-04). 그 안의 `.go` 16 개(패키지 breakoutlane · engine 혼재)가 공유 트리의 `go vet ./...` 를 깨고 있었음(`found packages breakoutlane … and engine …` — rtk 요약은 「No issues」 로 숨김, `rtk proxy` 로 실측). Go 는 `_` 시작 디렉터리를 건너뜀 — 파일 이름 · 내용 불변, 이름 변경 뒤 `go vet ./...` rc=0. `codex-output-r2.md` · `_codex-r2/REPORT.md` 안의 `codex-r2/` 경로 언급은 원문 그대로 둠.
- 잔여 비대칭(Y): 철회 경합 시 조정 앞 닫힘 여섯이 낡은 관측값을 실음 — 관측 전용, 브리프 §6 · 응답 로트 FLM 에 기록.
