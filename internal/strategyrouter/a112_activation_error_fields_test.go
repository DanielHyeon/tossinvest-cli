package strategyrouter

// a112 태스크 8.8.4 항목 1(Manager 판정 2026-10-04, Q-B1 = (c)): 4-가족 활성화 적재의 거절이 **어느 필드 때문인지**를 말한다. 앞 판은 약 서른
// 원인이 맨 sentinel 하나로 뭉쳤다 — 운영자는 만료 · 폐기 · 결속 불일치 · 파일 없음을 구별할 수 없었다.
//   - 단일 조건 갈래: `%w: <field>`.
//   - 복합 결속(설정 결속 · 몸통 결속 · 수명): 분기는 하나 그대로, 그 안에서 필드별 비교를 모아 **불일치 필드 전부**를 한 메시지에 싣는다.
//   - 파일 읽기 결함과 digest 불일치는 다른 종류다 — 읽기 결함은 읽기 함수의 오류를 `%w` 사슬에 보존하고, 불일치만 필드명(manifest_digest)으로.
//   - 모든 갈래에서 sentinel 동일성(errors.Is)은 그대로다 — 엔진의 판별(미선언만 기존 경로)이 그것에 기댄다.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// a112FieldGroup 은 "<group>: a, b" 꼴 메시지에서 그 목록을 정렬해 돌려준다(group 이 없으면 nil).
func a112FieldGroup(message, group string) []string {
	index := strings.Index(message, group+": ")
	if index < 0 {
		return nil
	}
	fields := strings.Split(message[index+len(group)+2:], ", ")
	sort.Strings(fields)
	return fields
}

func TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel(t *testing.T) {
	type shape struct {
		body    func(*familyActivationFixture, *productionFamilyActivationBody)
		config  func(*FamilyActivationConfig)
		file    func(*testing.T, *familyActivationFixture)
		want    error
		group   string   // 복합 결속이면 그 묶음 이름 — 목록이 fields 와 정확히 같아야 한다
		fields  []string // group 이 없으면 메시지에 들어 있어야 할 낱말
		absent  []string // 메시지에 없어야 할 낱말(결함/불일치 구별)
		readErr bool     // 파일 읽기 결함의 원래 오류(공유 읽기 함수의 ErrProductionRouteUnavailable)가 사슬에 남아야 하는가
	}
	shapes := map[string]shape{
		"undeclared pin": {config: func(c *FamilyActivationConfig) { c.ManifestDigest = "" },
			want: ErrProductionFamilyActivationUndeclared, fields: []string{"manifest_digest"}},
		// 설정 결속(복합) — 한 필드씩, 그리고 여럿 동시.
		"config binding: relative config dir": {config: func(c *FamilyActivationConfig) { c.ConfigDir = "relative/dir" },
			want: ErrProductionFamilyActivationUnavailable, group: "config binding", fields: []string{"config_dir"}},
		"config binding: zero observed instant": {config: func(c *FamilyActivationConfig) { c.ObservedAt = time.Time{} },
			want: ErrProductionFamilyActivationUnavailable, group: "config binding", fields: []string{"observed_at"}},
		"config binding: malformed digest pin": {config: func(c *FamilyActivationConfig) { c.ManifestDigest = "sha256:short" },
			want: ErrProductionFamilyActivationUnavailable, group: "config binding", fields: []string{"manifest_digest"}},
		"config binding: unknown market": {config: func(c *FamilyActivationConfig) { c.Market = "XX" },
			want: ErrProductionFamilyActivationUnavailable, group: "config binding", fields: []string{"market"}},
		"config binding: empty calibration": {config: func(c *FamilyActivationConfig) { c.CalibrationDigest = "" },
			want: ErrProductionFamilyActivationUnavailable, group: "config binding", fields: []string{"calibration_digest"}},
		"config binding: malformed route manifest digest": {config: func(c *FamilyActivationConfig) { c.RouteManifestDigest = "route" },
			want: ErrProductionFamilyActivationUnavailable, group: "config binding", fields: []string{"route_manifest_digest"}},
		"config binding: malformed risk policy digest": {config: func(c *FamilyActivationConfig) { c.RiskPolicyDigest = "risk" },
			want: ErrProductionFamilyActivationUnavailable, group: "config binding", fields: []string{"risk_policy_digest"}},
		"config binding: empty calendar": {config: func(c *FamilyActivationConfig) { c.CalendarVersion = "" },
			want: ErrProductionFamilyActivationUnavailable, group: "config binding", fields: []string{"calendar_version"}},
		"config binding: empty build": {config: func(c *FamilyActivationConfig) { c.BuildDigest = "" },
			want: ErrProductionFamilyActivationUnavailable, group: "config binding", fields: []string{"build_digest"}},
		"config binding: three fields at once": {config: func(c *FamilyActivationConfig) {
			c.CalendarVersion, c.BuildDigest, c.RiskPolicyDigest = "", "", "risk"
		}, want: ErrProductionFamilyActivationUnavailable, group: "config binding",
			fields: []string{"build_digest", "calendar_version", "risk_policy_digest"}},
		// 파일 읽기 결함(원래 오류 보존) vs digest 불일치(필드명) — 다른 종류.
		"manifest file missing": {file: func(t *testing.T, f *familyActivationFixture) {
			if err := os.Remove(filepath.Join(f.dir, ProductionFamilyActivationFileName(MarketKR))); err != nil {
				t.Fatal(err)
			}
		}, want: ErrProductionFamilyActivationUnavailable, fields: []string{"manifest file"}, absent: []string{"manifest_digest"}, readErr: true},
		"pin does not match the file bytes": {config: func(c *FamilyActivationConfig) { c.ManifestDigest = "sha256:" + strings.Repeat("0", 64) },
			want: ErrProductionFamilyActivationUnavailable, fields: []string{"manifest_digest"}, absent: []string{"manifest file"}},
		// 바이트 형식.
		"bytes not canonical": {file: func(t *testing.T, f *familyActivationFixture) {
			a112RewriteManifest(t, f, func(data []byte) []byte {
				var out bytes.Buffer
				if err := json.Indent(&out, data, "", "  "); err != nil {
					t.Fatal(err)
				}
				return out.Bytes()
			})
		}, want: ErrProductionFamilyActivationUnavailable, fields: []string{"canonical"}},
		"trailing data after the document": {file: func(t *testing.T, f *familyActivationFixture) {
			a112RewriteManifest(t, f, func(data []byte) []byte { return append(append([]byte(nil), data...), []byte("{}")...) })
		}, want: ErrProductionFamilyActivationUnavailable, fields: []string{"trailing"}},
		"an unknown field": {file: func(t *testing.T, f *familyActivationFixture) {
			a112RewriteManifest(t, f, func(data []byte) []byte {
				return append([]byte(`{"surprise":1,`), data[1:]...)
			})
		}, want: ErrProductionFamilyActivationUnavailable, fields: []string{"manifest json", "surprise"}},
		"revoked": {body: func(_ *familyActivationFixture, b *productionFamilyActivationBody) { b.Revoked = true },
			want: ErrProductionFamilyActivationRevoked, fields: []string{"revoked=true"}},
		// 몸통 결속(복합).
		"body binding: schema": {body: func(_ *familyActivationFixture, b *productionFamilyActivationBody) { b.SchemaVersion = "old" },
			want: ErrProductionFamilyActivationUnavailable, group: "body binding", fields: []string{"schema_version"}},
		"body binding: domain": {body: func(_ *familyActivationFixture, b *productionFamilyActivationBody) { b.Domain = "other" },
			want: ErrProductionFamilyActivationUnavailable, group: "body binding", fields: []string{"domain"}},
		"body binding: generation zero": {body: func(_ *familyActivationFixture, b *productionFamilyActivationBody) { b.Generation = 0 },
			want: ErrProductionFamilyActivationUnavailable, group: "body binding", fields: []string{"generation"}},
		"body binding: market": {body: func(_ *familyActivationFixture, b *productionFamilyActivationBody) { b.Market = MarketUS },
			want: ErrProductionFamilyActivationUnavailable, group: "body binding", fields: []string{"market"}},
		"body binding: calibration": {config: func(c *FamilyActivationConfig) { c.CalibrationDigest = "sha256:other" },
			want: ErrProductionFamilyActivationUnavailable, group: "body binding", fields: []string{"calibration_digest"}},
		"body binding: route manifest": {config: func(c *FamilyActivationConfig) { c.RouteManifestDigest = "sha256:" + strings.Repeat("b", 64) },
			want: ErrProductionFamilyActivationUnavailable, group: "body binding", fields: []string{"route_manifest_digest"}},
		"body binding: calendar": {config: func(c *FamilyActivationConfig) { c.CalendarVersion = "calendar-other" },
			want: ErrProductionFamilyActivationUnavailable, group: "body binding", fields: []string{"calendar_version"}},
		"body binding: build": {config: func(c *FamilyActivationConfig) { c.BuildDigest = "build-other" },
			want: ErrProductionFamilyActivationUnavailable, group: "body binding", fields: []string{"build_digest"}},
		"body binding: risk policy": {config: func(c *FamilyActivationConfig) { c.RiskPolicyDigest = "sha256:" + strings.Repeat("c", 64) },
			want: ErrProductionFamilyActivationUnavailable, group: "body binding", fields: []string{"risk_policy_digest"}},
		"body binding: protection floor": {body: func(_ *familyActivationFixture, b *productionFamilyActivationBody) {
			b.ProtectionReadyMinGeneration = 0
		},
			want: ErrProductionFamilyActivationUnavailable, group: "body binding", fields: []string{"protection_ready_min_generation"}},
		"body binding: actor": {body: func(_ *familyActivationFixture, b *productionFamilyActivationBody) { b.Actor = "" },
			want: ErrProductionFamilyActivationUnavailable, group: "body binding", fields: []string{"actor"}},
		"body binding: three fields at once": {body: func(_ *familyActivationFixture, b *productionFamilyActivationBody) {
			b.Generation, b.Actor, b.Domain = 0, "", "other"
		}, want: ErrProductionFamilyActivationUnavailable, group: "body binding", fields: []string{"actor", "domain", "generation"}},
		// 수명(복합).
		"lifetime: approved_at not a canonical instant": {body: func(_ *familyActivationFixture, b *productionFamilyActivationBody) { b.ApprovedAt = "yesterday" },
			want: ErrProductionFamilyActivationUnavailable, group: "lifetime", fields: []string{"approved_at"}},
		"lifetime: issued before approved": {body: func(f *familyActivationFixture, b *productionFamilyActivationBody) {
			b.ApprovedAt = f.now.Add(-time.Minute).Format(time.RFC3339Nano)
		}, want: ErrProductionFamilyActivationUnavailable, group: "lifetime", fields: []string{"issued_at before approved_at"}},
		"lifetime: issued in the future": {body: func(f *familyActivationFixture, b *productionFamilyActivationBody) {
			b.IssuedAt = f.now.Add(time.Minute).Format(time.RFC3339Nano)
		}, want: ErrProductionFamilyActivationUnavailable, group: "lifetime", fields: []string{"issued_at after observed_at"}},
		"lifetime: longer than the ceiling": {body: func(f *familyActivationFixture, b *productionFamilyActivationBody) {
			b.ExpiresAt = f.now.Add(productionFamilyActivationMaximumLife).Add(time.Hour + time.Nanosecond).Format(time.RFC3339Nano)
		}, want: ErrProductionFamilyActivationUnavailable, group: "lifetime", fields: []string{"lifetime over maximum"}},
		"lifetime: two faults at once": {body: func(f *familyActivationFixture, b *productionFamilyActivationBody) {
			b.ApprovedAt = f.now.Add(-time.Minute).Format(time.RFC3339Nano)
			b.ExpiresAt = f.now.Add(productionFamilyActivationMaximumLife).Add(time.Hour + time.Nanosecond).Format(time.RFC3339Nano)
		}, want: ErrProductionFamilyActivationUnavailable, group: "lifetime", fields: []string{"issued_at before approved_at", "lifetime over maximum"}},
		"expired": {body: func(f *familyActivationFixture, b *productionFamilyActivationBody) {
			b.IssuedAt = f.now.Add(-2 * time.Hour).Format(time.RFC3339Nano)
			b.ExpiresAt = f.now.Add(-time.Minute).Format(time.RFC3339Nano)
		}, want: ErrProductionFamilyActivationExpired, fields: []string{"expires_at"}},
		// 서술자.
		"descriptor: unknown lane": {body: func(_ *familyActivationFixture, b *productionFamilyActivationBody) {
			b.Descriptors[0].LaneID = "kr_unknown_v1"
		},
			want: ErrProductionFamilyActivationUnavailable, group: "descriptors[lane_id=kr_unknown_v1]", fields: []string{"lane_id"}},
		"descriptor: horizon drift": {body: func(_ *familyActivationFixture, b *productionFamilyActivationBody) {
			b.Descriptors[0].Horizon = HorizonWeekly
		},
			want: ErrProductionFamilyActivationUnavailable, group: "descriptors[lane_id=" + orderedLaneIDs(MarketKR)[0] + "]", fields: []string{"horizon"}},
		"descriptor: effective without desired": {body: func(_ *familyActivationFixture, b *productionFamilyActivationBody) {
			b.Descriptors[0].Desired = StateOff
		},
			want: ErrProductionFamilyActivationUnavailable, fields: []string{"effective ON without desired ON"}},
		"descriptor: duplicate lane": {body: func(_ *familyActivationFixture, b *productionFamilyActivationBody) {
			b.Descriptors[1] = b.Descriptors[0]
		},
			want: ErrProductionFamilyActivationUnavailable, fields: []string{"duplicate"}},
		"descriptor: three of four": {body: func(_ *familyActivationFixture, b *productionFamilyActivationBody) { b.Descriptors = b.Descriptors[:3] },
			want: ErrProductionFamilyActivationUnavailable, fields: []string{"3 of 4"}},
	}
	for name, s := range shapes {
		t.Run(name, func(t *testing.T) {
			fixture := newFamilyActivationFixture(t)
			body := fixture.body(MarketKR)
			if s.body != nil {
				s.body(fixture, &body)
			}
			config := fixture.write(t, MarketKR, body)
			if s.file != nil {
				s.file(t, fixture)
				// 바이트를 바꾼 모양은 핀도 새 바이트에 맞춘다 — 형식 검사는 핀 대조 뒤에 있어 핀이 맞아야 닿는다(파일을 지운 모양은 그대로).
				if data, err := os.ReadFile(filepath.Join(fixture.dir, ProductionFamilyActivationFileName(MarketKR))); err == nil {
					config.ManifestDigest = productionRouteDigest(data)
				}
			}
			if s.config != nil {
				s.config(&config)
			}
			activation, err := LoadProductionFamilyActivation(context.Background(), config)
			if err == nil || activation.Verified() {
				t.Fatalf("accepted: %v", err)
			}
			if !errors.Is(err, s.want) {
				t.Fatalf("err=%v does not wrap %v", err, s.want)
			}
			message := err.Error()
			if s.group != "" {
				want := append([]string(nil), s.fields...)
				sort.Strings(want)
				if got := a112FieldGroup(message, s.group); strings.Join(got, "|") != strings.Join(want, "|") {
					t.Fatalf("%s fields=%q (message %q), want exactly %q", s.group, got, message, want)
				}
			} else {
				for _, field := range s.fields {
					if !strings.Contains(message, field) {
						t.Fatalf("message %q does not name %q", message, field)
					}
				}
			}
			for _, field := range s.absent {
				if strings.Contains(message, field) {
					t.Fatalf("message %q names %q — the two kinds are mixed", message, field)
				}
			}
			// 읽기 결함만 공유 읽기 함수의 오류를 사슬에 싣는다 — 불일치 · 형식 · 결속 거절에는 없다(두 종류가 섞이지 않음).
			// 그 함수는 OS 원인(없음 · 심링크 · 권한 · 소유자 · 크기)을 자기 sentinel 하나로 접는다 — OS 원인 노출은 이 로트 밖(잔여).
			if errors.Is(err, ErrProductionRouteUnavailable) != s.readErr {
				t.Fatalf("file read fault preserved=%v, want %v (message %q)", errors.Is(err, ErrProductionRouteUnavailable), s.readErr, message)
			}
		})
	}
}

