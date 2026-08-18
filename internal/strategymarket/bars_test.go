package strategymarket

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func minutes() []RawMinuteCandle {
	out := make([]RawMinuteCandle, 5)
	for i := range out {
		out[i] = RawMinuteCandle{Timestamp: time.Date(2026, 7, 31, 9, i+1, 0, 0, time.FixedZone("KST", 9*3600)).Format(time.RFC3339), Open: "100.01", High: "101.10", Low: "99.90", Close: "100.20", Volume: "0.1", Currency: "KRW"}
	}
	out[4].Close = "100.55"
	return out
}
func officialPage(t *testing.T, symbol string, candles []RawMinuteCandle, adjusted bool) OfficialMinutePage {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/oauth2/token":
			_ = json.NewEncoder(writer).Encode(map[string]any{"access_token": "AT", "expires_in": 3600, "token_type": "Bearer"})
		case "/api/v1/candles":
			rows := make([]map[string]string, len(candles))
			for index, candle := range candles {
				rows[index] = map[string]string{"timestamp": candle.Timestamp, "openPrice": candle.Open, "highPrice": candle.High, "lowPrice": candle.Low, "closePrice": candle.Close, "volume": candle.Volume, "currency": candle.Currency}
			}
			_ = json.NewEncoder(writer).Encode(map[string]any{"result": map[string]any{"candles": rows}})
		default:
			http.NotFound(writer, request)
		}
	}))
	t.Cleanup(server.Close)
	client := official.New(official.Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "token.json"), official.WithBaseURL(server.URL), official.WithHTTPClient(server.Client()))
	page, err := client.RawMinuteCandles(context.Background(), "KR", symbol, 5, "", adjusted)
	if err != nil {
		t.Fatal(err)
	}
	adapted := make([]RawMinuteCandle, len(page.Candles()))
	for index, candle := range page.Candles() {
		adapted[index] = RawMinuteCandle{Timestamp: candle.Timestamp, Open: candle.Open, High: candle.High, Low: candle.Low, Close: candle.Close, Volume: candle.Volume, Currency: candle.Currency}
	}
	return SealAdaptedOfficialMinutePage(page.Market(), page.Symbol(), page.Interval(), page.Adjusted(), SourceOfficialOpenAPI, adapted)
}
func TestAggregateClosedKRXFiveMinutePreservesExactDecimals(t *testing.T) {
	got, err := SealOfficialClosedKRXFiveMinute(officialPage(t, "005930", minutes(), false), time.Date(2026, 7, 31, 9, 5, 0, 0, time.FixedZone("KST", 9*3600)))
	if err != nil {
		t.Fatal(err)
	}
	if !got.Valid() || got.Market() != "KR" || got.Symbol() != "005930" || got.Source() != string(SourceOfficialOpenAPI) ||
		got.Adjusted() || got.Open() != "100.01" || got.High() != "101.1" || got.Low() != "99.9" || got.Close() != "100.55" || got.Volume() != "0.5" {
		t.Fatalf("bar=%+v", got)
	}
}

func TestLegacyRawSliceCannotMintVerifiedBar(t *testing.T) {
	bar, err := AggregateClosedKRXFiveMinute(minutes(), time.Now())
	var integrity *IntegrityError
	if bar.Valid() || !errors.As(err, &integrity) || integrity.Kind != RefusalSource {
		t.Fatalf("bar=%+v err=%v", bar, err)
	}
}

func TestAggregateClosedKRXFiveMinuteFailsClosed(t *testing.T) {
	base := minutes()
	now := time.Date(2026, 7, 31, 9, 5, 0, 0, time.FixedZone("KST", 9*3600))
	tests := []struct {
		name string
		page OfficialMinutePage
		now  time.Time
		kind Refusal
	}{
		{"missing", officialPage(t, "005930", base[:4], false), now, RefusalIncompleteBucket},
		{"open", officialPage(t, "005930", minutes(), false), now.Add(-time.Nanosecond), RefusalOpenBucket},
		{"naive", officialPage(t, "005930", func() []RawMinuteCandle { v := minutes(); v[0].Timestamp = "2026-07-31T09:00:00"; return v }(), false), now, RefusalNaiveTimestamp},
		{"gap", officialPage(t, "005930", func() []RawMinuteCandle {
			v := minutes()
			v[2].Timestamp = time.Date(2026, 7, 31, 9, 7, 0, 0, time.FixedZone("KST", 9*3600)).Format(time.RFC3339)
			return v
		}(), false), now.Add(3 * time.Minute), RefusalMinuteGap},
		{"outside", officialPage(t, "005930", func() []RawMinuteCandle {
			v := minutes()
			for i := range v {
				v[i].Timestamp = time.Date(2026, 7, 31, 8, 55+i, 0, 0, time.FixedZone("KST", 9*3600)).Format(time.RFC3339)
			}
			return v
		}(), false), now, RefusalOutsideRegularSession},
		{"decimal", officialPage(t, "005930", func() []RawMinuteCandle { v := minutes(); v[0].Open = "1e2"; return v }(), false), now, RefusalInvalidDecimal},
		{"wrong symbol", officialPage(t, "000660", minutes(), false), now, RefusalIdentity},
		{"adjusted", officialPage(t, "005930", minutes(), true), now, RefusalAdjusted},
		{"unofficial", SealAdaptedOfficialMinutePage("KR", "005930", IntervalOneMinute, false, "wts", minutes()), now, RefusalSource},
		{"wrong interval", SealAdaptedOfficialMinutePage("KR", "005930", "5m", false, SourceOfficialOpenAPI, minutes()), now, RefusalInterval},
		{"zero page", OfficialMinutePage{}, now, RefusalSource},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := SealOfficialClosedKRXFiveMinuteFor("KR", "005930", tc.page, tc.now)
			var ie *IntegrityError
			if !errors.As(err, &ie) || ie.Kind != tc.kind {
				t.Fatalf("err=%v want=%s", err, tc.kind)
			}
		})
	}
}

