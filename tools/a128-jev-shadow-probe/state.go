package main

// state.go 는 외부 판단 API 로 나가는 state 를 조립함.
//
// 안전 요구(spec 「shadow 판단은 … 계좌 데이터를 내보내지 않는다」): 전송 state 는 공개 시장
// 데이터만. 차단은 세 겹임 —
//  1. 타입: PublicState 의 필드는 공개 시세 값뿐이고, 조립기 입력(marketInputs)도 공개 읽기
//     결과 타입만 받음. 계좌·보유·잔고·토큰을 담을 필드가 어디에도 없음.
//  2. 바이트: 직렬화된 요청을 assertPublicJSON 이 열쇠 이름으로 다시 검사(judge.go).
//  3. 시험: PublicState 의 JSON 열쇠 집합을 고정 목록으로 못 박고, 계좌 필드를 섞은 mock
//     응답이 전송 바이트에 새지 않음을 확인(state_test.go).

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// newsAbsent 는 뉴스 소스가 없음을 모델에 알리는 값 — 조용히 빼면 p 의 의미가 달라짐(design.md).
// 이 로트는 WTS 뉴스를 배선하지 않음(WTS 세션 = 사람 관문 0.3) → 언제나 이 값.
const (
	newsAbsent    = "absent"
	newsGapReason = "news: absent (WTS news source is not wired in this probe)"
)

// PublicState 는 판단 요청의 state 전부임. 필드를 더하려면 state_test.go 의 고정 목록도
// 같이 바꿔야 시험이 통과함 — 새 필드가 리뷰 없이 외부로 나가지 못하게 하는 장치.
type PublicState struct {
	Market               string          `json:"market"`
	Symbol               string          `json:"symbol"`
	Name                 string          `json:"name,omitempty"`
	ListingMarket        string          `json:"listing_market,omitempty"`
	Currency             string          `json:"currency"`
	ObservedAt           string          `json:"observed_at"`
	HorizonMinutes       int             `json:"horizon_minutes"`
	BasePrice            float64         `json:"base_price"`
	ReferencePrice       float64         `json:"reference_price,omitempty"`
	ChangePctVsReference *float64        `json:"change_pct_vs_reference,omitempty"`
	RankingType          string          `json:"ranking_type"`
	RankingRank          int             `json:"ranking_rank"`
	TradingAmount        float64         `json:"trading_amount"`
	TradingVolume        float64         `json:"trading_volume"`
	TopOfBook            *TopOfBookState `json:"top_of_book,omitempty"`
	MinutesSinceOpen     int             `json:"minutes_since_open"`
	MinutesToClose       int             `json:"minutes_to_close"`
	News                 string          `json:"news"`
	MissingInputs        []string        `json:"missing_inputs"`
}

// TopOfBookState 는 호가 맨 위 한 줄과 그 나이(브로커 시각 기준).
type TopOfBookState struct {
	BestAsk          float64 `json:"best_ask"`
	BestAskVolume    float64 `json:"best_ask_volume"`
	BestBid          float64 `json:"best_bid"`
	BestBidVolume    float64 `json:"best_bid_volume"`
	SpreadBps        float64 `json:"spread_bps"`
	SourceAgeSeconds float64 `json:"source_age_seconds"`
}

// marketInputs 는 조립기 입력 — 공개 읽기 결과 타입만 받음(계좌 읽기 결과 타입은 받을 자리가 없음).
type marketInputs struct {
	Market      string
	ObservedAt  time.Time
	Horizon     time.Duration
	Session     session
	RankingType string
	Rank        domain.RankingItem
	Meta        *domain.Quote             // /stocks — nil 이면 결손
	LastPrice   float64                   // /prices — 0 초과만 여기까지 옴
	Book        *official.StrictTopOfBook // /orderbook — nil 이면 결손
	Gaps        []string                  // 수집 단계에서 이미 난 결손
}

// assembleState 는 입력을 PublicState 로 옮기고, 빠진 입력을 MissingInputs 에 이름으로 남김.
func assembleState(in marketInputs) PublicState {
	state := PublicState{
		Market:           in.Market,
		Symbol:           in.Rank.Symbol,
		Currency:         in.Rank.Currency,
		ObservedAt:       in.ObservedAt.UTC().Format(time.RFC3339),
		HorizonMinutes:   int(in.Horizon / time.Minute),
		BasePrice:        in.LastPrice,
		ReferencePrice:   in.Rank.BasePrice,
		RankingType:      in.RankingType,
		RankingRank:      in.Rank.Rank,
		TradingAmount:    in.Rank.TradingAmount,
		TradingVolume:    in.Rank.TradingVolume,
		MinutesSinceOpen: int(in.ObservedAt.Sub(in.Session.Open) / time.Minute),
		MinutesToClose:   int(in.Session.Close.Sub(in.ObservedAt) / time.Minute),
		News:             newsAbsent,
		MissingInputs:    append([]string{}, in.Gaps...),
	}
	// 변화율은 단위가 확실한 값만 — ranking 의 lastPrice/basePrice 로 직접 계산(changeRate 의 단위는 미확인).
	if in.Rank.BasePrice > 0 && in.Rank.LastPrice > 0 {
		pct := round((in.Rank.LastPrice-in.Rank.BasePrice)/in.Rank.BasePrice*100, 4)
		state.ChangePctVsReference = &pct
	}
	if in.Meta != nil {
		state.Name = in.Meta.Name
		state.ListingMarket = in.Meta.MarketCode
		if state.Currency == "" {
			state.Currency = in.Meta.Currency
		}
	}
	if in.Book != nil {
		if top, ok := topOfBookState(*in.Book, in.ObservedAt); ok {
			state.TopOfBook = top
		} else {
			state.MissingInputs = append(state.MissingInputs, "top_of_book: decimal not parseable")
		}
	}
	state.MissingInputs = append(state.MissingInputs, newsGapReason)
	sort.Strings(state.MissingInputs)
	return state
}

