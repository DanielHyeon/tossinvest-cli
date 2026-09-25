# a114 구현 중 발견 — 분류와 처리

**blocking 0건.**

## ② safe local

### S1. 답 판정이 다른 파일의 문구 셋에 기댄다 (freeze P1-1)

코드가 붙은 모든 rpcError 를 「답」으로 보려면 두 decoder 의 remote failure 문구와 엔진 auth 의 토큰 거절
문구를 읽어야 한다. 정석은 `internal/positionpolicyrpc` 에 타입 있는 `ErrUnauthorized`·원격 실패 타입을 두는
것이고 그 패키지는 이 change 의 파일 표면 밖이다. 문구는 한 곳(상수 셋)에 두고
`TestTheAnswerPhrasesAreTheSourcesOwn` 이 원본 파일의 문자열 리터럴로 고정한다. **후속 후보**: 타입 도입 후 문구 판정 삭제.

### S1a. 사유 없는 remote failure 는 답이 아니다 (post-review P2-1)

decoder 는 코드가 비어 있어도 JSON 이면 같은 문구로 감싼다. 우리 엔진은 거절에 언제나 `err.Error()` 를 싣으므로,
문구 뒤 사유가 비면 탈착으로 본다(`engineRemoteFailure`). 옛 포트에 다른 JSON 서버가 사유를 붙여 답하면 여전히
답으로 읽힌다 — 에페메럴 포트 충돌 + 사유 있는 JSON 이라는 이중 우연이며, S1 의 타입 도입이 닫는다.

### S2. 밀려난 `positionpolicyrpc.Client` 는 닫을 수 없다

그 client 에는 Close 가 없다. wrapper 의 `io.Closer` 분기는 오늘 운영에서 no-op 이고, 유휴 연결은 엔진 서버
`IdleTimeout: 15s` 또는 엔진 사망 시 커널이 닫는다(얼어붙은 엔진이면 그보다 길 수 있다 — freeze P2). a109 가
strategyprojectionrpc 에 한 것(`Client.Close` = `CloseIdleConnections`)을 positionpolicyrpc 에 하는 것이 후속 후보.
⛔ 그 Close 는 **유휴 연결만** 닫아야 한다 — 요청은 잠금 밖에서 옛 client 를 쥐므로 강제 종료는 진행 중인
Apply·격리 해제를 끊어 적용 여부를 모르게 만든다(post-review P2-4, 코드 주석에도 적음).

### S3. 기존 소스 핀 두 개가 console.go 만 읽었다

`TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes` ·
`TestConsolePolicyWiringCannotOpenOrMigrateTheTradingJournal` 은 `positionpolicyrpc.Dial` 문자열을 console.go 에서
찾았다. dial 이 wrapper 파일로 옮겨 갔으므로 두 파일을 함께 읽게 했다(검사 목록 무변경, 금지 검사는 범위가 넓어짐).

## 잔존·후속 후보

### R1. 거절 화면의 「아무것도 변경되지 않았다」 (freeze P2)

`writePositionPolicyError`·`writeQuarantineError` 의 기본 갈래는 transport 실패에도 그 문장을 붙인다. 부착된
자리에서 Apply·격리 해제가 timeout 나면 엔진이 이미 적용했을 수 있어 거짓일 수 있다. internal/console — 표면 밖.

### R2. a081 캐시가 detached 실패를 간격 동안 들고 있다

부착 뒤 첫 캐시 갱신까지 포트폴리오 행이 「관리 여부 불명」으로 남을 수 있다(≤ 캐시 간격). 캐시 설계대로이며 무변경.

### R3. 콘솔 실측의 「부착」 절반은 테스트로만 증명했다 (규칙 13)

엔진 없이 콘솔을 띄워 detached 화면은 실측했다. 진짜 엔진을 띄워 부착 화면을 보는 것은 엔진 기동(계좌·
브로커)이 필요해 사람 승인 대상이다 — not-applicable. 부착은 진짜 `engine.StartPositionPolicyCommandServer` 로
`TestTheConsoleAttachesWhenTheEngineStartsLater`·`…ReattachesAfterTheEngineRestarts` 가 잰다.
