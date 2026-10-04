// Command a112-family-shadow 는 SHADOW 매니페스트의 **정규 바이트**를 낸다(a112 태스크 7.3.1 — 결정 63 v3).
//
// 검증기(strategyshadow.LoadProductionFamilyShadow)는 파일 바이트가 이 빌드의 `json.Marshal` 출력과 한 바이트도 다르지 않기를 요구하므로
// 에디터로 쓴 JSON 은 통과할 수 없다 — 사람은 값을 정하고 바이트는 이 도구가 만든다. 비밀을 갖지 않는다: 신뢰 앵커는 배포가 env 로 핀하는
// SHA-256 하나이고, 사람이 그 digest 를 `TOSSOS_STRATEGY_FAMILY_SHADOW_<MARKET>_MANIFEST_SHA256` 에 핀해야만 shadow 가 선다.
//
// **SHADOW 는 노출을 열지 않는다.** 이 매니페스트는 어느 OFF 레인의 반사실을 투영에 보일지만 정한다 — desired/effective · 활성화 ·
// dispatch 를 바꾸는 필드가 없다.
//
// 쓰는 법(결속 다섯을 어디서 읽는지는 docs/operations.md 의 같은 절):
//
//	go run ./tools/a112-family-shadow \
//	  -market KR -generation 1 \
//	  -route-manifest-digest sha256:… -calibration-digest … \
//	  -calendar-version … -risk-policy-digest sha256:… -build-digest … \
//	  -actor "이름" -approved-at … -issued-at … -expires-at … \
//	  -shadow CONTINUATION,REVERSAL \
//	  -out strategy-family-shadow-KR.json
//
// 파일을 `0400` 으로 **덮어쓰지 않고** 쓰고, 그 SHA-256 을 stdout 에 낸다 — 그 줄이 env 핀에 그대로 들어간다.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyshadow"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "a112-family-shadow:", err)
		os.Exit(1)
	}
}

type options struct {
	market              string
	generation          uint64
	routeManifestDigest string
	calibrationDigest   string
	calendarVersion     string
	riskPolicyDigest    string
	buildDigest         string
	actor               string
	approvedAt          string
	issuedAt            string
	expiresAt           string
	revoked             bool
	shadow              string
	out                 string
}

func run() error {
	var opts options
	flag.StringVar(&opts.market, "market", "", "KR 또는 US")
	flag.Uint64Var(&opts.generation, "generation", 0, "이 shadow 매니페스트의 세대 (0 금지)")
	flag.StringVar(&opts.routeManifestDigest, "route-manifest-digest", "", "이 시장의 서명된 경로 권한 digest")
	flag.StringVar(&opts.calibrationDigest, "calibration-digest", "", "경로 권한이 말하는 보정 digest")
	flag.StringVar(&opts.calendarVersion, "calendar-version", "", "공식 달력 버전")
	flag.StringVar(&opts.riskPolicyDigest, "risk-policy-digest", "", "TOSSOS_RISK_BUCKET_<MARKET>_MANIFEST_SHA256 과 같은 값")
	flag.StringVar(&opts.buildDigest, "build-digest", "", "엔진이 보고하는 BuildDigest")
	flag.StringVar(&opts.actor, "actor", "", "승인한 사람")
	flag.StringVar(&opts.approvedAt, "approved-at", "", "RFC3339, 예 2026-10-05T00:00:00Z")
	flag.StringVar(&opts.issuedAt, "issued-at", "", "RFC3339, approved-at 이후")
	flag.StringVar(&opts.expiresAt, "expires-at", "", "RFC3339, issued-at 로부터 24시간 이내")
	flag.BoolVar(&opts.revoked, "revoked", false, "폐기 표시")
	flag.StringVar(&opts.shadow, "shadow", "", "shadow 로 관측할 가족을 쉼표로. 비면 넷 다 OFF")
	flag.StringVar(&opts.out, "out", "", "쓸 파일 경로")
	flag.Parse()

	document, err := opts.document()
	if err != nil {
		return err
	}
	data, err := strategyshadow.EncodeProductionFamilyShadow(document)
	if err != nil {
		return fmt.Errorf("정규 바이트를 만들지 못했다: %w", err)
	}
	if opts.out == "" {
		return fmt.Errorf("-out 이 필요하다")
	}
	// 0400 으로, 덮어쓰지 않고 쓴다 — 살아 있는 매니페스트를 건드리기 전에 실패하게(활성화 도구와 같은 규칙).
	file, err := os.OpenFile(opts.out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o400)
	if err != nil {
		return fmt.Errorf("매니페스트 파일을 만들지 못했다 — 이미 있으면 덮어쓰지 않는다: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Chmod(opts.out, 0o400); err != nil {
		return err
	}
	digest := sha256.Sum256(data)
	fmt.Printf("wrote %s (%d bytes)\n", opts.out, len(data))
	fmt.Printf("TOSSOS_STRATEGY_FAMILY_SHADOW_%s_MANIFEST_SHA256=sha256:%s\n",
		strings.ToUpper(string(document.Market)), hex.EncodeToString(digest[:]))
	return nil
}

// document 는 문자열 인자를 문서로 옮긴다. 값의 의미는 검사하지 않는다(판정은 적재기 하나) — 옮길 수 없는 문자열만 막는다.
func (opts options) document() (strategyshadow.Document, error) {
	market, err := parseMarket(opts.market)
	if err != nil {
		return strategyshadow.Document{}, err
	}
	approved, err := parseTime("approved-at", opts.approvedAt)
	if err != nil {
		return strategyshadow.Document{}, err
	}
	issued, err := parseTime("issued-at", opts.issuedAt)
	if err != nil {
		return strategyshadow.Document{}, err
	}
	expires, err := parseTime("expires-at", opts.expiresAt)
	if err != nil {
		return strategyshadow.Document{}, err
	}
	families, err := parseFamilies(opts.shadow)
	if err != nil {
		return strategyshadow.Document{}, err
	}
	return strategyshadow.Document{Market: market, Generation: opts.generation,
		RouteManifestDigest: opts.routeManifestDigest, CalibrationDigest: opts.calibrationDigest,
		CalendarVersion: opts.calendarVersion, RiskPolicyDigest: opts.riskPolicyDigest, BuildDigest: opts.buildDigest,
		Actor: opts.actor, ApprovedAt: approved, IssuedAt: issued, ExpiresAt: expires, Revoked: opts.revoked, Shadow: families}, nil
}

func parseMarket(value string) (strategyrouter.Market, error) {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "KR":
		return strategyrouter.MarketKR, nil
	case "US":
		return strategyrouter.MarketUS, nil
	}
	return "", fmt.Errorf("-market 은 KR 또는 US 여야 한다: %q", value)
}

func parseTime(name, value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, fmt.Errorf("-%s 를 RFC3339 로 읽지 못했다: %w", name, err)
	}
	return parsed.UTC(), nil
}

// parseFamilies 는 쉼표 목록을 가족들로 옮긴다 — 모르는 이름은 조용히 빠지지 않게 거절한다.
func parseFamilies(value string) ([]strategyrouter.Family, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}
	families := []strategyrouter.Family{}
	for _, name := range strings.Split(trimmed, ",") {
		family := strategyrouter.Family(strings.ToUpper(strings.TrimSpace(name)))
		if !family.Known() {
			return nil, fmt.Errorf("-shadow 에 모르는 가족이 있다: %q", name)
		}
		families = append(families, family)
	}
	return families, nil
}
