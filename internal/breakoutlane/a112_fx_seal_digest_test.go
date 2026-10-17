package breakoutlane

// a112 breakout 덮개 B2(Manager 판정 2026-10-01 — 결함, 수리): 역방향(instrument→account, 통화 다름) FX 봉인에서 NewFXSeal 이 호출자 digest 를
// **덮어쓴 뒤** 자기와 비교했다 — 그 방향은 호출자 digest 가 검증되지 않는 공허 검사(fail-open 모양)였다. 수리 뒤: 주어진 모양 그대로의 digest 를
// 먼저 검증하고, 정규화(비율 뒤집기 · 방향 바꾸기)한 값을 다시 봉인한다. 이전에 수락되던 입력(digest 없음 · 틀림 · 봉인 뒤 변조)을 거절하는 쪽 = 보수.

import "testing"

func a112InverseFX() FXSealInput {
	return FXSealInput{AccountCurrency: "KRW", InstrumentCurrency: "USD", Direction: FXInstrumentToAccount, RateNum: 3, RateDen: 2,
		Scale: 6, AsOfMS: 1, FreshUntilMS: 10}
}

func TestAnInverseFXSealVerifiesTheCallersDigest(t *testing.T) {
	sealed := a112InverseFX()
	sealed.Digest = FXSealDigest(sealed)
	canonical := sealed
	canonical.Direction, canonical.RateNum, canonical.RateDen = FXAccountToInstrument, sealed.RateDen, sealed.RateNum
	for name, input := range map[string]FXSealInput{
		"no digest": a112InverseFX(),
		"digest of another seal": func() FXSealInput {
			i := a112InverseFX()
			i.Digest = FXSealDigest(FXSealInput{AccountCurrency: "KRW"})
			return i
		}(),
		"canonical-form digest":      func() FXSealInput { i := a112InverseFX(); i.Digest = FXSealDigest(canonical); return i }(),
		"rate tampered after seal":   func() FXSealInput { i := sealed; i.RateNum = 4; return i }(),
		"window tampered after seal": func() FXSealInput { i := sealed; i.FreshUntilMS = 99; return i }(),
	} {
		if _, err := NewFXSeal(input); err == nil {
			t.Errorf("%s: an inverse FX seal whose digest does not cover its own fields was accepted", name)
		}
	}
	got, err := NewFXSeal(sealed)
	if err != nil {
		t.Fatalf("a correctly sealed inverse FX input was refused: %v", err)
	}
	if got.value.Direction != FXAccountToInstrument || got.value.RateNum != 2 || got.value.RateDen != 3 || got.value.Scale != 6 ||
		got.value.Digest != FXSealDigest(got.value) {
		t.Fatalf("normalized seal=%+v, want account→instrument 2/3 scale 6 re-sealed over its normalized form", got.value)
	}
	// 정규화된 봉인은 다시 검증해도 통과한다(사이징이 fxValid 로 재검증한다).
	if !fxValid(got) {
		t.Fatal("the normalized seal does not re-validate")
	}
}
