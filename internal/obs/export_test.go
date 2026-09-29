package obs

// 시험 전용 접근자 — _test.go 이므로 빌드된 바이너리에는 없음.

// SetDeliveryHookForTest 는 동기 발송 경로의 단계 훅을 꽂음(a092 단위 ③). 단계 이름은 "evidence:<자리>" (근거 확정 직후,
// 해제 세대 읽기 전)와 "epoch:<자리>" (세대 읽기 직후, 적용 전)이고 자리는 unrecorded(:484) · vanished(:520) · exhausted(:571).
// 시험이 운영자 해제를 두 순간 사이에 결정적으로 끼워 넣는 데 씀. TESTS ONLY.
func SetDeliveryHookForTest(n *Notifier, f func(stage string)) { n.deliveryHook = f }
