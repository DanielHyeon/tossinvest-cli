package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"runtime/debug"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/protectionreadiness"
	"github.com/JungHoonGhae/tossinvest-cli/internal/version"
)

const protectionManifestDigestEnv = "TOSSOS_PROTECTION_MANIFEST_SHA256"

func protectionManifestPin(options Options) string {
	getenv := options.Getenv
	if getenv == nil {
		getenv = os.Getenv
	}
	return strings.TrimSpace(getenv(protectionManifestDigestEnv))
}

func productionProtectionDigests() (string, string) {
	info := version.Current()
	buildParts := []string{info.Version, info.Commit, info.Date}
	if build, ok := debug.ReadBuildInfo(); ok {
		buildParts = append(buildParts, build.GoVersion, build.Main.Path, build.Main.Version, build.Main.Sum)
		for _, setting := range build.Settings {
			switch setting.Key {
			case "vcs", "vcs.revision", "vcs.time", "vcs.modified":
				buildParts = append(buildParts, setting.Key, setting.Value)
			}
		}
	}
	return digestProtectionIdentity(buildParts...), digestProtectionIdentity("TossOS", protectionreadiness.ReadinessRelease, "production-provider/v1", "official-conditional-gateway/v1")
}

func productionProtectionAssemblies(buildDigest string) []protectionreadiness.SupervisorAssembly {
	return []protectionreadiness.SupervisorAssembly{
		{Market: protectionreadiness.MarketKR, ComponentDigest: digestProtectionIdentity("protection-supervisor/v1", buildDigest, "KR", "fill-lifecycle-unwired"), Wired: false},
		{Market: protectionreadiness.MarketUS, ComponentDigest: digestProtectionIdentity("protection-supervisor/v1", buildDigest, "US", "fill-lifecycle-unwired"), Wired: false},
	}
}

func digestProtectionIdentity(parts ...string) string {
	hash := sha256.New()
	for _, part := range parts {
		_, _ = hash.Write([]byte{byte(len(part) >> 24), byte(len(part) >> 16), byte(len(part) >> 8), byte(len(part))})
		_, _ = hash.Write([]byte(part))
	}
	return hex.EncodeToString(hash.Sum(nil))
}
