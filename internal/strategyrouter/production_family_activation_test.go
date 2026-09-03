package strategyrouter

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

// 태스크 8.7.1: 서명된 4-가족 활성화만이 레인을 effective ON 으로 만든다.

type familyActivationFixture struct {
	dir     string
	now     time.Time
	public  ed25519.PublicKey
	private ed25519.PrivateKey
}

func newFamilyActivationFixture(t *testing.T) *familyActivationFixture {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return &familyActivationFixture{dir: t.TempDir(), now: time.Date(2026, 9, 3, 1, 0, 0, 0, time.UTC),
		public: public, private: private}
}

// descriptors 는 그 시장의 네 서술자를 생산 표에서 읽어 만든다.
//
// 네 줄을 손으로 적지 않는 이유: 적으면 레인 ID 나 수평선이 바뀔 때 이 fixture 가
// 조용히 옛 값을 들고 "거절이 맞다" 는 초록을 낸다.
func (fixture *familyActivationFixture) descriptors(market Market,
	on map[Family]bool,
) []productionFamilyActivationDescriptor {
	values := []productionFamilyActivationDescriptor{}
	for _, laneID := range orderedLaneIDs(market) {
		table := productionRouteDescriptors(market)[laneID]
		desired, effective := StateOff, StateOff
		if on == nil || on[table.Family] {
			desired, effective = StateOn, StateOn
		}
		values = append(values, productionFamilyActivationDescriptor{Family: table.Family, Horizon: table.Horizon,
			LaneID: laneID, LaneVersion: table.LaneVersion, Desired: desired, Effective: effective})
	}
	return values
}

// orderedLaneIDs 는 골든 서술자 순서(가족 순서)대로 레인 ID 를 준다.
// map 순회 순서에 기대면 같은 시험이 실행마다 다른 바이트를 서명한다.
func orderedLaneIDs(market Market) []string {
	if market == MarketKR {
		return []string{continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID,
			weeklyvaluelane.KRWeeklyLaneID, breakoutlane.KRLaneID}
	}
	return []string{continuationlane.USContinuationLaneID, reversallane.USReversalLaneID,
		weeklyvaluelane.USWeeklyLaneID, breakoutlane.USLaneID}
}

func (fixture *familyActivationFixture) body(market Market) productionFamilyActivationBody {
	return productionFamilyActivationBody{
		SchemaVersion: productionFamilyActivationSchema, Domain: productionFamilyActivationDomain,
		SignatureAlgorithm: productionFamilyActivationAlgorithm, KeyID: "family-activation-key-v1",
		Generation: 7, Market: market,
		RouteManifestDigest: "sha256:" + strings.Repeat("a", 64),
		CalibrationDigest:   "sha256:calibration-" + string(market), CalendarVersion: "calendar-" + string(market),
		RiskBundleDigest: "risk-bundle-" + string(market), BuildDigest: "build-digest-1",
		ProtectionReadyDigest: "sha256:protection-" + string(market),
		Actor:                 "human-approver",
		ApprovedAt:            fixture.now.Add(-2 * time.Hour).Format(time.RFC3339Nano),
		IssuedAt:              fixture.now.Add(-time.Hour).Format(time.RFC3339Nano),
		ExpiresAt:             fixture.now.Add(time.Hour).Format(time.RFC3339Nano),
		Descriptors:           fixture.descriptors(market, nil),
	}
}

func (fixture *familyActivationFixture) config(market Market,
	body productionFamilyActivationBody, data []byte,
) FamilyActivationConfig {
	return FamilyActivationConfig{ConfigDir: fixture.dir, Market: market,
		ManifestDigest: productionRouteDigest(data), TrustedKeyID: "family-activation-key-v1",
		TrustedKey: fixture.public, ObservedAt: fixture.now,
		RouteManifestDigest: body.RouteManifestDigest,
		CalibrationDigest:   body.CalibrationDigest, CalendarVersion: body.CalendarVersion,
		BuildDigest: body.BuildDigest}
}

