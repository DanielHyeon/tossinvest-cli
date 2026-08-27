package officialbars

// quote.go는 공식 API에서 읽은 호가창 맨 위 한 줄을 breakout 증거로 바꿔 적재한다.
//
// 이 파일이 지키는 것은 네 가지다.
//
//  1. 한 봉인은 두 응답에서 온다. bid/ask는 /orderbook, last는 /prices에서 오고, 어느 한쪽이
//     실패하면 아무것도 적재하지 않는다. 실패한 절반을 다시 부르지도 않는다 — 재시도는 두
//     절반 사이의 틈을 보이지 않게 벌린다(a112 결정 34).
//  2. 시각은 브로커의 것이어야 한다. 우리 시계는 신선도 검사의 기준이 될 수 없다(결정 35).
//  3. 두 절반을 묶을 때는 거절하는 방향으로 묶는다. 관측 시각은 둘 중 **이른** 것,
//     받은 시각은 둘 중 **늦은** 것이다. 어느 쪽으로 치우쳐도 호가는 더 낡아 보일 뿐
//     더 싱싱해 보이지 않는다. 새 허용 오차 상수는 만들지 않는다 — 기존 MaxQuoteAgeMS
//     거부가 그 일을 이미 한다(결정 36).
//  4. 신원은 관측 순간이다. 그래서 판 번호는 언제나 1이고, 같은 순간에 다른 본문이 오면
//     덮어쓰지 않고 격리된다(결정 40).
//
// 이 파일 안에는 time.Now도, 고루틴도, 로그도, 부동소수도 없다.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

// QuoteReader는 이 생산자가 필요로 하는 읽기 능력 전부다.
// 실제로는 official.Client의 엄격한 리더 둘이 들어오고, 시험에서는 가짜가 들어온다.
type QuoteReader interface {
	StrictOrderbookTop(ctx context.Context, market, symbol string) (official.StrictTopOfBook, error)
	StrictLastPrice(ctx context.Context, market, symbol string) (official.StrictLastPrice, error)
}

// QuoteStore는 이 생산자가 필요로 하는 저장 능력 전부다.
// 봉과 달리 이미 저장된 것을 미리 읽지 않는다 — 신원이 관측 순간이므로 같은 줄인지 아닌지는
// 저장소가 자기 규칙으로 판정하고, 그 답(멱등·충돌)을 그대로 세면 된다.
type QuoteStore interface {
	Append(context.Context, strategyevidence.Envelope) (strategyevidence.AppendResult, error)
}

// QuotePollInput은 호가 한 번 읽기에 필요한 값 전부다.
type QuotePollInput struct {
	Market marketclock.Market
	Symbol string
}

// QuotePollResult는 읽기 한 번의 결과 전부다.
type QuotePollResult struct {
	Market   string
	Symbol   string
	Currency string
	// SourceObservedAt은 두 브로커 시각 중 이른 것이다.
	SourceObservedAt time.Time
	// ReceivedAt은 두 읽기 시각 중 늦은 것이다.
	ReceivedAt           time.Time
	SourceResponseDigest string
	Revision             uint64
	Outcome              string
}

// 결과 이름표. 하나의 읽기는 셋 중 정확히 하나로 끝난다.
const (
	QuoteOutcomeAdmitted  = "ADMITTED"
	QuoteOutcomeUnchanged = "UNCHANGED"
	QuoteOutcomeConflict  = "CONFLICT"
)

// 호가 거절 사유 이름표. 봉 폴의 사유와 한 벌로 쓰인다.
const (
	RefusalQuoteReaderError = "QUOTE_READER_ERROR"
	RefusalQuoteIdentity    = "QUOTE_IDENTITY_MISMATCH"
	RefusalQuoteCurrency    = "QUOTE_CURRENCY_MISMATCH"
	RefusalQuoteInstant     = "QUOTE_INSTANT_MISSING"
	RefusalQuoteDigest      = "QUOTE_DIGEST_INVALID"
	RefusalQuoteEnvelope    = "QUOTE_ENVELOPE_REFUSED"
	RefusalQuoteStoreError  = "QUOTE_STORE_ERROR"
)

const (
	// 합성 지문의 영역 이름표. 판이 올라가면 이 글자가 바뀐다.
	quoteDigestDomain = "a112-quote-l1-v1"
	// 지문 표기의 접두사와 16진수 길이(sha256).
	quoteDigestPrefix = "sha256:"
	quoteDigestHexLen = 64
	// 판 번호는 언제나 1이다. 신원이 관측 순간이라 "같은 호가의 다음 판"이라는 것이
	// 있을 수 없다 — 순간이 다르면 다른 줄이고, 순간이 같은데 본문이 다르면 충돌이다.
	quoteRevision = 1
)

