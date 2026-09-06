package strategyevidence

// 선언된 fatal fact 가 **조용히 사라지지 않는지** 잰다.
//
// Project 의 fatal 루프는 refusal 이 나면 requirement.Required 가 참일 때만 이유를
// 적고 그렇지 않으면 continue 한다. lane 루프에는 Unavailable 지도가 있어 흔적이 남지만
// fatal 루프에는 없었다. "거래정지 증거를 fatal 로 선언했는데 그 증거가 없었다"는
// 사실이 결과 어디에도 남지 않으면, 읽는 쪽은 그것을 "fatal 없음"과 구별할 수 없다.

import (
	"testing"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func fatalProjectionSnapshot(t *testing.T) Snapshot {
	t.Helper()
	header := validHeader(marketclock.MarketUS, KindUSParticipation, "fatal-scope", "rev-1")
	return Snapshot{
		ID:                   "snapshot-" + goldenSnapshotDigest,
		Digest:               goldenSnapshotDigest,
		Market:               marketclock.MarketUS,
		Symbol:               "AAPL",
		IssuerIdentity:       header.IssuerIdentity,
		IssuerMappingVersion: header.IssuerMappingVersion,
		EvaluationAt:         header.IngestedAt.Add(time.Hour),
		IngestionCutoff:      header.IngestedAt.Add(time.Hour),
		Items:                []Envelope{mustEnvelope(t, header, `{"value":1}`)},
	}
}

func TestProjectRecordsMissingFatalEvidenceEvenWhenItIsOptional(t *testing.T) {
	t.Parallel()
	snapshot := fatalProjectionSnapshot(t)
	for _, test := range []struct {
		name        string
		required    bool
		wantBlocked bool
		wantReasons int
	}{
		{name: "required", required: true, wantBlocked: true, wantReasons: 1},
		{name: "optional", required: false, wantBlocked: false, wantReasons: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result, err := Project(snapshot, ProjectionPolicy{
				Version: "fatal-trace-v1",
				Market:  marketclock.MarketUS,
				Fatal: []Requirement{{
					Kind:        KindDisclosureRisk,
					Required:    test.required,
					MaxAge:      48 * time.Hour,
					Authorities: []SourceAuthority{AuthoritySEC},
				}},
			})
			if err != nil {
				t.Fatal(err)
			}
			if result.Fatal.Blocked != test.wantBlocked || len(result.Fatal.Reasons) != test.wantReasons {
				t.Fatalf("fatal assessment blocked=%v reasons=%d, want %v/%d",
					result.Fatal.Blocked, len(result.Fatal.Reasons), test.wantBlocked, test.wantReasons)
			}
			refusal, recorded := result.Fatal.Unavailable[KindDisclosureRisk]
			if !recorded {
				t.Fatalf("declared fatal fact %s went missing without a trace: %+v", KindDisclosureRisk, result.Fatal)
			}
			if refusal.Code != RefusalEvidenceMissing || refusal.Kind != KindDisclosureRisk {
				t.Fatalf("fatal unavailability recorded as %+v", refusal)
			}
		})
	}
}