// write 는 서명해 파일로 쓰고, 그 파일에 맞는 config 를 돌려준다.
func (fixture *familyActivationFixture) write(t *testing.T, market Market,
	body productionFamilyActivationBody,
) FamilyActivationConfig {
	t.Helper()
	canonical, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	manifest := productionFamilyActivationManifest{productionFamilyActivationBody: body,
		Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(fixture.private, canonical))}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	fixture.writeRaw(t, market, data)
	return fixture.config(market, body, data)
}

func (fixture *familyActivationFixture) writeRaw(t *testing.T, market Market, data []byte) {
	t.Helper()
	path := filepath.Join(fixture.dir, ProductionFamilyActivationFileName(market))
	if _, err := os.Stat(path); err == nil {
		if err := os.Chmod(path, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(path, data, 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o400); err != nil {
		t.Fatal(err)
	}
}

// 서명된 활성화가 정확히 자기가 이름 부른 네 레인을 승격한다.
//
// 대조군이 함께 있다: 같은 값에 **없는** 레인(다른 시장의 네 레인)을 물으면 OFF 다.
// 없으면 "모든 질문에 ON 을 돌려주는" 판본도 이 시험을 통과한다.
func TestASignedFourFamilyActivationPromotesExactlyTheLanesItNames(t *testing.T) {
	fixture := newFamilyActivationFixture(t)
	for _, market := range []Market{MarketKR, MarketUS} {
		body := fixture.body(market)
		config := fixture.write(t, market, body)
		activation, err := LoadProductionFamilyActivation(context.Background(), config)
		if err != nil {
			t.Fatalf("%s: %v", market, err)
		}
		if !activation.Verified() || activation.Generation() != 7 || activation.Market() != market {
			t.Fatalf("%s: verified=%v generation=%d market=%s", market,
				activation.Verified(), activation.Generation(), activation.Market())
		}
		for _, laneID := range orderedLaneIDs(market) {
			table := productionRouteDescriptors(market)[laneID]
			if got := activation.Effective(market, table.Family, laneID, table.LaneVersion); got != StateOn {
				t.Errorf("%s/%s effective = %s, want ON", market, table.Family, got)
			}
			if got := activation.Desired(market, table.Family, laneID, table.LaneVersion); got != StateOn {
				t.Errorf("%s/%s desired = %s, want ON", market, table.Family, got)
			}
		}
		other := MarketUS
		if market == MarketUS {
			other = MarketKR
		}
		// 대조군 둘이고 둘이 서로 다른 것을 가른다.
		//
		// 첫째: 다른 시장의 레인을 물으면 OFF. 이것만으로는 시장 결속을 재지
		// 못한다 — 레인 ID 가 시장 접두사를 달고 있어 열쇠가 저절로 안 맞는다
		// (반증 M26 이 그 자리를 살아남았다: 5.1.1 이 만난 "가족 유도와 시장
		// 접두 레인 ID 가 서로를 가려 준다"와 같은 모양).
		//
		// 둘째: **이 시장의 레인 ID 를 다른 시장 이름으로** 물으면 OFF. 열쇠는
		// 맞고 시장만 틀리므로 시장 결속만이 이것을 막을 수 있다.
		for _, laneID := range orderedLaneIDs(other) {
			table := productionRouteDescriptors(other)[laneID]
			if got := activation.Effective(other, table.Family, laneID, table.LaneVersion); got != StateOff {
				t.Errorf("%s activation answered for %s/%s: %s", market, other, table.Family, got)
			}
		}
		for _, laneID := range orderedLaneIDs(market) {
			table := productionRouteDescriptors(market)[laneID]
			if got := activation.Effective(other, table.Family, laneID, table.LaneVersion); got != StateOff {
				t.Errorf("%s activation answered for market %s on its own lane %s: %s",
					market, other, laneID, got)
			}
			if got := activation.Desired(other, table.Family, laneID, table.LaneVersion); got != StateOff {
				t.Errorf("%s activation answered desired for market %s on its own lane %s: %s",
					market, other, laneID, got)
			}
		}
	}
}

// 영값은 아무것도 승격하지 않는다. 이 패키지 밖에서 만들 수 있는 유일한 값이다.
func TestTheZeroActivationIsNotVerifiedAndPromotesNothing(t *testing.T) {
	var activation FamilyActivation
	if activation.Verified() {
		t.Fatal("the zero activation reports itself verified")
	}
	if activation.Generation() != 0 || activation.Market() != "" || !activation.ExpiresAt().IsZero() {
		t.Errorf("the zero activation carries evidence: %d/%s/%v",
			activation.Generation(), activation.Market(), activation.ExpiresAt())
	}
	for _, market := range []Market{MarketKR, MarketUS} {
		for _, laneID := range orderedLaneIDs(market) {
			table := productionRouteDescriptors(market)[laneID]
			if got := activation.Effective(market, table.Family, laneID, table.LaneVersion); got != StateOff {
				t.Errorf("the zero activation promoted %s/%s: %s", market, table.Family, got)
			}
		}
	}
}

// 파일이 없으면 승격이 없다. 그것이 오늘 생산의 상태다 — measured:
// ~/.config/tossctl 에 strategy-* 매니페스트가 하나도 없다.
func TestWithNoActivationFileNothingIsPromoted(t *testing.T) {
	fixture := newFamilyActivationFixture(t)
	body := fixture.body(MarketKR)
	canonical, _ := json.Marshal(body)
	manifest := productionFamilyActivationManifest{productionFamilyActivationBody: body,
		Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(fixture.private, canonical))}
	data, _ := json.Marshal(manifest)
	config := fixture.config(MarketKR, body, data)
	activation, err := LoadProductionFamilyActivation(context.Background(), config)
	if !errors.Is(err, ErrProductionFamilyActivationUnavailable) {
		t.Fatalf("want unavailable, got %v", err)
	}
	if activation.Verified() {
		t.Fatal("a missing file produced a verified activation")
	}
}

// 서술자 집합이 정확히 넷이 아니면 아무것도 승격하지 않는다.
//
// design.md:208 이 이름 부른 네 가지를 각각 만든다: missing(셋), duplicate,
// unknown lane, family drift. 하나의 시험으로 뭉치지 않는 이유는 어느 규칙이
// 사라졌는지를 실패 메시지가 말해야 하기 때문이다.
func TestAnActivationWhoseDescriptorSetIsNotExactlyTheFourPromotesNothing(t *testing.T) {
	table := productionRouteDescriptors(MarketKR)
	cases := map[string]func([]productionFamilyActivationDescriptor) []productionFamilyActivationDescriptor{
		"legacy three families (breakout missing)": func(values []productionFamilyActivationDescriptor) []productionFamilyActivationDescriptor {
			return values[:3]
		},
		"partial three of four with a duplicate": func(values []productionFamilyActivationDescriptor) []productionFamilyActivationDescriptor {
			return []productionFamilyActivationDescriptor{values[0], values[1], values[2], values[2]}
		},
		"an unknown lane id": func(values []productionFamilyActivationDescriptor) []productionFamilyActivationDescriptor {
			values[3].LaneID = "kr_short_something_else_v1"
			return values
		},
		"a family that drifts from the table": func(values []productionFamilyActivationDescriptor) []productionFamilyActivationDescriptor {
			values[0].Family = FamilyReversal
			return values
		},
		"a horizon that drifts from the table": func(values []productionFamilyActivationDescriptor) []productionFamilyActivationDescriptor {
			values[0].Horizon = HorizonWeekly
			return values
		},
		"a lane version that drifts from the table": func(values []productionFamilyActivationDescriptor) []productionFamilyActivationDescriptor {
			values[0].LaneVersion = "v2"
			return values
		},
		"a fifth descriptor": func(values []productionFamilyActivationDescriptor) []productionFamilyActivationDescriptor {
			extra := values[0]
			return append(values, extra)
		},
		"effective ON without desired ON": func(values []productionFamilyActivationDescriptor) []productionFamilyActivationDescriptor {
			values[0].Desired = StateOff
			values[0].Effective = StateOn
			return values
		},
		"an unknown desired state": func(values []productionFamilyActivationDescriptor) []productionFamilyActivationDescriptor {
			values[0].Desired = DesiredState("MAYBE")
			values[0].Effective = StateOff
			return values
		},
	}
	if len(table) != 4 {
		t.Fatalf("the production table has %d KR lanes, so these mutations do not mean what they say", len(table))
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			fixture := newFamilyActivationFixture(t)
			body := fixture.body(MarketKR)
			body.Descriptors = mutate(fixture.descriptors(MarketKR, nil))
			config := fixture.write(t, MarketKR, body)
			activation, err := LoadProductionFamilyActivation(context.Background(), config)
			if !errors.Is(err, ErrProductionFamilyActivationUnavailable) {
				t.Fatalf("want unavailable, got %v", err)
			}
			if activation.Verified() {
				t.Fatal("the activation was verified anyway")
			}
		})
	}
}