type stateStub struct {
	reading StateReading
	err     error
}

func (s stateStub) ReadSymbolState(context.Context, string, string) (StateReading, error) {
	return s.reading, s.err
}
func TestAuthoritativeSymbolStateMustBePresentFreshAndNormal(t *testing.T) {
	now := time.Now().UTC()
	proof, err := RequireFreshNormalState(context.Background(), stateStub{reading: StateReading{Market: "KR", Symbol: "005930", State: StateNormal, ObservedAt: now, Source: SourceOfficialSymbolState}}, "KR", "005930", now)
	if err != nil {
		t.Fatal(err)
	}
	if !proof.Valid() || proof.Market() != "KR" || proof.Symbol() != "005930" || proof.Authority() != string(SourceOfficialSymbolState) {
		t.Fatalf("proof=%+v", proof)
	}
	if (FreshNormalState{}).Valid() || (VerifiedBar{}).Valid() {
		t.Fatal("zero proof accepted")
	}
	for _, source := range []SymbolStateSource{nil, stateStub{err: context.Canceled},
		stateStub{reading: StateReading{Market: "KR", Symbol: "005930", State: StateNormal, ObservedAt: now.Add(-31 * time.Second), Source: SourceOfficialSymbolState}},
		stateStub{reading: StateReading{Market: "KR", Symbol: "005930", State: StateHalt, ObservedAt: now, Source: SourceOfficialSymbolState}},
		stateStub{reading: StateReading{Market: "KR", Symbol: "000660", State: StateNormal, ObservedAt: now, Source: SourceOfficialSymbolState}},
		stateStub{reading: StateReading{Market: "KR", Symbol: "005930", State: StateNormal, ObservedAt: now, Source: "caller-claimed"}}} {
		if _, err := RequireFreshNormalState(context.Background(), source, "KR", "005930", now); err == nil {
			t.Fatal("blocked state accepted")
		}
	}
}

type positionStub struct {
	reading PositionReading
	err     error
}

func (s positionStub) ReadPosition(context.Context, string, string) (PositionReading, error) {
	return s.reading, s.err
}

func TestNoPositionProofRequiresFreshAuthoritativeZeroPositionAndOrders(t *testing.T) {
	now := time.Now().UTC()
	valid := positionStub{reading: PositionReading{Market: "KR", Symbol: "005930", Quantity: "0", ObservedAt: now, Source: SourceOfficialPosition}}
	proof, err := RequireNoPosition(context.Background(), valid, "KR", "005930", now)
	if err != nil || !proof.Valid() || proof.Market() != "KR" || proof.Symbol() != "005930" || proof.Authority() != string(SourceOfficialPosition) {
		t.Fatalf("proof=%+v err=%v", proof, err)
	}
	if (NoPositionProof{}).Valid() {
		t.Fatal("zero position proof accepted")
	}
	for _, source := range []PositionSource{
		nil,
		positionStub{err: context.Canceled},
		positionStub{reading: PositionReading{Market: "KR", Symbol: "005930", Quantity: "0", ObservedAt: now.Add(-31 * time.Second), Source: SourceOfficialPosition}},
		positionStub{reading: PositionReading{Market: "KR", Symbol: "005930", Quantity: "1", ObservedAt: now, Source: SourceOfficialPosition}},
		positionStub{reading: PositionReading{Market: "KR", Symbol: "005930", Quantity: "0", OpenOrders: 1, ObservedAt: now, Source: SourceOfficialPosition}},
		positionStub{reading: PositionReading{Market: "KR", Symbol: "005930", Quantity: "0.0", ObservedAt: now, Source: SourceOfficialPosition}},
		positionStub{reading: PositionReading{Market: "KR", Symbol: "000660", Quantity: "0", ObservedAt: now, Source: SourceOfficialPosition}},
		positionStub{reading: PositionReading{Market: "KR", Symbol: "005930", Quantity: "0", ObservedAt: now, Source: "caller-claimed"}},
	} {
		if _, err := RequireNoPosition(context.Background(), source, "KR", "005930", now); err == nil {
			t.Fatal("untrusted or non-empty position accepted")
		}
	}
}

