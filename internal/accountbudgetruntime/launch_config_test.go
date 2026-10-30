package accountbudgetruntime

import "testing"

func TestLaunchConfigMatrix(t *testing.T) {
	for _, tt := range []struct {
		mode        string
		modeSet     bool
		manifest    string
		manifestSet bool
		want        Mode
		fail        bool
	}{
		{want: ModeOff}, {mode: "off", modeSet: true, want: ModeOff},
		{mode: "required", modeSet: true, manifest: "fixture", manifestSet: true, want: ModeRequired},
		{mode: "", modeSet: true, fail: true}, {mode: "unknown", modeSet: true, fail: true},
		{manifest: "fixture", manifestSet: true, fail: true}, {mode: "off", modeSet: true, manifestSet: true, fail: true},
		{mode: "required", modeSet: true, fail: true}, {mode: "required", modeSet: true, manifestSet: true, fail: true},
	} {
		got, err := ParseLaunchConfig(tt.mode, tt.modeSet, tt.manifest, tt.manifestSet)
		if (err != nil) != tt.fail || (!tt.fail && got.Mode != tt.want) {
			t.Fatalf("mode=%q set=%v manifestSet=%v got=%v err=%v", tt.mode, tt.modeSet, tt.manifestSet, got.Mode, err)
		}
	}
}
