package strategyshadow

import (
	"context"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// a112 0.5 리뷰 시험#11: 7.3.1 새 함수의 진입 0 블록 중 거절 · 짧은 회로 갈래를 직접 탐(영수증
// analysis/measurements/lot-0.5-response/branch-coverage-new-functions.txt). 각 갈래가 이 패키지 sentinel 로 거절하는지 셈.
func TestTheShadowEncoderAndLoaderRefuseTheirShortCircuits(t *testing.T) {
	unknownMarket := shadowDocument()
	unknownMarket.Market = "XX"
	unknownFamily := shadowDocument()
	unknownFamily.Shadow = []strategyrouter.Family{"MOMENTUM"}
	for name, document := range map[string]Document{"unknown market": unknownMarket, "unknown family": unknownFamily} {
		if data, err := EncodeProductionFamilyShadow(document); err == nil || data != nil || !errors.Is(err, ErrProductionFamilyShadowUnavailable) {
			t.Errorf("%s: encoder returned %d bytes, err=%v — want no bytes and the Unavailable sentinel", name, len(data), err)
		}
	}
	pin := "sha256:" + "0000000000000000000000000000000000000000000000000000000000000000"
	if _, err := LoadProductionFamilyShadow(nil, shadowConfig(t.TempDir(), pin)); !errors.Is(err, ErrProductionFamilyShadowUnavailable) { //nolint:staticcheck
		t.Errorf("nil ctx: err=%v, want Unavailable", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if shadow, err := LoadProductionFamilyShadow(cancelled, shadowConfig(t.TempDir(), pin)); !errors.Is(err, context.Canceled) || shadow.Verified() {
		t.Errorf("cancelled ctx: verified=%v err=%v, want context.Canceled before any read", shadow.Verified(), err)
	}
	for name, data := range map[string][]byte{"empty": {}, "oversized": make([]byte, productionFamilyShadowMaximumBytes+1),
		"trailing": append(encoded(t, shadowDocument()), []byte("{}")...)} {
		if _, err := decodeProductionFamilyShadow(data); !errors.Is(err, ErrProductionFamilyShadowUnavailable) {
			t.Errorf("decode %s: err=%v, want Unavailable", name, err)
		}
	}
}
