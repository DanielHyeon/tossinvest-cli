# Transport.RoundTrip 사전 FLM / BTM

base b0f7f1d93, 계획 40ac5b116. AST는 변경 전 source SHA 포함. source 손읽기: caller → collector sticky와 URL/proxy 검사 → Budget.Acquire → observed Begin/poll/context → TLS gated first Write → nil response stop/drain/Finish/Release 또는 response Finish → responseOwner Close/cancel로 gate stop/Release. 외부 responseOwner.Close 본문 변경 없음. ErrNotSent는 gate drain 뒤 첫 Write 여부로만 판정. 현재 nil fence 기존 collector·cleanup 분기 그대로 유지.

추가 분기: 유효 fence 선확인(무 fence는 기존), Acquire 후 동일 fence mutex/sticky 재확인 → 동일 Lease withLease·event lock 아래 cursor/generation/pending 확인 → Begin. 실패는 body Close와 정확 Release, Begin 전 Finish 없음. Begin 실패·Finish/429·body close·gate stop·release 오류는 최초 sticky; 성공 Begin 후 단순 cancellation 및 일반 Aux busy는 sticky 아님. response body Close 후 finish 순서는 private wrapper, fence mutex 밖 body Close/gate stop 후 mutex 안 Release·오류 게시. 재검사 실패 시 mutex 놓고 release helper가 재획득한다. 외부 net I/O 중 mutex 보유 금지.

BTM: nil fence 기존 회귀; baseline historical event/gap/overrun 허용, pending Busy/Orphan·dirty·closed/owner 오류 거절; baseline 뒤 event/gap/rollback/overrun 첫 write 0; Begin·Finish·postcommit·body·release 실패 sticky와 동일 fence 새 collector; B 이미 Acquire 대기 중 A 실패 시 B의 Acquire 후 재확인; clean release B 성공; Aux busy/cancel 정상 settlement 후 재시도; 실 TLS token/401 및 crossprocess429 각 시도 첫 write0; response/ErrNotSent 분리; body close·release 한번·race. 통제는 channel barriers, sleep을 경합 증거로 사용하지 않는다. 구조 검사 JSON의 RoundTrip body shape만 재생성·재검토, 기존 AST helper 편집 없음.

새 fence/OS stub은 신규 함수만, 공용 기존 함수는 RoundTrip 외 수정하지 않는다. 기존 hook으로 B 진입 증명이 부족하여 Manager에 private 시험 hook 추가 허가 요청. no-fence 동작·공식 클라이언트·생산 caller/설정 불변. 미실행 전체 tests/gate·운영·실 broker/실계좌. CodeGraph index 부재면 source/AST 대체를 표시한다.
