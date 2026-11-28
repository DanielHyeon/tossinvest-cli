package attest

// NamesAccount 는 이 attestation 이 accountRef 의 계좌를 가리키는지를 Verify 의 계좌 판정과 **같은 비교**로 답함(a199 D1a).
//
// Verify 는 첫 실패만 돌려주고 만료 검사가 계좌 검사보다 앞이라, 만료된 파일은 계좌가 달라도 ErrExpired 로만 보임.
// 낡은 증명을 청산 전용으로 내릴지 판정하는 쪽은 그래서 만료일 때 계좌를 따로 물어야 하고, 그 물음이 Verify 와 다른 규칙이면
// 두 정본이 됨 — 그래서 Verify 가 쓰는 sameAccount(숫자열 동일 — 하이픈은 같고, 가림 꼴은 다름)를 그대로 부름.
// SameAccountMasked(가림 꼴도 같다고 보는 감독 증거 결속 규칙)는 쓰지 않음.
func (a Attestation) NamesAccount(accountRef string) bool {
	return sameAccount(a.AccountRef, accountRef)
}