// 한 시장의 파일에 **다른 시장의 온전한 매니페스트**가 들어 있으면 거절이다.
//
// 이 시험이 따로 있는 이유는 반증이 가르쳐 준 것이다. `body.Market !=
// config.Market` 를 지운 변이(M1)가 첫 배터리에서 **살아남았다** — 다른 시장의
// *몸통에 KR 서술자를 담은* 기존 사례는 서술자 표 대조가 대신 잡았기 때문이다.
// 즉 시장 결속이 아니라 우연이 지키고 있었다. 여기서는 몸통·서술자·digest 가
// 모두 US 로 온전하고 어긋난 것은 **파일 이름이 말하는 시장** 하나뿐이므로,
// 그 결속만이 이것을 막을 수 있다. 막지 못하면 KR 파일이 US 레인을 켠다.
func TestTheOtherMarketsWholeManifestInThisMarketsFilePromotesNothing(t *testing.T) {
	fixture := newFamilyActivationFixture(t)
	body := fixture.body(MarketUS)
	canonical, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	manifest := productionFamilyActivationManifest{productionFamilyActivationBody: body,
		Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(fixture.private, canonical))}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	// KR 이 읽는 이름으로 쓴다. 서명·digest·결속은 전부 US 몸통에 맞춰 온전하다.
	fixture.writeRaw(t, MarketKR, data)
	config := fixture.config(MarketKR, body, data)
	activation, err := LoadProductionFamilyActivation(context.Background(), config)
	if !errors.Is(err, ErrProductionFamilyActivationUnavailable) {
		t.Fatalf("want unavailable, got %v", err)
	}
	if activation.Verified() {
		t.Fatalf("the KR file promoted a %s activation", activation.Market())
	}
}