func TestTheDocumentPathNamesItsRefusalToo(t *testing.T) {
	if _, err := (FamilyActivationDocument{Market: "XX"}).body(); !errors.Is(err, ErrProductionFamilyActivationUnavailable) ||
		!strings.Contains(err.Error(), "market") {
		t.Fatalf("unknown market: %v", err)
	}
	if _, err := (FamilyActivationDocument{Market: MarketKR, On: []Family{"NOT_A_FAMILY"}}).body(); !errors.Is(err, ErrProductionFamilyActivationUnavailable) ||
		!strings.Contains(err.Error(), "NOT_A_FAMILY") {
		t.Fatalf("unknown family: %v", err)
	}
	var noContext context.Context
	if _, err := LoadProductionFamilyActivation(noContext, FamilyActivationConfig{ManifestDigest: "sha256:" + strings.Repeat("a", 64)}); !errors.Is(err, ErrProductionFamilyActivationUnavailable) || !strings.Contains(err.Error(), "context") {
		t.Fatalf("nil context: %v", err)
	}
}

// a112RewriteManifest 는 쓰인 매니페스트 바이트를 바꿔 다시 쓴다(핀은 시험 몸통이 새 바이트에 맞춘다).
func a112RewriteManifest(t *testing.T, f *familyActivationFixture, change func([]byte) []byte) {
	t.Helper()
	path := filepath.Join(f.dir, ProductionFamilyActivationFileName(MarketKR))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	f.writeRaw(t, MarketKR, change(data))
}
