# gstack review r1 (2026-11-23) — 브랜치 fix/a196-attestation-deadline-notice, base 064f59fd9(main merge-base)

- 스킬: gstack `review`(Skill 도구), 1회차로 종료. REVIEW_START `37a1e951…`. PR type: mixed(application Go + docs). Diff read: full(Go 4 파일), docs 는 의도 확인용.
- Scope Check: CLEAN — 의도(OpenSpec a196 tasks·design)와 Go diff 일치. Plan completion audit: not run(gstack plan 파일 없음 — 계약은 OpenSpec change).
- Gate edits: 6 listed, 3 read, 0 findings — 셋 다 a135 시험 두 함수의 수정(design §5-11·Manager Q5 결정으로 정당화, diff 안의 design.md 에 기록).
- Banned terms: none. Approval classes: none.
- 전문가 디스패치 6(testing·maintainability·security·performance·api-contract·simplification) + red-team(diff > 200줄) + 네이티브 적대 + Codex 외부 적대.
  security·performance·api-contract·simplification: NO FINDINGS. security 는 브리프의 `':(exclude)*test*'` 가 `attestation_watch.go`(at**test**ation)까지 거른다는 것을 짚고 본문을 경로로 직접 읽음.
- QA(Step 4.7, 기능 표면, 5분·12 probe 한도): probe 2 — 001 `go test -run 'A196|A135|Attestation' ./internal/app/engine ./internal/soak` rc 0, 002 위험 갈래 3 시험 -v rc 0. 증거는 worktree 의 `.gstack/tmp/qa-reports/run-20261123T142253Z/functional/`(무시 파일, 커밋 안 함).

## 발견과 처분

| # | 출처 | 등급 | 발견 | 처분 |
|---|---|---|---|---|
| G1 | maintainability | INFO 9 | `ProofAt` 주석 「판정에는 쓰지 않음」이 코드와 반대 | **AUTO-FIXED** — 주석을 「판정 입력이기도 함(restartDeadline)」으로 |
| G2 | maintainability | INFO 8 | `attestationVerdict.IssuedAt` 을 읽는 곳이 없음(죽은 필드) | 보류(nit) — 무해, 변이 하네스 M03·M04 가 이 필드로 옛 동작을 재현함. 기록만 |
| G3 | maintainability | INFO 7 | 시험 상수 `a196Overdue` 가 `stageProofOverdue` 와 같은 리터럴(RED 때의 흔적) | 보류(nit) — RED 원문 재현성 유지. 기록만 |
| G4 | testing | INFO 7 | 만료가 기한인 통지의 문구를 고정하는 시험 없음 — 갈래 무시 변이 생존 | **FIXED** — `TestA196WithoutSupervisedProofCountsTheExpiry` 에 제목·본문 단언, 변이 N8 CAUGHT(run-3) |
| G5 | 적대 F1 | Medium | 실린 증거가 KR 고정이라 US 에서 `verify run` 을 해도 경보가 안 풀리고, 알림이 어느 시장인지 말하지 않음 | **Manager 결정 대기** — 시장 이름을 필드·본문에 싣기(새 표시) 또는 발급자 수정(a063) |
| G6 | 적대 F2 | Low | 기본보다 긴 `--validity` 로 손수 발급한 파일은 기한이 실제보다 늦음 — 주석의 「놓치는 쪽 아님」 이 그 경로에서 거짓 | **AUTO-FIXED(주석)** — restartDeadline 주석에 예외 명시, design §6 잔여 추가. 처리 여부는 Manager 결정 대기 |
| G7 | 적대 F3 | Low | `IssuedAt` 이 0 인 파일은 Sub 포화로 감독 증거 기한이 조용히 꺼짐(결과는 만료 기준 — 문서의 폴백과 같음) | **Manager 결정 대기**(새 가드) — 발급자 파일에는 해당 없음 |
| G8 | 적대 F4 | Low | 기한 지남 + 갱신 실패면 만료 D-N 단계가 안 보임(매일 critical 은 나감) | 기록만 |
| G9 | 적대 F5 | Low | `SupervisedBy` 없이 POST 를 가진 옛 파일은 사전 통지 없음(이번 변경 전부터) | 기록만 |
| G10 | red-team · Codex 2 | INFO 8 / Medium(Blocks: yes) | 할 일 「`tossctl verify run`」 은 엔진과 같은 journal 잠금(enginelock)을 잡아 **엔진이 도는 동안 실행 불가**(cmd/tossctl/verify.go:440 · engine.go:209 — 팀메이트 확인). 따라서 「지금 실행 중인 엔진은 계속 돈다」+「verify run」 은 엔진 정지(손절 없음)를 숨긴 지시. 기존 검증 기록이 있으면 맨 `verify run` 은 실패하고 `--resume` 은 끝난 측정을 다시 재지 않음(Codex 주장, 팀메이트 미확인) | **Manager 결정 대기(우선)** — 할 일 문구를 실행 가능한 절차(두 시장 닫힌 창에서 엔진 정지 → 감독 검증(재측정) → `soak attest` 즉시 → 재기동)로 바꿀지. D7 문구 결정 영역 |
| G11 | red-team | INFO 7 | 콘솔 attestation 패널은 여전히 만료만 봄 — 알림은 D-2·지남인데 콘솔은 「30일 남음」 | **Manager 결정 대기**(범위 밖 — 콘솔) |
| G12 | Codex 1 | High(Blocks: yes) | 거절 → healthy → 같은 날 같은 거절 재발 시, 옛 거절 알림을 운영자가 **확인(ack)** 했거나 첫 전달 뒤 1시간 안이면 outbox 가 같은 key 를 `ClaimSettled` 로 삼키고(journal/outbox.go:382-410 — 팀메이트 확인: 확인·전달 시각 중 늦은 것부터 RemindAfter), `Notify` 가 nil 이라 감시가 「보냄」으로 기록 → **그날 남은 시간 동안 침묵**. base 는 key 에 발급 신원이 있어 새 행이었으므로 a196 D5 가 만든 회귀. 가짜 알림기 시험(P2-1)은 이것을 못 잰다 | **Manager 결정 대기(우선)** — key 에 「회복 뒤 재발」을 가르는 값(예: 이 거절 에피소드의 첫 관측 시각)을 넣을지. D5 신원 결정 영역 |

## 남은 것

- 미실행: 전체 스위트·gate·race. Codex 구조 리뷰(200줄+ 대상)는 1회차 상한·Manager 지시로 돌리지 않음(외부 적대 패스만).