func topOfBookState(book official.StrictTopOfBook, observedAt time.Time) (*TopOfBookState, bool) {
	ask, errAsk := strconv.ParseFloat(book.Ask.Price, 64)
	askVolume, errAskVolume := strconv.ParseFloat(book.Ask.Volume, 64)
	bid, errBid := strconv.ParseFloat(book.Bid.Price, 64)
	bidVolume, errBidVolume := strconv.ParseFloat(book.Bid.Volume, 64)
	if errAsk != nil || errAskVolume != nil || errBid != nil || errBidVolume != nil {
		return nil, false
	}
	top := &TopOfBookState{
		BestAsk: ask, BestAskVolume: askVolume, BestBid: bid, BestBidVolume: bidVolume,
		SourceAgeSeconds: round(observedAt.Sub(book.SourceInstant).Seconds(), 1),
	}
	// 스프레드는 중간가 대비 bp. 중간가가 0 이하이면 계산하지 않음(0 으로 둠).
	if mid := (ask + bid) / 2; mid > 0 {
		top.SpreadBps = round((ask-bid)/mid*10000, 2)
	}
	return top, true
}

func round(value float64, digits int) float64 {
	scale := math.Pow(10, float64(digits))
	return math.Round(value*scale) / scale
}

// forbiddenStateKey 는 계좌·보유·잔고·시크릿·세션을 가리키는 열쇠 이름임.
// 공개 시장 필드 이름과 겹치지 않음을 state_test.go 가 확인함.
var forbiddenStateKey = regexp.MustCompile(`(?i)account|acct|holding|balance|position|cash|deposit|asset|` +
	`token|secret|password|passwd|session_?id|cookie|api_?key|bearer|credential|seq|quantity|qty|owner|user`)

// assertPublicJSON 은 직렬화된 요청의 모든 열쇠를 재귀로 훑어 금지 이름이 있으면 전송을 거부함.
// questions 맵의 열쇠(질문 id)도 함께 검사됨 — 질문 id 도 전송 바이트의 일부이기 때문.
func assertPublicJSON(raw []byte) error {
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return fmt.Errorf("state guard: request is not JSON: %w", err)
	}
	var walk func(path string, node any) error
	walk = func(path string, node any) error {
		switch typed := node.(type) {
		case map[string]any:
			for key, child := range typed {
				if forbiddenStateKey.MatchString(key) {
					return fmt.Errorf("state guard: refusing to send key %q at %s — account/secret-shaped fields never leave the probe", key, path)
				}
				if err := walk(path+"."+key, child); err != nil {
					return err
				}
			}
		case []any:
			for index, child := range typed {
				if err := walk(fmt.Sprintf("%s[%d]", path, index), child); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return walk("$", decoded)
}

// probeQuestions 는 J1·J2 원문임(design.md 「판단 설계」). 영어인 이유: models.md 「English is the
// primary training language」. 원문을 바꾸면 questionRev 를 올릴 것 — digest 는 자동으로 바뀜.
const questionRev = "a128-r1"

func probeQuestions() []question {
	return []question{
		{
			ID:  "j1_up_within_horizon",
			Rev: questionRev,
			Body: noulQuestion{
				Type: "noul",
				Instructions: "Will this stock's last traded price, exactly `horizon_minutes` minutes after " +
					"`observed_at`, be higher than `base_price`?",
				Criteria: &noulCriteria{
					True:  "The last traded price at `observed_at` plus `horizon_minutes` is strictly above `base_price`.",
					False: "That price is equal to or below `base_price`.",
				},
			},
		},
		{
			ID:  "j2_hold_off_red_flag",
			Rev: questionRev,
			Body: noulQuestion{
				Type: "noul",
				Instructions: "Does this state show a clear red flag that should make a cautious buyer hold off " +
					"on buying this stock right now?",
				Criteria: &noulCriteria{
					True:  "The state contains an explicit adverse signal for buying now.",
					False: "No adverse signal for buying now is visible in the state.",
				},
			},
		},
	}
}
