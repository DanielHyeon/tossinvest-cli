# Function Logic Map: `attestationWatch.eventFor`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-23, base `064f59fd9`.** 2단계 GREEN 뒤 AST·지도를 다시 만든다.

- Source: `internal/app/engine/attestation_watch.go` (314-375)
- AST evidence: `ast.json` — branches **13** / returns **1** / calls **18**, source SHA-256 `29252ace6e7338abe7fa91bf9a7cfe894b8edde2c87657e60e9f54e1e9c7d68d`(파일 단위).
  명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/app/engine/attestation_watch.go --func attestationWatch.eventFor`.
- Risk scan: `risk-pattern-report.md`(일치 0). a166 census 0건. 호출자: `attestationWatch.prepare`(call w.eventFor@298) 1(CodeGraph 1.6.0 + grep).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `v attestationVerdict` | judge 결과 | `judgeAttestation` | — |
| `key` | `stageKey` 결과 | prepare | — |
| `w.critical` | `notifications.enabled` | `newAttestationWatch` L190 | false 면 critical 단계도 normal 타입(B1) |
| `w.identity` | `IssuedAt|ExpiresAt` | prepare L286-288 | outbox Key 접미사(L370) |
| 불변식 | 제목·본문·필드·Key 에 계좌·파일 경로·원문 오류 없음 | engine-safety, `TestA135AlertCarriesNoAccount`·`TestA135UnreadableAlertCarriesNoAccountOrPath` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@316) | `v.critical() && w.critical` | `kind = EventCredentialExpiring`(critical, 정본 등급표 obs/event.go:369), 아니면 `EventCredentialExpiryNotice`(normal) | — | `TestA135WatchAnnouncesEachStageOnce` · `TestA135NotificationsOffNeverRaisesCritical` |
| B2 (if@321) | 만료 시각 있음 | `expires` = KST 문자열(call@322) | — | 날수·만료 시험 |
| B3 (switch@329) | 단계별 문구 | — | — | — |
| B4 (case@330) | notice·imminent | 제목 `자동매매 검증서(attestation) 만료 D-%d`(Sprintf@331), 본문 만료 시각·날수·consequence·remedy(@332) | — | `TestA135WatchAnnouncesEachStageOnce` |
| B5 (case@333) | expired | 제목 「…만료됨」, 본문(@335) | — | `TestA135InvalidStatesAreRemindedDaily` · `TestA135AlertCarriesNoAccount`「expired」 |
| B6 (case@336) | scope_lost | 제목 「자동매매 검증서(attestation)에 필요한 주문 권한이 없다」(L337 리터럴), 본문 누락 endpoint + 「갱신이 감독 증거 없이 쓰였을 수 있다」 + consequence(조건문 「…재시작하면…」) + remedy(@338·Join@339) — **a196 D7 의 편집 지점** | — | `TestA135RenewalResetsAndScopeLossIsImmediate`(본문에 endpoint 이름) |
| B7 (case@340) | incomplete | 고정 문구 | — | `TestA135AlertCarriesNoAccount`「incomplete」 |
| B8 (case@343) | account_mismatch | 고정 문구(remedy 없음) | — | `TestA135AlertCarriesNoAccount`(불일치) |
| B9 (case@346) | 그 밖(unreadable) | 고정 문구 | — | `TestA135UnreadableAlertCarriesNoAccountOrPath` |
| B10 (if@354) | 만료 시각 있음 | `fields["expires_at"]` UTC RFC3339 | — | 누출 시험 |
| B11 (if@357) | notice·imminent | `fields["days_left"]` | — | 누출 시험 |
| B12 (if@360) | 누락 있음 | `fields["missing_endpoints"]` | — | 누출 시험「scope_lost」 |
| B13 (if@363) | ProofAt 있음 | `fields["supervised_proof_at"]` UTC RFC3339 | — | 누출 시험(모든 a135 픽스처가 증거를 실음) |
| (끝) | — | `obs.Event{Type: kind, Key: "attestation-watch:"+key+":"+w.identity, Title, Body, Fields}` — **Key 에 발급 신원(L370), a196 D5 의 편집 지점** | return@367 | `TestA135WatchAnnouncesEachStageOnce`(Key 중복 없음) · `TestA135SameStageOnANewAttestationIsAnnouncedAgain`(신원 다르면 Key 다름) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `v.critical`(call@316) | 거절·임박이면 true | 순수 | AST `attestationverdict.critical` |
| `fmt.Sprintf`·`strings.Join`·`Format` | 문구 | 순수 | AST |
| (하류) `obs.Notifier.Notify` | Type 으로 등급(`SeverityOf`), critical 이면 outbox Key = `e.Key`(notifier.go eventKey :991-993) | outbox 는 같은 Key 를 `RemindAfter`(1h) 안에 `ClaimSettled` 로 삼킨다 | cited-ast `internal-obs--notifier.eventkey` |
| (하류) `obs.Ntfy.Publish` | 제목은 `Title` 헤더에 **UTF-8 원바이트 그대로**(ntfy.go:113-114, RFC 2047 인코딩 없음), 첫 태그 = Type(:116), 본문 = Body + 필드 줄(`notificationFor` :1004-1013) | Go http 는 0x80 이상 바이트를 헤더에 허용(httpguts `ValidHeaderFieldValue` 는 CTL 만 거절, go1.26.5). 서버가 제목을 어떻게 보여 주는지는 저장소 밖(ntfy.go:112 주석은 「서버가 ASCII-fold」라고 적음 — 미검증) | cited-ast `internal-obs--ntfy.publish` + GOROOT 소스 |

## State mutations and fallbacks

- 지역 값만. 수신자 읽기(`w.critical`·`w.identity`)만 하고 쓰지 않는다.
- fallback: 만료 시각 없음 → 「알 수 없음」(B2 거짓).

## Safety conclusion

- Manager 결정 2026-11-23: B3 switch 에 `proof_overdue` case 하나 추가(Q4), 거절 상태 결과 문장은 제목과 본문 첫 문장 둘 다(Q3).
- Safe edit boundary: B4·B6 의 제목·본문(D7), L370 의 Key 접미사(D5), 날수 본문에 기한 종류·시각(spec 「어느 기한을 세는지」). Type 선택(B1)·필드 이름은 바꾸지 않는다(D6·D8).
- High-risk impact: no — 문구·키. 단 critical Key 가 outbox 중복 판정의 입력이므로 Key 를 줄이면 알림을 삼킬 수 있다(변이 「거절 key 에 발급 신원 되돌리기」의 반대 방향 — 누락 집합 빠뜨리기가 그것).