// 사람이 서명해 일부만 켠 것은 정당하다. 켠 것만 켜지고 나머지는 OFF 다.
func TestAnActivationPromotesOnlyTheFamiliesItTurnsOn(t *testing.T) {
	fixture := newFamilyActivationFixture(t)
	body := fixture.body(MarketKR)
	body.Descriptors = fixture.descriptors(MarketKR, map[Family]bool{
		FamilyContinuation: true, FamilyWeeklyValue: true})
	config := fixture.write(t, MarketKR, body)
	activation, err := LoadProductionFamilyActivation(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	if !activation.Verified() {
		t.Fatal("a partially-on activation must still be verified — a signed all-OFF file is not the same as no file")
	}
	want := map[Family]DesiredState{FamilyContinuation: StateOn, FamilyReversal: StateOff,
		FamilyWeeklyValue: StateOn, FamilyBreakoutRetest: StateOff}
	for _, laneID := range orderedLaneIDs(MarketKR) {
		entry := productionRouteDescriptors(MarketKR)[laneID]
		got := activation.Effective(MarketKR, entry.Family, laneID, entry.LaneVersion)
		if got != want[entry.Family] {
			t.Errorf("%s effective = %s, want %s", entry.Family, got, want[entry.Family])
		}
	}
}

// 결속과 신뢰 핀이 어긋나면 아무것도 승격하지 않는다.
func TestAnActivationWhoseBindingsDoNotMatchPromotesNothing(t *testing.T) {
	type mutation struct {
		body   func(*productionFamilyActivationBody)
		config func(*FamilyActivationConfig)
		want   error
	}
	cases := map[string]mutation{
		"a calibration digest the caller did not bind": {
			config: func(c *FamilyActivationConfig) { c.CalibrationDigest = "sha256:another-calibration" },
			want:   ErrProductionFamilyActivationUnavailable},
		"a calendar version the caller did not bind": {
			config: func(c *FamilyActivationConfig) { c.CalendarVersion = "calendar-other" },
			want:   ErrProductionFamilyActivationUnavailable},
		"a route manifest digest the caller did not bind": {
			config: func(c *FamilyActivationConfig) {
				c.RouteManifestDigest = "sha256:" + strings.Repeat("b", 64)
			},
			want: ErrProductionFamilyActivationUnavailable},
		"a build digest the caller did not bind": {
			config: func(c *FamilyActivationConfig) { c.BuildDigest = "build-digest-2" },
			want:   ErrProductionFamilyActivationUnavailable},
		"another key id": {
			config: func(c *FamilyActivationConfig) { c.TrustedKeyID = "family-activation-key-v2" },
			want:   ErrProductionFamilyActivationUnavailable},
		"a digest pin that does not match the bytes": {
			config: func(c *FamilyActivationConfig) {
				c.ManifestDigest = "sha256:" +
					"0000000000000000000000000000000000000000000000000000000000000000"
			},
			want: ErrProductionFamilyActivationUnavailable},
		"the other market's file name": {
			config: func(c *FamilyActivationConfig) { c.Market = MarketUS },
			want:   ErrProductionFamilyActivationUnavailable},
		"a schema version this build does not know": {
			body: func(b *productionFamilyActivationBody) { b.SchemaVersion = "strategy-four-family-activation:v0" },
			want: ErrProductionFamilyActivationUnavailable},
		"another signing domain": {
			body: func(b *productionFamilyActivationBody) { b.Domain = "TossOS/other/ed25519/v1" },
			want: ErrProductionFamilyActivationUnavailable},
		"generation zero": {
			body: func(b *productionFamilyActivationBody) { b.Generation = 0 },
			want: ErrProductionFamilyActivationUnavailable},
		"a market the body disagrees with": {
			body: func(b *productionFamilyActivationBody) { b.Market = MarketUS },
			want: ErrProductionFamilyActivationUnavailable},
		"no risk bundle digest for the later binding": {
			body: func(b *productionFamilyActivationBody) { b.RiskBundleDigest = "" },
			want: ErrProductionFamilyActivationUnavailable},
		"no protection-ready digest for the later binding": {
			body: func(b *productionFamilyActivationBody) { b.ProtectionReadyDigest = "" },
			want: ErrProductionFamilyActivationUnavailable},
		"no named approver": {
			body: func(b *productionFamilyActivationBody) { b.Actor = "" },
			want: ErrProductionFamilyActivationUnavailable},
		"revoked by the approver": {
			body: func(b *productionFamilyActivationBody) { b.Revoked = true },
			want: ErrProductionFamilyActivationRevoked},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			fixture := newFamilyActivationFixture(t)
			body := fixture.body(MarketKR)
			if mutate.body != nil {
				mutate.body(&body)
			}
			config := fixture.write(t, MarketKR, body)
			if mutate.config != nil {
				mutate.config(&config)
			}
			activation, err := LoadProductionFamilyActivation(context.Background(), config)
			if !errors.Is(err, mutate.want) {
				t.Fatalf("want %v, got %v", mutate.want, err)
			}
			if activation.Verified() {
				t.Fatal("the activation was verified anyway")
			}
		})
	}
}

