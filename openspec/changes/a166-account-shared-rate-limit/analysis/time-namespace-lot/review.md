# time namespace 로트 1단계 — 리뷰 기록

- 대상: 커밋 `aaad76d7e`(investigation.md · contract-delta.md · overlay · 재현 로그 · AST/census).
- 등급: 동결 계약의 개정(WORKFLOW 「requirement 변경」) + 위험 가중(엔진 폴링 · 보호 가격 조회 보류) → 적대적 Eng 보이스.
- 상한: 2회차. 리뷰가 제안한 새 가드 · 정책 · 범위 확장은 계획에 넣지 않고 잔여 위험 또는 Manager 결정으로 둔다.

## 1회차 (2026-11-25)

보이스 2 — 둘 다 적대 Eng, 읽기 전용, 저장소 파일 생성 0.
- **C** = 다른 계열: `codex exec -s read-only`(codex-cli 0.165.0) `</dev/null`. 판정 「권장안 가 채택 가능」. P0/P1 0, P2 4, P3 1.
- **K** = Claude 독립 서브에이전트(코드 먼저 읽고 판단한 뒤 문서 대조). 판정 「권장안 가 채택 가능」. P0/P1 0, P2 1, P3 5.
  overlay 를 자기 임시 디렉터리에서 재실행해 같은 3개 실패와 execgw · app/engine `A166` ok 를 확인했다고 보고.

| ID | 등급 | 발견 | 처분 |
|---|---|---|---|
| C1 | P2 | 같은 boot UUID 를 보이는 서로 다른 실제 boot(가상화 · 복원)에서 가 는 uptime 을 차감 — §5 표 누락 | 수용(문서). §5 에 「지원 전제 밖 · 가드 없음」 행 추가, §1-4 정정. 지원은 **범위 확장** → 잔여 위험(Manager 결정 6) |
| K2 | P3 | 같은 취지 + 초기 ns inode 는 상수 4026531834 라 호스트 ns 참여자끼리는 지금도 통과 | 수용(문서). 같은 행에 반영 |
| C2 | P2 | `/proc/self` 는 스레드 그룹 리더 기준, A→B→A 중간 교체는 앞뒤 비교로 못 잡음 — 「모든 중간 변경 검증」은 과장 | 수용(문서). §5 표 문구를 「관측된 앞뒤 불일치 거절」로 제한, 원 계약 전제에 기댐을 명시. contract-delta §1 에 기존 지원 전제 유지 문장 추가. 새 가드 없음 |
| C3 | P2 | §4.1-3 의 L531 진입 조건(RateLimited · Poll · hint>0) 과 prune 진입 조건 누락 | 수용. §4.1-3 · 4 정정 |
| C4 · K1 | P2 | execgw 통과 주장의 원문 로그 없음, 「3개뿐」은 선택 실행 범위의 주장 | 수용. `selected-tests-execgw.log`(대조군 · overlay × 무태그 · `tossos_testseams`, 4회 모두 ok) 추가, 문구를 「선택 실행에서 실패 전환된 함수 3개」로 정정 |
| C5 | P3 | AST 함수 수 17 → 18 | 수용. 정정(`ls ast` 18) |
| K3 | P3 | contract-delta 「진단에만 쓰고」 과장 — 생산 코드가 `Namespace` 를 읽는 곳 0 | 수용. 「관측된 앞뒤 불일치 거절과 저장 형식(0 아님 검사 · 호환)에만」으로 정정 |
| K4 | P3 | S5 로그는 B6/L152 어느 쪽이 거절했는지 구분 못 함 | 수용. 「추론」 명시 |
| K5 | P3 | §8 의 커널 전제 공백(setns EUSERS 등) | 부분 수용. 리뷰어 설명을 「미검증 기억」으로 §8 에 기록, 커널 소스 인용은 2단계 필요 시 |
| K6 | P3 | RED 6 을 기존 시험 파일에 두면 FLM 번들이 늘 수 있음 | 수용. RED 6 은 새 시험 파일로 명시 |

FLM 4 번들 계산(생산 1 + 시험 3)은 두 보이스 모두 맞다고 확인. 자리 표 · 실패 순서의 file:line 불일치 0(K), 「코드와
일치」(C, 조건 누락 C3 제외).

## 2회차 (2026-11-25, 마지막)

보이스: C(codex exec, 같은 설정)만 — diff `aaad76d7e..19078e806` 범위. 원문 `review-r2-codex.md`. 판정 「권장안 가 채택 가능」.
처분 반영 확인: C1 · C4 · C5 · K1–K4 · K6 반영됨, C2 · C3 · K5 부분.