// TestProjectLeavesFatalUnavailableEmptyWhenTheFactIsPresent 는 위 시험이 "무조건
// 채운다"로 통과하지 않게 하는 양성 대조군이다.
func TestProjectLeavesFatalUnavailableEmptyWhenTheFactIsPresent(t *testing.T) {
	t.Parallel()
	header := validHeader(marketclock.MarketUS, KindDisclosureRisk, "fatal-present", "rev-1")
	snapshot := fatalProjectionSnapshot(t)
	snapshot.Items = append(snapshot.Items, mustEnvelope(t, header, `{"blocked":false}`))
	result, err := Project(snapshot, ProjectionPolicy{
		Version: "fatal-trace-v1",
		Market:  marketclock.MarketUS,
		Fatal: []Requirement{{
			Kind:        KindDisclosureRisk,
			Required:    true,
			MaxAge:      48 * time.Hour,
			Authorities: []SourceAuthority{AuthoritySEC},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Fatal.Blocked || len(result.Fatal.Reasons) != 0 || len(result.Fatal.Unavailable) != 0 {
		t.Fatalf("present fatal fact produced %+v", result.Fatal)
	}
}

func TestProjectRefusesAnInvalidPolicyOrSnapshotScope(t *testing.T) {
	t.Parallel()
	// Project 의 첫 early return 은 어떤 시험도 밟은 적이 없다. 네 시험 모두 정상
	// 범위를 주고 `err != nil` 이면 t.Fatal 한다 — 즉 거짓 갈래만 돌았다.
	for _, test := range []struct {
		name   string
		mutate func(*Snapshot, *ProjectionPolicy)
	}{
		{"blank-version", func(_ *Snapshot, p *ProjectionPolicy) { p.Version = "  " }},
		{"market-mismatch", func(_ *Snapshot, p *ProjectionPolicy) { p.Market = marketclock.MarketKR }},
		{"blank-issuer", func(s *Snapshot, _ *ProjectionPolicy) { s.IssuerIdentity = " " }},
		{"blank-mapping", func(s *Snapshot, _ *ProjectionPolicy) { s.IssuerMappingVersion = "" }},
		{"zero-evaluation", func(s *Snapshot, _ *ProjectionPolicy) { s.EvaluationAt = time.Time{} }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			snapshot := fatalProjectionSnapshot(t)
			policy := ProjectionPolicy{Version: "scope-v1", Market: marketclock.MarketUS}
			test.mutate(&snapshot, &policy)
			result, err := Project(snapshot, policy)
			if err == nil {
				t.Fatalf("invalid %s scope produced a projection: %+v", test.name, result)
			}
		})
	}

	// 양성 대조군: 손대지 않은 범위는 통과한다.
	if _, err := Project(fatalProjectionSnapshot(t), ProjectionPolicy{Version: "scope-v1", Market: marketclock.MarketUS}); err != nil {
		t.Fatalf("a valid scope was refused: %v", err)
	}
}

func TestProjectOrdersFatalReasonsAndLaneRefusalsDeterministically(t *testing.T) {
	t.Parallel()
	// 두 정렬 비교자는 이유가 둘 이상일 때만 실행된다. 지금까지 어떤 시험도 둘을
	// 만들지 않아, `<` 를 `>` 로 뒤집어도 아무것도 깨지지 않았다. 선언 순서는 기대
	// 순서의 반대로 준다 — 정렬이 실제로 자리를 바꿔야만 통과한다.
	snapshot := fatalProjectionSnapshot(t)
	result, err := Project(snapshot, ProjectionPolicy{
		Version: "ordering-v1",
		Market:  marketclock.MarketUS,
		Fatal: []Requirement{
			{Kind: KindTradability, Required: true, MaxAge: time.Hour, Authorities: []SourceAuthority{AuthoritySEC}},
			{Kind: KindDisclosureRisk, Required: true, MaxAge: time.Hour, Authorities: []SourceAuthority{AuthoritySEC}},
		},
		Required: []Requirement{
			{Kind: KindUSParticipation, Required: true, MaxAge: time.Nanosecond, Authorities: []SourceAuthority{AuthoritySEC}},
			{Kind: KindDisclosure, Required: true, MaxAge: time.Hour, Authorities: []SourceAuthority{AuthoritySEC}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Fatal.Reasons) != 2 {
		t.Fatalf("fatal reasons = %+v, want two", result.Fatal.Reasons)
	}
	if result.Fatal.Reasons[0].Kind != KindDisclosureRisk || result.Fatal.Reasons[1].Kind != KindTradability {
		t.Fatalf("fatal reasons are not ordered by kind: %+v", result.Fatal.Reasons)
	}
	if len(result.Lane.Refusals) != 2 {
		t.Fatalf("lane refusals = %+v, want two", result.Lane.Refusals)
	}
	if result.Lane.Refusals[0].Kind != KindDisclosure || result.Lane.Refusals[1].Kind != KindUSParticipation {
		t.Fatalf("lane refusals are not ordered by kind: %+v", result.Lane.Refusals)
	}
}