// 수명은 세 방향으로 틀릴 수 있고 셋을 따로 잰다.
func TestAnActivationOutsideItsApprovedLifetimePromotesNothing(t *testing.T) {
	cases := map[string]struct {
		mutate func(*familyActivationFixture, *productionFamilyActivationBody)
		want   error
	}{
		"already expired at the observed instant": {
			mutate: func(f *familyActivationFixture, b *productionFamilyActivationBody) {
				b.IssuedAt = f.now.Add(-2 * time.Hour).Format(time.RFC3339Nano)
				b.ExpiresAt = f.now.Add(-time.Minute).Format(time.RFC3339Nano)
			},
			want: ErrProductionFamilyActivationExpired},
		"issued in the future": {
			mutate: func(f *familyActivationFixture, b *productionFamilyActivationBody) {
				b.IssuedAt = f.now.Add(time.Minute).Format(time.RFC3339Nano)
			},
			want: ErrProductionFamilyActivationUnavailable},
		"issued before it was approved": {
			mutate: func(f *familyActivationFixture, b *productionFamilyActivationBody) {
				b.ApprovedAt = f.now.Add(-time.Minute).Format(time.RFC3339Nano)
			},
			want: ErrProductionFamilyActivationUnavailable},
		"a lifetime longer than the approved ceiling": {
			mutate: func(f *familyActivationFixture, b *productionFamilyActivationBody) {
				b.ExpiresAt = f.now.Add(productionFamilyActivationMaximumLife).
					Add(time.Hour + time.Nanosecond).Format(time.RFC3339Nano)
			},
			want: ErrProductionFamilyActivationUnavailable},
		"a timestamp that is not canonical UTC": {
			mutate: func(f *familyActivationFixture, b *productionFamilyActivationBody) {
				b.ExpiresAt = f.now.Add(time.Hour).Format(time.RFC1123)
			},
			want: ErrProductionFamilyActivationUnavailable},
	}
	for name, entry := range cases {
		t.Run(name, func(t *testing.T) {
			fixture := newFamilyActivationFixture(t)
			body := fixture.body(MarketKR)
			entry.mutate(fixture, &body)
			config := fixture.write(t, MarketKR, body)
			activation, err := LoadProductionFamilyActivation(context.Background(), config)
			if !errors.Is(err, entry.want) {
				t.Fatalf("want %v, got %v", entry.want, err)
			}
			if activation.Verified() {
				t.Fatal("the activation was verified anyway")
			}
		})
	}
}

