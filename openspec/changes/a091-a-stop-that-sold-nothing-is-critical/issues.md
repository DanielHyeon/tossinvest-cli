# a091 issues — 소비자 조사·낡음 대장·freeze 입력

> 리뷰 M5가 요구한 파일. 아래 「측정은 잰 순간을 달고 다닌다」 — 시점 없는 항목 금지.

## 소비자 조사 (base `ec29dc72`, 2026-08-06 리뷰가 검증 — 재고정 시 재조사, tasks 1.4)

| 소비자 | 당시 사실 |
| --- | --- |
| `CriticalEvents()` 호출자 | 테스트 2개뿐(집합·개수 미고정) — 새 종류 등록이 깨는 시험 없음 |
| 콘솔·httpapi | 이벤트명 필터 없음 |
| `alert_outbox.event_type` | CHECK 없는 자유 문자열(`outbox.go:51`) — 스키마 무변경 성립 |
| rename 규칙 | `execgw ReasonCode` 소유 — 이벤트 종류 추가와 무관 |

**재조사 의무**: 위는 a092 이전 세계다. a092가 기록 입구·배달 실행자·재알림 창을
세웠으므로 재고정 시점의 소비자는 다시 센다(특히 배달 실행자와 episode 관련 신규 소비자).

## 소비자 재조사 (base `b30318d6`, 2026-10-01 — tasks 1.4)

| 소비자 | base 사실 | 새 종류가 깨는가 |
| --- | --- | --- |
| `CriticalEvents()` 호출자 | 시험 셋 — `internal/obs/measurement_test.go:48`(subject `measurement` 금지) · `internal/obs/a074_quarantine_event_test.go:34`(격리 사건 등재) · `cmd/tossctl/a109_the_engine_outlives_its_sibling_endpoints_test.go:293`(강등 사건 비등재). 개수 · 집합 고정 없음 | 아니오 — subject `exit` |
| `EventExitProposalCapped` 참조 | 생산 `exitloop.go:1626` · `:1645-1646`, 시험 `exitloop_test.go:958`(**부분** 캡 — 종류 유지) · `a092_normal_relay_test.go:37` · `:66`(일반 등급 표본 — 여전히 normal) · `a092_r26b_internal_test.go:147` · `:186`(문자열 표본) | 아니오 — 부분 캡 · 익절 0주는 옛 종류 그대로 |
| 콘솔 · httpapi | 이벤트명 필터 없음(`"exit.` 는 설정 키 `exit.common-policy` 뿐) | 아니오 |
| 배달 실행자(a124 `alertdelivery.go`) | 종류별 분기는 없다(`:328`). **그러나 발송만 하지 않는다**(4판 정정 — 2라운드 보이스 B): publisher 부재를 실패 시도로 세고(`:316-326`) 한도에서 진입 게이트 래치 · ENTRY_BLOCKED 승격(`:446-495`), publisher 없으면 사이클마다 `alert_undelivered` 줄 | **예 — 알림 켜진 엔진에서** 새 critical 행이 이 판정에 든다. 알림 꺼진 엔진은 게이트(design D1)로 행이 생기지 않는다 |
| `tossctl engine alerts ack` (`cmd/tossctl/engine_alerts.go:77`, `mutating: true`, `--operator`) | 미전달 행을 사람이 승인 — 래치 해제 조건 | 예 — 알림 켜짐 · 전송 실패에서 새 행마다 사람 승인이 필요(의도된 a092 의미론) |
| `tossctl engine mode-release` (`cmd/tossctl/engine_mode_release.go:31`) | ENTRY_BLOCKED 해제 | 예 — 위 승격 뒤 |
| 콘솔 알림 꺼짐 안내 「critical 알림은 다시 outbox에 쌓이기만 한다」(`internal/console/settings_notifications.go:169`) | 꺼짐 안내 | 아니오 — 게이트로 꺼짐 엔진에서 새 critical 이 생기지 않으므로 안내가 그대로 참 |
| `docs/operations.md` 「이 절차로 오게 되는 알림」(`:514-520` 모양) | critical 종류별 런북 | 예 — 새 절 필요(tasks 6.8) |
| `alert_outbox.event_type` | `TEXT NOT NULL`, CHECK 없음(`outbox.go:51`) | 아니오 — 스키마 무변경 |
| 구조화 로그 `subject` 필드 | `Subject()` 가 `.` 앞(`log.go:196`) | 아니오 — `exit` |