// a117 RED: 브로커가 주는 분봉 시각표는 봉이 "닫힌" 시각이다(a112 결정 30·31).
// 그래서 라벨 t인 1분봉은 [t-1분, t) 구간을 담고, 5분 버킷의 여는 시각은
// 첫 라벨보다 1분 이르다. 아래 세 테스트는 그 규칙을 각 방향에서 못박는다.

// labelledMinutes는 first부터 1분 간격으로 붙은 다섯 개의 봉 라벨을 만든다.
func labelledMinutes(first time.Time) []RawMinuteCandle {
	out := make([]RawMinuteCandle, 5)
	for i := range out {
		out[i] = RawMinuteCandle{
			Timestamp: first.Add(time.Duration(i) * time.Minute).Format(time.RFC3339),
			Open:      "100.01", High: "101.10", Low: "99.90", Close: "100.20", Volume: "0.1", Currency: "KRW",
		}
	}
	return out
}

// TestTheBucketOpensOneMinuteBeforeItsFirstLabel은 장 시작 첫 5분 버킷을 확인한다.
// 09:00~09:05 버킷은 라벨 09:01~09:05인 다섯 봉으로 이루어진다.
func TestTheBucketOpensOneMinuteBeforeItsFirstLabel(t *testing.T) {
	kst := time.FixedZone("KST", 9*3600)
	first := time.Date(2026, 7, 31, 9, 1, 0, 0, kst)
	got, err := SealOfficialClosedKRXFiveMinute(
		officialPage(t, "005930", labelledMinutes(first), false),
		time.Date(2026, 7, 31, 9, 5, 0, 0, kst))
	if err != nil {
		t.Fatalf("첫 정규 버킷이 거절됐다: %v", err)
	}
	wantOpen := time.Date(2026, 7, 31, 9, 0, 0, 0, kst)
	wantClose := time.Date(2026, 7, 31, 9, 5, 0, 0, kst)
	if !got.OpenAt().Equal(wantOpen) || !got.ClosedAt().Equal(wantClose) {
		t.Fatalf("open=%s close=%s want open=%s close=%s",
			got.OpenAt(), got.ClosedAt(), wantOpen, wantClose)
	}
}

// TestABarLabelledAtTheOpenIsOutsideTheRegularSession은 09:00 라벨 봉이
// 08:59~09:00, 즉 개장 전 1분을 담는다는 사실을 못박는다.
func TestABarLabelledAtTheOpenIsOutsideTheRegularSession(t *testing.T) {
	kst := time.FixedZone("KST", 9*3600)
	first := time.Date(2026, 7, 31, 9, 0, 0, 0, kst)
	_, err := SealOfficialClosedKRXFiveMinute(
		officialPage(t, "005930", labelledMinutes(first), false),
		time.Date(2026, 7, 31, 9, 5, 0, 0, kst))
	var integrity *IntegrityError
	if !errors.As(err, &integrity) || integrity.Kind != RefusalOutsideRegularSession {
		t.Fatalf("개장 전 1분을 담은 봉이 통과했다: err=%v", err)
	}
}

// TestTheLastRegularMinuteLabelIsAdmitted는 15:30 라벨 봉이 15:29~15:30,
// 즉 정규장 마지막 1분이라는 사실을 못박는다.
func TestTheLastRegularMinuteLabelIsAdmitted(t *testing.T) {
	kst := time.FixedZone("KST", 9*3600)
	first := time.Date(2026, 7, 31, 15, 26, 0, 0, kst)
	got, err := SealOfficialClosedKRXFiveMinute(
		officialPage(t, "005930", labelledMinutes(first), false),
		time.Date(2026, 7, 31, 15, 30, 0, 0, kst))
	if err != nil {
		t.Fatalf("정규장 마지막 버킷이 거절됐다: %v", err)
	}
	wantOpen := time.Date(2026, 7, 31, 15, 25, 0, 0, kst)
	wantClose := time.Date(2026, 7, 31, 15, 30, 0, 0, kst)
	if !got.OpenAt().Equal(wantOpen) || !got.ClosedAt().Equal(wantClose) {
		t.Fatalf("open=%s close=%s want open=%s close=%s",
			got.OpenAt(), got.ClosedAt(), wantOpen, wantClose)
	}
}