| ID | 등급 | 발견 | 처분 |
|---|---|---|---|
| C2-r2 | P2 | contract-delta 시나리오의 「표본 중 namespace가 바뀌면」이 모든 중간 변경 검출을 약속 | 수용. 「표본 앞뒤에서 관측한 namespace·boot ID가 다르면」으로 정정 |
| C3-r2 | P2 | §4.1-3 정정 문장이 「사건 기록됨」과 「state 미기록 · pending 잔존」을 한 사슬로 이어 모순 | 수용. L531 오류 경로와 hint 없음 경로(L545–556 정상 완료)를 분리 |
| K5 | — | 커널 근거 검증 유보 | 유지(2단계 필요 시 소스 인용). 권장안 안전성은 이것에 기대지 않음 |

2회차 상한 도달 — 이후 문서 정정은 리뷰 없이 Manager 검증으로 넘긴다.

## 2단계 구현 리뷰 R3 (2026-11-25, 대상 `13be09fbd..33e4074cf`)

등급: High-risk 취급(결정 1) — gstack review(체크리스트 수동 적용) + 적대 Eng, 다른 계열 1. 상한 2회차 중 1회차만 썼다.
- **C** = codex exec(read-only, `</dev/null`), 코드 적대 리뷰 + 계약 본문 경량 확인(결정 5). 원문 `review-r3-codex.md`. 판정 「머지 가능」, P3 1.
  계약 본문 14문장 모두 `contract-delta.md` 와 「같음」, 새 모순 0.
- **K** = Claude 서브에이전트, gstack review 체크리스트 수동 적용(설치 · 텔레메트리 · specialist 실행 없음). 판정 「머지 가능」, P0–P2 0, P3 3.
  K 는 옛 `bootstamp.go` overlay 로 새 시험이 L52 · L104 에서 RED 임을, 자체 변이 2건(Finish 오류 삼킴 · 같은 boot anchor 무시)이 잡힘을
  직접 돌려 확인했다고 보고(저장소 변경 0, scratch 삭제).

| ID | 등급 | 발견 | 처분 |
|---|---|---|---|
| R3-C1 | P3 | 원장 「CAUGHT 13」은 2차 로그상 12 | 수용. 원장 · tasks 정정(커밋 `33e4074cf` 메시지의 13 은 고칠 수 없어 원장에 집계 오류로 기록). 3차 원장은 14 |
| R3-K1 | P3 | `TestBootRemainingSameProvenanceStillRejectsNamespaceDriftAfterAnchor` 이름이 새 계약(수락)과 반대 | **Manager 결정 대기** — 이름을 바꾸면 gate ⑤ 에서 base 함수 삭제 + 새 함수가 되어 revision=base 번들이 늘어난다. 이번엔 주석으로만 밝힘 |
| R3-K2 | P3 | `RejectsAnchorBefore…` 의 9s 단언은 base 가 anchor 인지 saved 인지 못 가름 | 수용. anchor 를 saved+0.5s 로 두고 9.5s 단언, 변이 M9(같은 boot anchor 무시) 추가 → CAUGHT |
| R3-K3 | P3 | 「교차 ns Finish 뒤 poison 없음」 단언은 옛 코드에서 그 앞(L52)에서 죽어 RED 가 된 적이 없음 | 수용. 변이 M10(`Finish` 가 다른 ns hint 를 만나면 기록 없이 성공) 추가 → L67 「poisoned=true」 로 CAUGHT |
| R3-K4 | 잔여 | 수리는 `Snapshot` 의 offset 0 검사에만 기댄다 — 다른 `BootSource` 구현이 생기면 같은 검사가 필요 | 잔여 위험으로 기록(생산 `BootSource` 는 `SystemBootSource` 하나, `accountbudgetruntime.Open` L95). 새 가드 없음 |

R3 반영(시험 단언 1곳 · 변이 2건 · 문서) 뒤 2회차 리뷰는 돌리지 않았다 — 변경은 시험 단언 강화와 원장 추가뿐이고 선택 시험 ·
3차 원장으로 확인했다(`selected-tests-after-review.log`, `mutation-run-3.log`). Manager 검증으로 넘긴다.

