# a074 · Design

## D1 — 사이클 실패는 로그이지 critical 알림이 아니다

`ExitObserver.Run`이 사이클 결과를 읽고 `cycle.Err`를 error 등급 구조화 로그로 남긴다.
알림으로 올리지 않는다.

**왜 알림이 아닌가.** `cycle.Err`는 단일한 의미가 없다. 원장 write 실패, 심볼 하나의
비용 계산 실패, 격리 생성, 판정 기록 실패가 전부 같은 필드에 들어온다. 이것을 critical로
만들면 일시적 SQLite busy 하나가 outbox 행을 만들고, 전송이 안 되면 entry gate를 잠그고,
지속되면 ENTRY_BLOCKED로 강화된다. 즉 **일시적 오류가 계좌를 멈춘다.** `event.go`가
measurement 이벤트에 대해 이미 명시한 판단과 같은 방향이다 — "an event listed there goes
to the durable outbox, and an outbox entry that cannot be delivered latches the entry gate".

알림이 필요한 개별 조건은 이미 자기 이벤트를 갖고 있다(관측 두절, 판정 거부, 청산 지연)
그리고 a074가 하나를 더한다(D2). 로그는 **모든** 실패를 세고, 알림은 이름 붙은 조건만
발송한다.

**왜 `Run`이고 `ObserveOnce`가 아닌가.** 버려지는 지점이 `Run`이다(`_ = o.ObserveOnce(ctx)`).
`ObserveOnce`는 tracer도 호출하며 tracer는 사이클을 직접 읽는다 — 거기에 로그를 넣으면
tracer 실행이 이중 보고를 한다. 그리고 `ObserveOnce`는 분기가 많고 `Run`은 세 개다.

## D2 — 격리 생성은 자기 critical 이벤트를 갖는다

새 이벤트 `exit.snapshot_quarantined`를 `criticalEvents`에 등록한다.

**왜 critical인가.** `event.go`가 exit 그룹의 등급 규칙을 한 문장으로 적어 두었다 —
"does the condition mean a position is *not being protected*?" 격리된 세대는 판정
대상 집합에서 아예 빠지고 손절 평가도 받지 않는다. 답은 예다.

**blast radius가 넓어지지 않는다.** 같은 포지션에 대해 `exit.judgement_refused`가 이미
critical이고 이미 발행된다. 새 이벤트는 그것을 **한 사이클 앞당기고 식별자를 싣는** 것이지
새로운 차단 조건을 만들지 않는다.

**중복은 in-process latch로 막는다.** `EnqueueAlert`는 `event_key`로 dedupe해서 같은 키의
두 번째 행을 만들지 않지만, `Notify`는 그 뒤 `deliver`를 매번 부른다 — Publisher가 nil이면
5초마다 gate block과 error 로그가 반복된다. `o.refused`·`o.delayAlerted`와 같은 모양의
map을 쓴다. 키는 `positionID|generation|version`이므로:

- 같은 격리는 프로세스 수명 동안 한 번만 알린다
- a063으로 해제됐다가 **다시** 격리되면 version이 올라가 새로 알린다
- 재기동하면 다시 한 번 알린다 — `o.refused`와 같은 의미이고, 보호받지 못하는 포지션은
  재기동한 운영자가 다시 봐야 하는 것이 맞다

## D3 — `ambiguous_recovery` 경로는 error에서 되읽는다

`RecordExitJudgementResult`(원장, High-risk)를 편집하지 않는다.

격리를 만드는 세 경로 중 둘(`exitloop.go:494`·`518`)은 관측자가 직접 호출하므로
`ExitSnapshotQuarantine`을 손에 쥐고 있다. 세 번째는 원장 트랜잭션 안에서 만들어지고
관측자는 `ErrExitSnapshotQuarantined`만 받는다.

선택지는 두 가지였다.

| 안 | 편집 대상 | 판단 |
|---|---|---|
| 원장이 격리 행을 error에 실어 반환 | `RecordExitJudgementResult` + `quarantineExitSnapshotTx` | 원장 High-risk 함수 2개 편집 |
| 관측자가 error를 알아보고 활성 격리를 되읽음 | `record` 1개 | **채택** |

되읽기가 정확한 이유: `exit_state.go`는 격리를 쓴 뒤 `tx.Commit()`을 **하고** error를
반환한다(`exit_state.go:503-506`). 그리고 격리의 세대는
`recomputed.Line.PositionGeneration`이고 그 값은 `snapshotContext`가
`m.position.InstanceSeq`로 채운 바로 그 숫자다(`exitloop.go:898`). 즉 관측자가
`ActiveExitSnapshotQuarantine(ctx, p.ID, p.InstanceSeq)`로 되읽으면 방금 커밋된 그 행이다.

