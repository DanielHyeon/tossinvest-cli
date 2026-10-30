package accountbudgetruntime

import (
	"fmt"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
)

// ParseLaunchConfig validates the two launch variables without opening files.
func ParseLaunchConfig(mode string, modeSet bool, manifest string, manifestSet bool) (Config, error) {
	if !modeSet && !manifestSet {
		return Config{Mode: ModeOff}, nil
	}
	switch {
	case modeSet && mode == "off" && !manifestSet:
		return Config{Mode: ModeOff}, nil
	case modeSet && mode == "required" && manifestSet && manifest != "":
		return Config{Mode: ModeRequired, Manifest: manifest}, nil
	default:
		return Config{}, fmt.Errorf("%w: invalid budget launch mode", accountbudget.ErrUnavailable)
	}
}