## M1 — 해소 (2026-10-01, design D6)

- `isZeroQuantity` 는 base 에서 **수치 비교**다(`exitloop.go:1871-1878`) — 첫 리뷰의 「정확히 `"0"` 비교」는 낡았다.
- 입력 쪽 불변식의 생산 출처(번들 근거): 포지션 수량 `ConvergeQuantities`(`converge.go:218` — 계좌 값) → exit 루프의 수치 0 건너뛰기
  (`exitloop.go:541`) · `canonicalSnapshotContext` 의 양수 강제 + `RatString` → `ProjectWholeShares` 의 `units.String()` →
  `orderable = projected != "0"` → `record` B11. 하한은 `MaxDecimal` → `CanonicalDecimal` 또는 `zeroFloor` 의 `"0"`.
- 결론: `applyFloor` 입력 `quantity` 는 양의 정수, 0주 반환은 B2 · 끝 두 자리뿐. 「수치 비교로 세우는 안」은 이미 base 의 모양이다.

## 낡음 대장 (2026-09-30 재작성 시점)

- **a089 참조 제거됨** — 불구현 아카이브(2026-09-28 사용자 결정). 재발·접힘·재알림
  의미론은 a092 소유(design D5)
- **`cmd/tossctl/engine_assembly.go:31-35` Publisher 주석** — nil transport 서술.
  2026-09-30 현재 그 줄에 실재함을 확인했다. 리뷰 M5가 stale로 지목 — 알림 배선
  커밋 `e540668f`(2026-08-04) 이후의 배선 현실과 대조해 갱신(구현 로트)
  → **4판(2026-10-01)**: 2라운드 보이스 B 가 대조 — 주석은 이미 참(nil transport → PENDING → 래치 → 승격). 갱신 불요, 대장에서 닫음
- **본문 코드 좌표 전부 base `ec29dc72` 시점** — 특히 `notifier.go`(a092 단위 ②·③이
  편집) 좌표는 이동했다. 재고정 시 FLM 재생성이 정본(tasks 0.3)

## freeze 입력 — 열린 질문

- ~~**M1 — `quantity` 정규형 불변식.**~~ **해소(위 「M1 — 해소」).** ~~(첫 리뷰 시점 서술 — base 에서 거짓) `isZeroQuantity`는 정확히 `"0"` 비교이고
  `""`·파싱 실패도 0 취급, `"0.0"`·`" 0"`은 통과한다(당시 `:1657-1664`). "0주 경로는
  둘"이 성립하려면 `floor.Quantity`가 정규형이어야 한다 — 그 불변식의 생산 출처
  (설정하는 코드)를 FLM Inputs 표에 인용한다. 인용 없으면 0주 판정 자체를
  `isZeroQuantity`가 아니라 수치 비교로 세우는 안을 freeze에 올린다~~
- **H3 후속 — 반복 사건의 운영자 가시성.** 당시 관측: 13회가 outbox 1행으로 접히고
  2회차부터 `MarkAlertDelivered`가 `state=PENDING`에 걸려 `alert_undelivered` ERROR
  12줄. a092 재알림 창(`remindAfter`) 아래에서 이 모양이 어떻게 바뀌는지 5.1 재생으로
  확인하고 여기 기록(tasks 5.2)

## 8/2 원장 재독 (2026-10-01, 읽기 전용 — Manager Q2 승인)

운영 원장 `journal.db`(`mode=ro` · `query_only`)와 `engine.log`, 계좌 필드는 읽는 자리에서 가림. 결과는 design 「8/2 원장 재독」:
13 관측 동안 보유 5 · 한정 항 매도가능 0, 같은 4분에 엔진 밖 매도로 10 → 5 → 2 → 0. 2라운드 보이스 B 의 추론(매도가능 잠금)을 측정이 확인.
**부수 관측**: 운영 `engine.log` 에 계좌 필드 원문 줄 66,009 개(`"account":"<10자리>"`, 가린 줄 73 개) — base 의 계좌 로그 관행이며
사람 결정 큐 「계좌 가림 설계」(a090 D12 · D13) 소관. a091 은 자기가 편집 · 추가하는 줄만 가린다(design D8).
