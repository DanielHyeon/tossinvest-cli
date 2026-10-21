# a163 issues

## I1 — `stableObservationID` 는 `ExitObserver` 메서드가 아니다 (safe local · 2026-10-22 구현 로트)

design D7 1 · tasks 1.2 는 「수신자 `ExitObserver` 의 … `stableObservationID` … 열 개」 라 적었다. HEAD(`exitloop.go:1202`)에서
`stableObservationID` 는 수신자 없는 함수다(`render-hash-base.txt` 는 이름 · 파일만 적었고 값은 그 자유 함수의 렌더). 렌더 해시 시험
(`a163_judgement_pin_test.go`)은 아홉 메서드는 수신자 `ExitObserver`, `stableObservationID` 는 수신자 없음으로 고른다. 기대값 열 개는
base 와 같다(편집 전 GREEN 실측). 설계 의도(판정 함수 열 개 본문 불변) 그대로 — 문구 정정만.

## I2 — `SourceTimestamp` 이름은 편집 전에도 `internal/official` 에 넷 있다 (safe local · Manager 확인 요청)

design D7 2 · tasks 1.2 는 census 가 「편집 전 출현 0 = 허용 밖 0」 이라 적었다. 실측(편집 전, HEAD `1eab9b931`):
`internal/official/strict_quote_reads.go` 의 a112 L1c 엄격 리더 — `type StrictLastPrice` 필드 1 · `type StrictTopOfBook` 필드 1 ·
`Client.StrictLastPrice` 리터럴 1 · `Client.StrictOrderbookTop` 리터럴 1. 다른 타입(`StrictLastPrice`·`StrictTopOfBook`)의 같은 이름
필드이고 생산 판정 소비자는 0(`officialbars` 만 읽음 — proposal 비목표 「a112 잔여 (k)」).

처분(구현 로트): census 의 허용 자리에 이 넷을 **정확한 수**로 더했다 — 새 출현(판정 경로로의 누출)은 여전히 실패한다. census 를
타입 정보(go/types)로 `domain.Quote`·`observedQuote` 필드만 세게 바꾸는 안은 고르지 않았다(설계는 식별자 census 를 정했고, 이름 기준이
더 넓게 막는다). 설계 의도를 넓히거나 좁히지 않는 판단이나, design 의 사실 진술과 다르므로 Manager 확인을 요청한다.

## I3 — 편집 대상 기존 함수가 tasks 1.1 목록보다 다섯 많다 (safe local)

design D8 이 정한 `recordGateSettings` 끝 인자 추가는 호출자 `engine.NewContext`(`engine.go:472`)와 시험 함수 넷
(`a074_notification_audit_test.go` 의 `TestNotificationSettingsAreAudited` · `TestTheAuditTrailCarriesNoNotificationSecret` ·
`TestARefusedNotificationBlockIsAudited` · `TestExistingAuditEntriesKeepTheirOrder`)의 본문을 바꾼다. tasks 1.1 은 이 다섯을 적지 않았으나
게이트 ⑤(`check_analysis.py`)는 바뀐 기존 함수를 직접 계산하므로 다섯 모두 번들을 만들었다(`analysis/function-logic/`). 편집은 각각
호출 인자 하나(생산 `cfg.Engine.ExitSourceAgeReport`, 시험 `false`)이고 분기 · 반환 수 불변.
