**PASS-with-findings — 1회차 P1 두 건은 정적 검토상 닫혔습니다. 신규 P0·P1은 발견하지 못했습니다. P2 시험 결함 1건이 남습니다.**

기준: `e391565e5 → 2338d468f6eac7f4c5d318b48d16a2ce8dd1af60`. 수리 diff 전체와 현재 `in_doubt_settle.go` 전체, 1회차 지적·처분, 변이 원장·M13 치환식을 확인했습니다.

**[P2] 비대기 시험의 준비 구간은 실패 시한 밖에서 계속 대기할 수 있다**

- **근거:** [a193_in_doubt_settle_test.go:562](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/app/engine/a193_in_doubt_settle_test.go:562), 같은 파일 543–549행·563–566행.
- **왜:** `<-res.entered`에 시한이 없고, 다음 채널 비우기는 생산자가 계속 채우는 동안 `len(cycles) == 0`을 기다립니다. 10초 시한은 두 구간이 끝난 뒤 생성됩니다. 따라서 준비 단계가 멈추면 로컬 실패 시한과 cleanup에 도달하지 못하고 패키지 전체 timeout에 의존합니다. 50ms 성공 조건은 제거됐지만 시험 전체가 유계가 된 것은 아닙니다.
- **최소 제안:** 준비 단계부터 실패 시한을 적용하고, 채널 비우기는 진입 시점의 길이만큼만 수행합니다. 실제 ExitObserver 제출을 측정하지 않는 기존 범위 공백은 처분대로 별도 잔여입니다.

**P1 수리 확인**

- **실패 알림 유실:** [in_doubt_settle.go:159](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/app/engine/in_doubt_settle.go:159), [같은 파일:346](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/app/engine/in_doubt_settle.go:346). 실패 이벤트를 `unsent`에 보존하고 원장 조회 전에 재시도합니다. attempt가 종결돼 조회에서 사라져도 유지됩니다. 성공하면 삭제·래치하며, `handled`가 같은 스윕의 중복 시도를 막습니다.
- **해소 오류의 주문번호 로그:** [in_doubt_settle.go:202](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/internal/app/engine/in_doubt_settle.go:202). 오류 원문 대신 고정 분류만 기록합니다. 기존 지적의 주문번호 포함 오류가 이 로그로 노출되는 경로는 닫혔습니다.

**M13 판정: 원인 서술 타당 — 해당 변이는 시험 공백으로 보지 않습니다.**

[하네스:32](/home/daniel/tossos-work/a193-in-doubt-runtime-resolve/openspec/changes/a193-an-in-doubt-order-is-resolved-while-the-engine-runs/analysis/mutation/harness.py.txt:32)는 `RecordCritical` 직전에 `recorded[key] = true`를 **추가**합니다. 실패하면 원본·변이 모두 `unsent`에 이벤트가 남고, `retryUnsent`는 `recorded`를 확인하지 않습니다. `handled` 역시 두 경우 모두 참입니다. 현재 단일 실행 경로에서 재시도·중복 방지 행동은 같습니다.

단, **성공 뒤 래치를 삭제해도 된다는 뜻은 아닙니다.** 성공 시 `unsent`가 삭제되므로 `recorded`까지 없애면 다음 스윕에서 같은 이벤트를 다시 기록하게 됩니다.

확인한 항목:

- 재시도는 실패가 지속되면 계속되지만 key당 스윕 1회이며, 스윕 종료 후 5초 대기가 있습니다. 즉시 무한 재시도 루프는 없습니다.
- 재시도마다 기록 입구의 로그·원장 접근·실패 시 게이트 차단/승격 시도는 반복됩니다. 성공 뒤에는 해당 재시도가 종료됩니다.
- `Run` 종료 후 같은 호출 흐름에서 `OnStop`이 실행됩니다. 추가된 맵의 병렬 접근 경로는 발견하지 못했습니다.
- 취소된 ctx는 기록 입구 호출 전에 차단됩니다. 검사 뒤 기록 중 취소 경합은 기존 D3 잔여입니다.
- 목록·기록 실패 및 정지 경로에는 원문 기반 로그가 남습니다. 확인한 실제 호출 경로에서 기존 P1과 같은 주문번호 누출은 추가로 입증되지 않았습니다. 모든 오류의 비밀정보 제거를 보장한다는 판정은 아닙니다.
- 수리에 새 주문 mutation·재생 호출·손절 루프의 동기 해소 호출은 없습니다. 공유 자원에 따른 실제 손절 지연은 이번 정적 리뷰로 측정하지 않았습니다.
- 파일 변경·시험·네트워크·운영 접근 0. 기존 `verify-engine-pkg.txt` 수정 상태는 전후 동일했습니다.
- 외부 스킬·memory·그래프 재조회는 경로·네트워크 제한으로 미수행. `impeccable: N/A — backend/documentation-only`, UI 변경 없음. 배포 승인이나 전체 시험 PASS를 뜻하지 않습니다.