// PollQuoteL1은 한 종목의 호가 맨 위 한 줄을 한 번 읽어서 증거로 적재한다.
func PollQuoteL1(ctx context.Context, reader QuoteReader, store QuoteStore, in QuotePollInput) (QuotePollResult, error) {
	var result QuotePollResult
	if reader == nil || store == nil {
		return result, refuse(RefusalDependencyMissing, "a reader and a store are required", nil)
	}
	market, err := marketclock.ParseMarket(string(in.Market))
	if err != nil {
		return result, refuse(RefusalMarketInvalid, "market "+strconv.Quote(string(in.Market))+" is not kr or us", nil)
	}
	code := marketCode(market)
	if err := checkSymbol(in.Symbol); err != nil {
		return result, err
	}

	// ---- (a) 두 응답을 등을 맞대고 읽는다. 순서는 고정이다(지문이 그 순서를 담는다). ----
	top, err := reader.StrictOrderbookTop(ctx, code, in.Symbol)
	if err != nil {
		return result, refuse(RefusalQuoteReaderError, "reading the top of book", err)
	}
	last, err := reader.StrictLastPrice(ctx, code, in.Symbol)
	if err != nil {
		return result, refuse(RefusalQuoteReaderError, "reading the last price", err)
	}

	// ---- (b) 두 절반이 같은 것을 말하는지 본다 ----
	if err := checkQuoteIdentity(code, in.Symbol, top.Market, top.Symbol, "top of book"); err != nil {
		return result, err
	}
	if err := checkQuoteIdentity(code, in.Symbol, last.Market, last.Symbol, "last price"); err != nil {
		return result, err
	}
	if top.Currency != last.Currency {
		return result, refuse(RefusalQuoteCurrency,
			"the top of book is in "+strconv.Quote(top.Currency)+" and the last price in "+strconv.Quote(last.Currency), nil)
	}
	for _, instant := range []struct {
		name  string
		value time.Time
	}{
		{"the top of book's source instant", top.SourceInstant},
		{"the top of book's read instant", top.ReadAt},
		{"the last price's source instant", last.SourceInstant},
		{"the last price's read instant", last.ReadAt},
	} {
		if instant.value.IsZero() {
			return result, refuse(RefusalQuoteInstant, instant.name+" is missing", nil)
		}
	}

	// ---- (c) 거절하는 방향으로 묶는다 ----
	observedAt := earlier(top.SourceInstant, last.SourceInstant)
	receivedAt := later(top.ReadAt, last.ReadAt)
	digest, err := quoteSealDigest(top.BodyDigest, last.BodyDigest)
	if err != nil {
		return result, err
	}

	envelope, err := strategyevidence.NewQuoteL1Envelope(strategyevidence.QuoteL1Input{
		Market: market, Symbol: in.Symbol, SourceObservedAt: observedAt, ReceivedAt: receivedAt,
		Revision: quoteRevision, Currency: top.Currency,
		Raw: strategyevidence.RawQuoteL1{
			Bid: top.Bid.Price, Ask: top.Ask.Price, Last: last.Last,
		},
		SourceResponseDigest: digest,
	})
	if err != nil {
		return result, refuse(RefusalQuoteEnvelope, "the quote is not evidence", err)
	}

	result = QuotePollResult{
		Market: code, Symbol: in.Symbol, Currency: top.Currency,
		SourceObservedAt: observedAt.UTC(), ReceivedAt: receivedAt.UTC(),
		SourceResponseDigest: digest, Revision: quoteRevision,
	}

	// ---- (d) 적재. 같은 줄인지 아닌지는 저장소가 판정한다 ----
	appended, err := store.Append(ctx, envelope)
	switch {
	case errors.Is(err, strategyevidence.ErrRevisionConflict):
		// 같은 순간에 다른 본문이 왔다. 먼저 있던 줄은 그대로 두고 이쪽을 격리한다.
		result.Outcome = QuoteOutcomeConflict
	case err != nil:
		return result, refuse(RefusalQuoteStoreError,
			"appending the quote observed at "+observedAt.UTC().Format(time.RFC3339Nano), err)
	case appended.Idempotent:
		result.Outcome = QuoteOutcomeUnchanged
	default:
		result.Outcome = QuoteOutcomeAdmitted
	}
	return result, nil
}

// checkQuoteIdentity는 응답이 우리가 물어본 그 종목에 대한 답인지 본다.
func checkQuoteIdentity(code, symbol, gotMarket, gotSymbol, half string) error {
	if gotMarket != code || gotSymbol != symbol {
		return refuse(RefusalQuoteIdentity,
			"the "+half+" answers for "+strconv.Quote(gotMarket+"/"+gotSymbol)+
				", not for "+strconv.Quote(code+"/"+symbol), nil)
	}
	return nil
}

// quoteSealDigest는 두 본문의 지문을 순서가 고정된 하나로 합친다(결정 37).
//
// 한쪽 본문만 덮는 지문은 증거의 절반을 출처 없는 것으로 만든다. 순서(호가가 먼저)와
// 영역 이름표를 함께 넣으므로, 같은 두 본문에서는 언제나 같은 값이 나오고 다른 조합에서는
// 나오지 않는다.
func quoteSealDigest(orderbookDigest, pricesDigest string) (string, error) {
	orderbook, err := quoteDigestHex(orderbookDigest, "the top of book")
	if err != nil {
		return "", err
	}
	prices, err := quoteDigestHex(pricesDigest, "the last price")
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(quoteDigestDomain + "\n" + orderbook + "\n" + prices))
	return quoteDigestPrefix + hex.EncodeToString(sum[:]), nil
}

// quoteDigestHex는 "sha256:"을 떼고 16진수만 돌려준다. 모양이 다르면 거절이다.
func quoteDigestHex(value, half string) (string, error) {
	if !strings.HasPrefix(value, quoteDigestPrefix) {
		return "", refuse(RefusalQuoteDigest,
			half+" carries the digest "+strconv.Quote(value)+", which does not start with "+quoteDigestPrefix, nil)
	}
	text := value[len(quoteDigestPrefix):]
	if len(text) != quoteDigestHexLen {
		return "", refuse(RefusalQuoteDigest,
			half+" carries a digest of "+strconv.Itoa(len(text))+" hex digits, not "+strconv.Itoa(quoteDigestHexLen), nil)
	}
	if _, err := hex.DecodeString(text); err != nil {
		return "", refuse(RefusalQuoteDigest, half+"'s digest is not hexadecimal", err)
	}
	return text, nil
}

func earlier(left, right time.Time) time.Time {
	if right.Before(left) {
		return right
	}
	return left
}

func later(left, right time.Time) time.Time {
	if right.After(left) {
		return right
	}
	return left
}