// 다른 키로 서명한 파일은 승격하지 않는다. digest 핀은 맞춰 준다 — 맞지 않으면
// 서명 검사에 닿기 전에 거절되므로 이 시험이 다른 것을 재게 된다.
func TestAnActivationSignedByAnotherKeyPromotesNothing(t *testing.T) {
	fixture := newFamilyActivationFixture(t)
	body := fixture.body(MarketKR)
	_, other, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	manifest := productionFamilyActivationManifest{productionFamilyActivationBody: body,
		Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(other, canonical))}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	fixture.writeRaw(t, MarketKR, data)
	activation, err := LoadProductionFamilyActivation(context.Background(), fixture.config(MarketKR, body, data))
	if !errors.Is(err, ErrProductionFamilyActivationUnavailable) {
		t.Fatalf("want unavailable, got %v", err)
	}
	if activation.Verified() {
		t.Fatal("a foreign signature produced a verified activation")
	}
}

// 바이트가 이 구조체의 정규 직렬화와 다르면 거절이다. 그 한 등식이 unknown
// field, 뒤에 붙은 JSON, 중복 키를 함께 거절한다.
func TestAnActivationWhoseBytesAreNotCanonicalPromotesNothing(t *testing.T) {
	cases := map[string]func([]byte) []byte{
		"an unknown field":     func(data []byte) []byte { return append(data[:len(data)-1], []byte(`,"extra":1}`)...) },
		"trailing JSON":        func(data []byte) []byte { return append(data, []byte("{}")...) },
		"a duplicate key":      func(data []byte) []byte { return append(data[:len(data)-1], []byte(`,"generation":9}`)...) },
		"reordered whitespace": func(data []byte) []byte { return append([]byte(" "), data...) },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			fixture := newFamilyActivationFixture(t)
			body := fixture.body(MarketKR)
			canonical, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			manifest := productionFamilyActivationManifest{productionFamilyActivationBody: body,
				Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(fixture.private, canonical))}
			data, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			data = mutate(data)
			fixture.writeRaw(t, MarketKR, data)
			activation, err := LoadProductionFamilyActivation(context.Background(), fixture.config(MarketKR, body, data))
			if !errors.Is(err, ErrProductionFamilyActivationUnavailable) {
				t.Fatalf("want unavailable, got %v", err)
			}
			if activation.Verified() {
				t.Fatal("the activation was verified anyway")
			}
		})
	}
}

