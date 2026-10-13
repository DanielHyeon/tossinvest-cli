package obs

// a133_export_test.go — 시험 전용 접근자(a133). _test.go 이므로 빌드된 바이너리에는 없음.

// RedactForTransportForTest 는 전송 경계의 치환 함수를 시험이 직접 부르게 함(리뷰 F-3 — 문자열 칸 완전성). TESTS ONLY.
func RedactForTransportForTest(msg Notification) Notification { return redactForTransport(msg) }
