// Command a112-family-activation 은 4-가족 활성화 매니페스트의 **정규 바이트**를 낸다.
//
// 태스크 8.8.3 [a112 결정 61]. 이 도구가 있어야 하는 이유는 서명이 아니라 정규성이다:
// 검증기는 파일 바이트가 이 빌드의 `json.Marshal` 출력과 **한 바이트도** 다르지 않기를
// 요구하므로(필드 순서·무들여쓰기·끝 개행 없음), 에디터로 쓴 JSON 은 통과할 수 없다.
// 그래서 사람은 값을 정하고 바이트는 이 도구가 만든다.
//
// **비밀을 갖지 않는다.** 앞 판본 설계는 ed25519 개인키로 서명했고 사람이 그것을 뺐다.
// 신뢰 앵커는 배포가 env 로 핀하는 SHA-256 하나다 — 그래서 이 도구의 출력은 그 자체로
// 아무 권한도 아니고, 사람이 그 digest 를 핀해야만 무언가가 켜진다.
//
// 쓰는 법:
//
//	go run ./tools/a112-family-activation \
//	  -market KR -generation 3 \
//	  -route-manifest-digest sha256:… -calibration-digest … \
//	  -calendar-version … -risk-policy-digest sha256:… -build-digest … \
//	  -protection-ready-min-generation 4 \
//	  -actor "이름" -approved-at … -issued-at … -expires-at … \
//	  -on CONTINUATION,REVERSAL,WEEKLY_VALUE,BREAKOUT_RETEST \
//	  -out strategy-family-activation-KR.json
//
// 다섯 결속 값을 어디서 읽는지는 `docs/operations.md` 의 같은 절에 있다.
// 이 도구는 파일을 `0400` 으로 쓰고, 그 SHA-256 을 stdout 에 낸다 — 그 줄이
// env 핀에 그대로 들어간다.
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
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "a112-family-activation:", err)
		os.Exit(1)
	}
}

type options struct {
	market                       string
	generation                   uint64
	routeManifestDigest          string
	calibrationDigest            string
	calendarVersion              string
	riskPolicyDigest             string
	buildDigest                  string
	protectionReadyMinGeneration uint64
	actor                        string
	approvedAt                   string
	issuedAt                     string
	expiresAt                    string
	revoked                      bool
	on                           string
	out                          string
}

func run() error {
	var opts options
	flag.StringVar(&opts.market, "market", "", "KR 또는 US")
	flag.Uint64Var(&opts.generation, "generation", 0, "이 활성화의 세대 (0 금지, 이전보다 커야 한다)")
	flag.StringVar(&opts.routeManifestDigest, "route-manifest-digest", "", "이 시장의 서명된 경로 권한 digest")
	flag.StringVar(&opts.calibrationDigest, "calibration-digest", "", "경로 권한이 말하는 보정 digest")
	flag.StringVar(&opts.calendarVersion, "calendar-version", "", "공식 달력 버전")
	flag.StringVar(&opts.riskPolicyDigest, "risk-policy-digest", "",
		"TOSSOS_RISK_BUCKET_<MARKET>_MANIFEST_SHA256 과 같은 값")
	flag.StringVar(&opts.buildDigest, "build-digest", "", "엔진이 보고하는 BuildDigest")
	flag.Uint64Var(&opts.protectionReadyMinGeneration, "protection-ready-min-generation", 0,
		"승인하는 보호 자세의 하한 세대 (0 금지)")
	flag.StringVar(&opts.actor, "actor", "", "승인한 사람")
	flag.StringVar(&opts.approvedAt, "approved-at", "", "RFC3339, 예 2026-09-04T00:00:00Z")
	flag.StringVar(&opts.issuedAt, "issued-at", "", "RFC3339, approved-at 이후")
	flag.StringVar(&opts.expiresAt, "expires-at", "", "RFC3339, issued-at 로부터 24시간 이내")
	flag.BoolVar(&opts.revoked, "revoked", false, "폐기 표시")
	flag.StringVar(&opts.on, "on", "",
		"켤 가족을 쉼표로. 비면 넷 다 OFF (그것도 정당한 매니페스트다)")
	flag.StringVar(&opts.out, "out", "", "쓸 파일 경로")
	flag.Parse()

	document, err := opts.document()
	if err != nil {
		return err
	}
	data, err := strategyrouter.EncodeProductionFamilyActivation(document)
	if err != nil {
		return fmt.Errorf("정규 바이트를 만들지 못했다: %w", err)
	}
	if opts.out == "" {
		return fmt.Errorf("-out 이 필요하다")
	}
	// 0400 으로, 그리고 **덮어쓰지 않고** 쓴다.
	//
	// 0400 은 검증기가 요구하는 모드다 — 여기서 맞춰 두지 않으면 배포 시각에야
	// 거절을 보게 된다. `O_EXCL` 은 그 모드의 결과다: 0400 파일은 소유자도 다시
	// 열어 쓸 수 없으므로 같은 경로에 두 번째로 쓰면 "permission denied" 가 난다.
	// 매니페스트 수명이 24시간이라 재발급은 매일 있는 일이고, 그때 이 도구가
	// 실패하는 자리를 **살아 있는 매니페스트를 건드리기 전**으로 옮긴다.
	// 재발급 절차(새 경로에 만들고 → 핀을 바꾸고 → 옛 파일을 치운다)는
	// docs/operations.md 의 같은 절에 있다.
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
	// umask 가 비트를 지웠을 수 있으므로 모드를 명시적으로 못 박는다.
	if err := os.Chmod(opts.out, 0o400); err != nil {
		return err
	}
	digest := sha256.Sum256(data)
	fmt.Printf("wrote %s (%d bytes)\n", opts.out, len(data))
	fmt.Printf("TOSSOS_STRATEGY_FAMILY_ACTIVATION_%s_MANIFEST_SHA256=sha256:%s\n",
		strings.ToUpper(string(document.Market)), hex.EncodeToString(digest[:]))
	return nil
}

