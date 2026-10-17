package strategyshadow

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// a112 0.5 리뷰 시험#5: 활성화 적재기의 거절 사례(설정 결속 · 몸통 결속 actor · 수명)를 shadow 적재기로 옮겨 적음.
// 옮겨 적은 코드는 양쪽을 다 못 박아야 함 — 앞 판은 shadow 쪽 설정 결속 거절 시험이 0 건이라 `len(fields) != 0 && false` 와
// 수명 · actor 항의 `&& false` 변이가 로트 시험 전부를 통과했음(활성화 쪽 같은 변이는 하위시험 9 개가 잡음).
//
// 복합 결속은 메시지의 「<group>: a, b」 목록이 기대 목록과 **정확히** 같아야 함 — 한 항을 꺼도 파생 거절이 대신 막아 sentinel 은
// 남으므로, 필드 목록이 그 항 자신을 못 박음. 목록은 fixture 시각(승인 −2h · 발급 −1h · 만료 +1h · 상한 24h — 활성화 fixture 와 같음)에서
// 활성화 시험의 목록을 그대로 옮김.
func TestEveryShadowRefusalNamesExactlyItsFields(t *testing.T) {
	name := ProductionFamilyShadowFileName(strategyrouter.MarketKR)
	type shape struct {
		body   func(*productionFamilyShadowBody)
		config func(*Config)
		group  string
		fields []string
	}
	at := func(d time.Duration) string { return shadowAt.Add(d).Format(time.RFC3339Nano) }
	shapes := map[string]shape{
		// 설정 결속(복합) — 한 필드씩, 그리고 여럿 동시.
		"config binding: relative config dir": {config: func(c *Config) { c.ConfigDir = "relative/dir" },
			group: "config binding", fields: []string{"config_dir"}},
		"config binding: zero observed instant": {config: func(c *Config) { c.ObservedAt = time.Time{} },
			group: "config binding", fields: []string{"observed_at"}},
		"config binding: malformed digest pin": {config: func(c *Config) { c.ManifestDigest = "sha256:short" },
			group: "config binding", fields: []string{"manifest_digest"}},
		"config binding: unknown market": {config: func(c *Config) { c.Market = "XX" },
			group: "config binding", fields: []string{"market"}},
		"config binding: empty calibration": {config: func(c *Config) { c.CalibrationDigest = "" },
			group: "config binding", fields: []string{"calibration_digest"}},
		"config binding: malformed route manifest digest": {config: func(c *Config) { c.RouteManifestDigest = "route" },
			group: "config binding", fields: []string{"route_manifest_digest"}},
		"config binding: malformed risk policy digest": {config: func(c *Config) { c.RiskPolicyDigest = "risk" },
			group: "config binding", fields: []string{"risk_policy_digest"}},
		"config binding: empty calendar": {config: func(c *Config) { c.CalendarVersion = "" },
			group: "config binding", fields: []string{"calendar_version"}},
		"config binding: empty build": {config: func(c *Config) { c.BuildDigest = "" },
			group: "config binding", fields: []string{"build_digest"}},
		"config binding: three fields at once": {config: func(c *Config) { c.CalendarVersion, c.BuildDigest, c.RiskPolicyDigest = "", "", "risk" },
			group: "config binding", fields: []string{"build_digest", "calendar_version", "risk_policy_digest"}},
		// 몸통 결속 — actor · generation 과 여럿 동시.
		"body binding: actor": {body: func(b *productionFamilyShadowBody) { b.Actor = "" },
			group: "body binding", fields: []string{"actor"}},
		"body binding: generation zero": {body: func(b *productionFamilyShadowBody) { b.Generation = 0 },
			group: "body binding", fields: []string{"generation"}},
		"body binding: three fields at once": {body: func(b *productionFamilyShadowBody) { b.Generation, b.Actor, b.Domain = 0, "", "other" },
			group: "body binding", fields: []string{"actor", "domain", "generation"}},
		// 수명(복합).
		"lifetime: approved_at not a canonical instant": {body: func(b *productionFamilyShadowBody) { b.ApprovedAt = "yesterday" },
			group: "lifetime", fields: []string{"approved_at"}},
		"lifetime: issued before approved": {body: func(b *productionFamilyShadowBody) { b.ApprovedAt = at(-time.Minute) },
			group: "lifetime", fields: []string{"issued_at before approved_at"}},
		"lifetime: issued in the future": {body: func(b *productionFamilyShadowBody) { b.IssuedAt = at(time.Minute) },
			group: "lifetime", fields: []string{"issued_at after observed_at"}},
		"lifetime: longer than the ceiling": {body: func(b *productionFamilyShadowBody) {
			b.ExpiresAt = at(productionFamilyShadowMaximumLife + time.Hour + time.Nanosecond)
		}, group: "lifetime", fields: []string{"lifetime over maximum"}},
		"lifetime: issued_at not a canonical instant": {body: func(b *productionFamilyShadowBody) { b.IssuedAt = "an hour ago" },
			group: "lifetime", fields: []string{"issued_at", "issued_at before approved_at", "lifetime over maximum"}},
		"lifetime: expires_at not a canonical instant": {body: func(b *productionFamilyShadowBody) { b.ExpiresAt = "in an hour" },
			group: "lifetime", fields: []string{"expires_at", "issued_at not before expires_at"}},
		"lifetime: issued equals expires": {body: func(b *productionFamilyShadowBody) { b.ExpiresAt = b.IssuedAt },
			group: "lifetime", fields: []string{"issued_at not before expires_at"}},
		"lifetime: two faults at once": {body: func(b *productionFamilyShadowBody) {
			b.ApprovedAt = at(-time.Minute)
			b.ExpiresAt = at(productionFamilyShadowMaximumLife + time.Hour + time.Nanosecond)
		}, group: "lifetime", fields: []string{"issued_at before approved_at", "lifetime over maximum"}},
		// 서술자 — 위치 + 필드명(시험#6: 「shadow」 만 찾으면 sentinel 문장 자체가 그 낱말을 품어 공허함).
		"descriptor: unknown shadow state": {body: func(b *productionFamilyShadowBody) { b.Descriptors[0].Shadow = "MAYBE" },
			group: "descriptors[0]", fields: []string{"shadow"}},
	}
	for label, s := range shapes {
		t.Run(label, func(t *testing.T) {
			body, err := shadowDocument().body()
			if err != nil {
				t.Fatal(err)
			}
			body.Descriptors = append([]productionFamilyShadowDescriptor(nil), body.Descriptors...)
			if s.body != nil {
				s.body(&body)
			}
			data, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			dir, pin := install(t, data, name)
			config := shadowConfig(dir, pin)
			if s.config != nil {
				s.config(&config)
			}
			shadow, err := LoadProductionFamilyShadow(context.Background(), config)
			if shadow.Verified() || !errors.Is(err, ErrProductionFamilyShadowUnavailable) {
				t.Fatalf("verified=%v err=%v, want %v", shadow.Verified(), err, ErrProductionFamilyShadowUnavailable)
			}
			got := shadowFieldGroup(err.Error(), s.group)
			want := append([]string(nil), s.fields...)
			sort.Strings(want)
			if strings.Join(got, "|") != strings.Join(want, "|") {
				t.Fatalf("%q fields=%v, want exactly %v (message %q)", s.group, got, want, err.Error())
			}
		})
	}
}

// shadowFieldGroup 은 「<group>: a, b」 꼴 메시지에서 그 목록을 정렬해 돌려줌(group 이 없으면 nil).
func shadowFieldGroup(message, group string) []string {
	index := strings.Index(message, group+": ")
	if index < 0 {
		return nil
	}
	fields := strings.Split(message[index+len(group)+2:], ", ")
	sort.Strings(fields)
	return fields
}