// 검증된 값은 이 단계에서 결속할 수 없었던 두 digest 를 그대로 실어 낸다.
// 뒤 단계가 그 둘을 살아 있는 값과 대조한다.
func TestAVerifiedActivationCarriesTheTwoDigestsTheLaterStageMustBind(t *testing.T) {
	fixture := newFamilyActivationFixture(t)
	body := fixture.body(MarketUS)
	config := fixture.write(t, MarketUS, body)
	activation, err := LoadProductionFamilyActivation(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	if activation.RiskBundleDigest() != body.RiskBundleDigest {
		t.Errorf("risk bundle digest = %q, want %q", activation.RiskBundleDigest(), body.RiskBundleDigest)
	}
	if activation.ProtectionReadyDigest() != body.ProtectionReadyDigest {
		t.Errorf("protection-ready digest = %q, want %q",
			activation.ProtectionReadyDigest(), body.ProtectionReadyDigest)
	}
	if activation.Actor() != body.Actor {
		t.Errorf("actor = %q, want %q", activation.Actor(), body.Actor)
	}
	if want, _ := productionRouteTime(body.ExpiresAt); !activation.ExpiresAt().Equal(want) {
		t.Errorf("expires at = %v, want %v", activation.ExpiresAt(), want)
	}
}

// 취소된 문맥은 파일을 읽지 않는다.
func TestACancelledContextPromotesNothing(t *testing.T) {
	fixture := newFamilyActivationFixture(t)
	config := fixture.write(t, MarketKR, fixture.body(MarketKR))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	activation, err := LoadProductionFamilyActivation(ctx, config)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled, got %v", err)
	}
	if activation.Verified() {
		t.Fatal("a cancelled load produced a verified activation")
	}
}

// 파일 이름이 있는 시장과 서술자 표가 있는 시장은 **같은 집합**이다.
//
// 이 등식이 위 검증에서 `len(want) == 0` 방어를 뺄 수 있는 근거다. 세 번째
// 시장에 파일 이름만 붙이면(표 없이) 그 시장의 매니페스트는 서술자 0 개로
// 완전해지고, 완전성 판정 `len(state) == len(want)` 이 0 == 0 으로 통과한다.
// 그때 이 시험이 실패한다. 이름 목록을 여기 옮겨 적지 않고 enum 을 훑는다 —
// 옮겨 적으면 새 시장이 이 시험 밖에 선다.
func TestEveryMarketWithAnActivationFileNameHasADescriptorTable(t *testing.T) {
	named, tabled := map[Market]bool{}, map[Market]bool{}
	// 후보는 유효 시장 두 개와, 표기 실수로 들어올 수 있는 값 몇 개다.
	for _, market := range []Market{MarketKR, MarketUS, Market(""), Market("kr"), Market("JP"), Market("KRUS")} {
		if ProductionFamilyActivationFileName(market) != "" {
			named[market] = true
		}
		if len(productionRouteDescriptors(market)) != 0 {
			tabled[market] = true
		}
	}
	if len(named) != len(tabled) {
		t.Fatalf("named markets %v but tabled markets %v", named, tabled)
	}
	for market := range named {
		if !tabled[market] {
			t.Errorf("%s has an activation file name but no descriptor table, so an empty"+
				" descriptor list would verify as complete", market)
		}
		if got := len(productionRouteDescriptors(market)); got != 4 {
			t.Errorf("%s has %d descriptors, not the four this manifest exists to bind", market, got)
		}
	}
	if len(named) != 2 {
		t.Fatalf("this test scanned %d named markets, so it is not measuring the pair it claims", len(named))
	}
}
