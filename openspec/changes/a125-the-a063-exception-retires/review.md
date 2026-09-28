# a125 review

## Proposal-freeze — 2026-09-29 (적대 Eng 보이스 1 · codex 는 Manager 슬롯 대기열)

- 구성: 게이트 판정 변경이므로 적대 Eng 보이스 1(독립 서브에이전트, 읽기 전용). 첫 실행(2026-09-28)은 세션 재시작으로
  handback 없이 끝났고 재실행했다. codex freeze 는 Manager 슬롯 대기열에 있다(a094 r5 다음).
- **판정: freeze 가능 — P0 없음.** P1 다섯은 아래처럼 design · tasks 에 반영했다(구현 전).

| # | 등급 | 발견 | 처분 |
|---|---|---|---|
| 1 | P2 | 구멍 없음. 기록 보유는 a063 하나이고 a063 에서 빌리는 change 는 0 이다. `SDD_BASE_REF` 는 persisted 하나로 좁아져 더 엄해진다. 비-a063 에 기록이 생기면 옛 도구는 fail-closed, 새 도구는 무시한다(D2 승인 사항). | 수용 |
| 2 | P2 | 델타 정합(scenario 48 = 49 − 6 + 2 + 3). RENAMED+MODIFIED 의 archive 적용 순서는 미검증이다. | scratch archive dry-run 으로 확인(아래 「archive dry-run」) |
| 3 | P2 | 남은 참조: `check_analysis.py` docstring 의 자식 프로세스 자리 수 주장(`execution_baseline.validate`) · README:95 · WORKFLOW:241 · README 247-290 · `fixture_git_env.py:3`. Makefile · CI 는 discover 라 무변경이다. `SDD_PYTHON` 은 `tools/sdd/test_sdd_doctor.py` 가 덮으므로 skip 시험(456) 삭제의 손실은 0 이다. | 2.1 · 2.3 에 포함 |
| 4 | **P1** | 재고정 조건 ① 이 오늘은 **성립하지 않는다.** base 번들 3 은 E(=`676bd4b4^`)의 파일 해시를 적는데, 세 함수는 오늘 존재하므로 P 기준으로는 current 여야 한다. 재고정 뒤에는 `676bd4b4` 가 창 밖이고, `validate_target` 이 base 번들 해시를 안 보므로 어떤 게이트도 ① 을 다시 재지 않는다. | tasks 4.2 를 바꿨다: base 3 을 current 로 재추출 → 특례 없는 도구로 **옛 base P 에서** `check_analysis` → 출력에서 9 함수 이름의 missing · stale · revision 오류 0 을 필터 영수증으로 남김 → 그다음 재고정. proposal 문구도 "재추출 후 해당" 으로 고침 |
| 5 | **P1** | 소급 고지가 사라진다. `retrospective-exception` · "missing original analysis remains debt" 를 적는 유일한 기계 기록이 삭제되는 `execution-baseline.json` 이다. | tasks 4.1 에 추가: a063 `review.md` 영수증에 그 두 사실을 명시하고, a063 `tasks.md` 4.0 에 "a125 에서 폐기" 주석 |
| 6 | **P1** | 재고정 시점. a063 의 4.2~4.4(사람 운영, 수 일~수 주)가 끝나기 전에 형제 Go 가 착지하면 창에 다시 요구되고, current 번들 6 의 파일 단위 해시가 낡는다. 둘째 재고정 · 재추출이 거의 확실하다. | **Manager 결정 요청**: (가) a125 에서 재고정 + 프로브, 둘째 재고정을 a063 게이트 직전으로 명기 (나) a125 는 ① 영수증까지, 재고정은 a063 게이트 직전 |
| 7 | P2 | `c727ad12`(S)는 HEAD 조상이 아니고, `676bd4b4` 가 main 사본이다. `0c563c6c`·`b8f31f27` "[a063]" 은 renumber 전 **다른** a063(현 a069) 것이다. 현재 a063 디렉터리는 `47a7f90a` 에서 태어났고 옛 이름이 없다. | 4.2 영수증에 적는다 |
| 8 | **P1** | "같은 픽스처로" 반전은 모듈 삭제 뒤 ImportError(`adoption.draft`). 반전 시험은 리터럴 a063 id 를 써야 id 우회 변이를 잡는다. | 이미 반영: RED 는 모듈 없이 세운 정적 기록 픽스처(`_a063_fixture`, 기록의 `execution_base` = 실제 픽스처 커밋) · 리터럴 `A063`. design D3 에 명기 |
| 9 | **P1** | 빠진 시험 · 변이: 기록이 FIFO · 디렉터리 · 심링크 · 해독 불가여도 일반 판정. 변이는 기록 `execution_base` 를 base 로 읽기 · id 로 착지 우회 · `_recording_refusal` 거절 복원. 시한 시험의 모듈 목록을 정확한 집합으로. | 1.2 보강(비정규 기록 시험) · 3.2 변이 목록에 추가 · 시한 시험은 디렉터리의 `subprocess` 쓰는 비시험 모듈 전부와 대조 |
| 10 | P2 | `effective_base` 는 `main` 창 줄이 쓴다 — 지우면 안 된다. 반전의 "이관 키 없음" 은 `execution_baseline_adoption` · `adoption_source` 둘로 한정한다. | 이미 그렇게 구현(초안). 반전 시험도 그 둘만 단언 |

미검증(보이스 자신이 적음): archive 적용 순서(→ dry-run), 착지 기록으로 a063 창을 고정할 수 있는지(6), 시험 밖 경로가
change 디렉터리 파일을 전수로 읽는지.
