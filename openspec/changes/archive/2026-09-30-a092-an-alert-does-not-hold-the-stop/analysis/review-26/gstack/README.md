# 게이트 준비 gstack /review (2026-09-30) — 전문가 4 원문 요약

- 범위: a092 커밋 슬라이스(`721d0338..9f6dc60d`, 제목에 `[a092` — 생산 33파일 · 2532줄, 시험 포함 7431줄). diff 는 rtk 압축 없이 파일로 떠서 전문가에게 넘김.
- 전문가: testing(ac08316b) · maintainability(ad633290) · security(add89b5d) · performance(a5b5ee3a) — 각자 fresh 문맥, 읽기 전용.
- red team / 적대 패스: 새로 돌리지 않음 — 같은 diff 에 26라운드 codex 1차 · 재확인 2회 + Claude 보이스 A/B/C 와 표적 재확인이 그 역할(Manager 수용, `review.md` §24.12).

## 발견(원문 요지)과 처분

| 전문가 | 확신 | 자리 | 요지 | 처분 |
|---|---|---|---|---|
| security | 6 | notifier.go:263 `judge` | 무조건 래치 설명에 승격 실패 원문(`… transition for <계좌>`) | **수리** `15b64676`(W01 · W02) |
| testing | 8 | cmd/tossctl/engine.go:358 | `engine run` 의 모드 제어 기동 실패 강등 경로 시험 없음(a108 하니스에 AccountRef 없음) | 이월 |
| testing | 7 | engine_mode_release_client_unix.go:58 | CLI 디스크립터 검사(토큰 보내기 전) 시험 없음 | **수리**(W06~W10) |
| testing | 7 | modeops.go:54 | 핸들 가드가 Journal 하나로만 재짐 | **수리**(W11~W13) |
| testing | 7 | mode_projection_wiring.go:32 | **M13 이 「닫은 RED」인데 시험 없음** | **수리**(W14) — 거짓 완료 정정 |
| testing | 7 | a092_mode_release_cmd_test.go:91 | 긴 $TMPDIR 에서 sun_path 로 bind 실패(실측) | **수리**(/tmp) |
| testing | 6 | mode_control_transport_unix.go:180 | 415 · 크기 · 뒤따르는 값 거절 시험 없음 | **수리**(W03~W05, W04 는 가드 가림 → 문구 단언) |
| testing | 6 | a092_mode_release_output_test.go:63 | not-pending 이 옛 문구만 금지 | **수리**(양성 단언) |
| maintainability | 8 | mode_control_transport_unix.go:112 | WriteTimeout 주석(「알림 제어와 같은 한도」)과 값(10s 대 5s) 불일치 | **수리**(주석을 값에) |
| maintainability | 8 | mode_control_transport_unix.go:60 · engine_mode_release.go:154 | 알림 제어 서버 · 클라이언트 사본(이미 413/400 · 5s/10s 로 갈림) | 이월(리팩터) |
| maintainability | 7 | alertdelivery.go:357 | 선점 판정이 실행자에 손으로 두 번 더 | 이월(리팩터 — journal 메서드로) |
| maintainability | 7 | notifier.go:310 | 기록 실패 처리 두 사본(동기 claim 쪽은 원문 · `%v`) | 이월 — 동기 쪽은 base 관행(721d0338:275), 불변식 8 (b) 큐 |
| maintainability | 6 | notifier.go:291 | 수동 Unlock 넷(패닉 시 n.mu 잠김) | 이월(리팩터 — defer 헬퍼) |
| maintainability | 7 | notifier.go:835 `Flush` 주석 | 새 n.mu 계약과 모순 | 이월(base 함수 본문 — 문서 로트) |
| maintainability | 7 | auxiliary.go:105 `runAuxiliary` 주석 | 보조 실행자 둘인데 「하나뿐」 | 이월(base 함수 본문 — 문서 로트) |
| maintainability | 5 | retry.go:415 외 3 | 승격 문구 판정 네 자리 손복사 | 이월(리팩터) |
| maintainability | 4 | record_only.go:116 | nil 원장 계약 둘 | 이월(= 보이스 B #6 잔여) |
| performance | 8 | operating_mode.go:645 | `ORDER BY rowid` 로 인덱스가 정렬에 못 쓰임 → temp B-tree(무한 성장) | 이월 — v33 마이그레이션 후보(EXPLAIN 영수증 `analysis/harness/explain-operating-modes-order.txt`) |
| performance | 6 | modeops.go:141 | 통지 확인이 PENDING 전체를 읽음 | 이월(성능 — event_key 점조회) |
