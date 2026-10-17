(prompt-common.md 전문을 먼저 읽는다 — 아래는 보이스 B 의 초점.)

## 보이스 B — 스펙 적합성 · 도달 경로 · 배선

- 저자 주장 1 · 4 · 5 · 6 을 코드로 판정하라. 특히 exit 관측 goroutine(`internal/app/engine/exitloop.go` 의 `ObserveOnce` 에서 시작)에서 **알림 경로(`Notify` · `AnnounceOperatingMode` · `Publish`)에 도달하는 모든 호출 경로**를 호출 자리가 아니라 도달 경로로 세고(델타 SHALL), 각각이 기록 전용 인스턴스를 받는지. 공유 부품(Retrier · floor · Guardian · Gateway · Recovery 등)을 거쳐 동기 알림기에 닿는 경로가 남았는지.
- `Context.ExitObserver` 의 값 복사본(Retrier · reconcileFloor)이 원본과 **Announcer 외** 동일한지, 복사가 안전한지(구조체 필드에 뮤텍스 · 가변 상태가 없는지 — `internal/execgw/retry.go`).
- 25.6: a066 완화 통지가 정본의 「세울 자기 사유가 없는 기록자는 입구를 써야 한다」를 만족하는지, 결과 보고(Notified · NotifyError)가 거짓 「통지됨」을 만들 경로가 있는지.
- frozen 24판 tasks §21~§25 가운데 이 세 커밋이 **닫았다고 주장한 항목**(tasks.md 25.3~25.7 의 체크 줄)이 실제로 닫혔는지, 다음 단위로 미룬 것이 정직하게 적혔는지.