// document 는 문자열 인자를 문서로 옮긴다.
//
// 값의 **의미**는 검사하지 않는다 — 그 판정은 LoadProductionFamilyActivation 하나가
// 갖고, 여기서 다시 하면 두 판정이 서로의 시험을 통과시킨다. 여기서 막는 것은
// 문자열을 값으로 못 옮기는 경우뿐이다.
func (opts options) document() (strategyrouter.FamilyActivationDocument, error) {
	market, err := parseMarket(opts.market)
	if err != nil {
		return strategyrouter.FamilyActivationDocument{}, err
	}
	approved, err := parseTime("approved-at", opts.approvedAt)
	if err != nil {
		return strategyrouter.FamilyActivationDocument{}, err
	}
	issued, err := parseTime("issued-at", opts.issuedAt)
	if err != nil {
		return strategyrouter.FamilyActivationDocument{}, err
	}
	expires, err := parseTime("expires-at", opts.expiresAt)
	if err != nil {
		return strategyrouter.FamilyActivationDocument{}, err
	}
	families, err := parseFamilies(opts.on)
	if err != nil {
		return strategyrouter.FamilyActivationDocument{}, err
	}
	return strategyrouter.FamilyActivationDocument{
		Market: market, Generation: opts.generation,
		RouteManifestDigest: opts.routeManifestDigest, CalibrationDigest: opts.calibrationDigest,
		CalendarVersion: opts.calendarVersion, RiskPolicyDigest: opts.riskPolicyDigest,
		BuildDigest:                  opts.buildDigest,
		ProtectionReadyMinGeneration: opts.protectionReadyMinGeneration,
		Actor:                        opts.actor,
		ApprovedAt:                   approved, IssuedAt: issued, ExpiresAt: expires,
		Revoked: opts.revoked, On: families,
	}, nil
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

// parseFamilies 는 쉼표 목록을 가족들로 옮긴다.
//
// 모르는 이름을 여기서 거절하는 이유는 조용히 빠지기 때문이다 — 오타 하나가
// "그 가족은 OFF" 로 읽히면 사람은 켰다고 믿고 시스템은 껐다고 믿는다.
func parseFamilies(value string) ([]strategyrouter.Family, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}
	families := []strategyrouter.Family{}
	for _, name := range strings.Split(trimmed, ",") {
		family := strategyrouter.Family(strings.ToUpper(strings.TrimSpace(name)))
		if !family.Known() {
			return nil, fmt.Errorf("-on 에 모르는 가족이 있다: %q", name)
		}
		families = append(families, family)
	}
	return families, nil
}