읽기 한 번이 실패 경로에만 추가된다. 정상 판정 경로에는 아무것도 추가되지 않는다.

## D4 — 토큰은 설정 파일에 넣지 않는다

```jsonc
"engine": {
  "notifications": {
    "enabled": true,
    "base_url": "https://ntfy.example.internal",
    "topic":    "tossos-alerts"
  }
}
```

토큰은 `TOSSCTL_NTFY_TOKEN`에서만 읽는다. `TOSSCTL_NTFY_TOPIC`이 설정돼 있으면 파일의
topic보다 우선한다.

**왜 topic도 환경변수로 덮을 수 있어야 하는가.** ntfy.sh에서는 **topic 이름 자체가 유일한
접근 제어**다(`ntfy.go` 파일 주석이 그렇게 적고 있다). 그것을 설정 파일에 강제하면
config를 읽을 수 있는 누구나 알림 채널에 쓸 수 있다. 자체 호스팅 + 토큰 구성에서는 topic이
비밀이 아니므로 파일에 두는 편이 diff 가능하고 낫다. 두 운영 형태가 실제로 다르므로 둘 다
허용한다.

**왜 config 패키지가 환경변수를 읽지 않는가.** `internal/config`는 파일을 파싱하는 순수
함수의 집합이고 그 테스트는 프로세스 환경에 의존하지 않는다. 해석(파일 ⊕ 환경)은 조립
지점의 일이다.

## D5 — audit는 값이 아니라 **설정 여부**를 남긴다

```
engine.notifications.enabled           true
engine.notifications.base_url          https://ntfy.example.internal
engine.notifications.topic_configured  true
engine.notifications.token_configured  true
```

topic과 token 자체는 기록하지 않는다. §0.8이 시크릿을 로그에 남기는 것을 금지하고,
ntfy.sh 형태에서 topic은 bearer secret과 같은 역할을 한다. §0.5가 요구하는 것은
"운영 설정이 언제 어떻게 바뀌었는지 추적 가능"이고 그 질문은 값 없이도 답해진다 —
운영자가 topic을 바꾸면 `topic_configured`는 그대로지만 `base_url`과 `enabled`의 변경은
남고, 무엇보다 **알림이 켜졌다는 사실**이 시각·주체와 함께 남는다.

`base_url`은 남긴다. 호스트 주소는 비밀이 아니고, "어느 서버로 보내고 있었나"는 사고
조사에서 실제로 묻는 질문이다.

## D6 — publisher 조립은 `NewContext`에서 한다

`assembleEngine`(cmd/tossctl)이 아니다. 이유는 audit이다. `recordGateSettings`는
`NewContext` 안에서, 어떤 거부보다도 **먼저** 호출된다(`engine.go:418-427`) — "an
operator's settings change is worth recording whether or not the engine then agrees to
start on it". 알림 설정도 같은 성질이므로 같은 호출에 들어가야 하고, 그러면 `cfg`가 있는
곳은 `NewContext`다. publisher 해석을 다른 곳에 두면 설정을 두 번 읽고 두 값이 갈라질 수
있다.

`NewContext`의 편집은 두 줄이다 — `recordGateSettings` 인자와 `buildGateway`의 `publisher`
인자. 나머지 로직은 새 파일의 함수가 갖는다.

## D7 — 잘못된 알림 설정은 엔진을 멈추지 않는다

`enabled: true`인데 topic이 없으면 refuse하고, **publisher 없이 오늘과 같이 기동한다.**
설정 오류를 로그와 audit에 남긴다.

**왜 기동을 거부하지 않는가.** 알림 전송은 보호 경로가 아니다. 오타 하나로 엔진이 뜨지
않으면 그때부터는 손절도 돌지 않는다 — 알림이 없는 상태보다 나쁘다. 그리고 거부하지 않아도
안전 방향은 유지된다: publisher가 없으면 critical 알림은 outbox에 남고 entry gate가
잠기며 지속되면 ENTRY_BLOCKED로 강화된다. 이것은 오늘 이미 벌어지고 있는 일이고
risk-management가 지정한 방향이다.

`Adoption`의 거부 규칙과 같은 모양이다 — 거부된 블록은 **통째로 0으로 만들고**, 왜
거부됐는지를 `Rejected`에 남긴다. 절반만 적용된 알림 설정은 없다.