## Manager 결정 (2026-11-25)
위 6건을 Manager 가 확정했다(권장안 가로 2단계 진행). Manager 는 `overlay-option-ga.diff` · 재현 로그 두 개 · `bootstamp.go` ·
`bootstamp_linux.go` 를 직접 대조했다.
1. 위험 등급 = High-risk 취급(변이 원장 · 적대 리뷰 유지).
2. B8 = `Namespace` 항만 빼고 분기 유지. M6 이 살아남으면 앞선 분기(B3 · B4 의 uptime 항)를 AST 분기 id 로 적는다. 구조 시험 신설 안 함.
3. 실제 커널 ns 시험은 기본 스위트에 넣지 않고 `run.sh` 를 2단계 구현 뒤 제품 코드 그대로 재실행해 S1–S6 원문을 남긴다.
4. tasks.md 에 4.1.1 을 4.2 선행으로 추가, 4.2 는 B1 착지와 B2 · B3 처분 뒤 재실측해 체크.
5. 이번 2보이스 기록(codex · Claude, 2회차)을 requirement 변경 리뷰로 인정. 본문 반영 뒤 다른 계열 경량 확인 1회.
6. 잔여 위험 4건(§7 의 1–3 + 보호 가격 조회의 `PollRemaining` 오류 소비 「안 봤음」)을 issues.md 로 옮기고 새 가드 없음.

## 잔여 위험(새 가드 없이 남김 — Manager 결정 6)

1. 같은 boot_id 를 보이는 서로 다른 실제 boot(boot_id 가상화 · bind 위조 · VM 스냅숏 복제) — 신뢰된 procfs · 단일 호스트
   지원 전제 밖. 가 적용 뒤 비초기 ns 참여자에도 열린다(호스트 ns 참여자에는 지금도 열려 있음).
2. 표본 앞뒤 비교로 못 잡는 중간 namespace 교체(A→B→A) · 스레드별 namespace 조작 — 원 계약의 지원 전제 밖(기존과 같음).
3. 손으로 고친 `event.state` — 관리자 비변조 전제(기존과 같음).
4. 보호 가격 조회가 `PollRemaining` 오류를 소비하는 방식 — 안 봤음(issues.md).
5. 다른 `BootSource` 구현이 생기면 `Snapshot` 과 같은 offset 0 · 현재=자식 · 앞뒤 일치 검사가 필요(R3-K4).

## Manager 검증 (2026-11-25) — 통과

- HEAD `8406e552f`, 작업 트리 깨끗. `origin/main..HEAD` 의 생산 Go diff 는 `internal/clock/bootstamp.go` 하나 — `RemainingBootHint` 네 조건의 `Namespace !=` 항 제거와 주석 1줄뿐임을 Manager 가 diff 로 확인. Go 변경 파일 4개(생산 1 · 기존 시험 1 · 새 시험 2), change 밖 비-Go 변경 0.
- Manager 독립 재실행(무 `-trimpath`, `-count=1`): `internal/clock` ok 0.064s · `internal/accountbudget` ok 50.552s · `internal/accountbudgetruntime` ok 23.129s · `tools/a166-budget-state` ok 0.050s · `internal/execgw -run 'A166|EnginePoll'` ok 20.320s(무태그) · ok 20.441s(`tossos_testseams`). `internal/clock -run Boot -v` PASS 20 · FAIL 0, 새 CrossNamespace 시험 6개와 고친 기존 시험 3개 PASS.
- 고친 기존 시험 3개의 diff 를 읽어 「남는 절반은 계속 거절 단언, 바뀌는 절반은 수락 단언」임을 확인. 변이 원장 3차 CAUGHT 14 · SURVIVED 1(M6 — B3 · B4 uptime 항이 먼저 거절)과 `run-after-fix.log` 의 S5 거절(Snapshot ERR 2줄)을 원문으로 확인.
- Manager 가 다시 돌리지 않은 것: 변이 원장, `run.sh`, `make lint`, 전체 시험 · 게이트.

### Manager 결정(검증 시)
1. 시험 이름 `TestBootRemainingSameProvenanceStillRejectsNamespaceDriftAfterAnchor` 는 바꾸지 않고 주석으로 둔다(R3-K1 처분). issues.md 에 잔여로 남긴다.
2. 남은 gbrain 프로세스(pid 1223578)는 Manager 확인 시점에 이미 스스로 끝났고 HEAD 는 `8406e552f` 그대로 — 조치 없음, 기록만.
3. sdd-sync 는 이 로트 반영에 요구하지 않는다. **이 로트 반영 시점의 CodeGraph fingerprint 는 기록되지 않았다 — 종결 전 재sync 필요.**
