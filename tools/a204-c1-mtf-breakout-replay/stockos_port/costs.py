# stockos-port: source=packages/trading/stockos_trading/costs.py commit=fcad869ecc531a6de80777f59b6fb842a2891101 — verbatim copy, do not edit
"""Fee + transaction-tax / regulatory-fee model for cash equities.

Default rates mirror the KIS retail commission schedule and the FSC
sell-side transaction tax in effect at the time of writing
(2026-05-03). They are conservative point-estimates — operators with
audited account-specific rate sheets should override via env so the
production calculation matches what KIS actually deducts. R3.1 Track
C wires that env-override seam without touching the broader pricing /
order / PnL chain (Decimal migration is a separate plan).

Env override keys (all numeric, decimal in [0, ``MAX_RATE``]):

* ``KIS_FEE_BUY_COMMISSION_RATE`` → ``buy_commission_rate``
* ``KIS_FEE_SELL_COMMISSION_RATE`` → ``sell_commission_rate``
* ``KIS_TAX_KOSPI_SELL_RATE`` → ``kospi_sell_tax_rate``
* ``KIS_TAX_KOSDAQ_SELL_RATE`` → ``kosdaq_sell_tax_rate``
* ``KIS_TAX_KONEX_SELL_RATE`` → ``konex_sell_tax_rate``
* ``KIS_US_FEE_BUY_COMMISSION_RATE`` → ``us_buy_commission_rate``
* ``KIS_US_FEE_SELL_COMMISSION_RATE`` → ``us_sell_commission_rate``
* ``KIS_US_SELL_REGULATORY_FEE_RATE`` → ``us_sell_regulatory_fee_rate``
* ``KIS_US_FX_CONVERSION_FEE_RATE`` → ``us_fx_conversion_fee_rate``

Validation rules (every override goes through the same gate):

1. Must parse as a non-NaN, non-Inf finite float.
2. Must be ``>= 0``. Negative rates would silently flip the sign of
   PnL and the break-even price calculation, masking a misconfigured
   env as a "free trade". Hard reject.
3. Must be ``<= MAX_RATE`` (default ``0.05`` = 5%). KRX retail
   commission is on the order of 0.015%; the FSC sell tax is 0.20%.
   Anything above 5% is almost certainly an off-by-100 typo
   (``0.20`` instead of ``0.0020``); we fail fast at startup rather
   than apply an absurdly hostile fee that would block every order.

Failure mode: any rejected override raises
:class:`stockos_trading.session_hours.ConfigurationError`. The API
lifespan loader (``apps/api/stockos_api/main.py``) lets the exception
propagate so uvicorn refuses to start with a malformed env, matching
the pattern PR-L4 established for ``SessionHoursConfig.from_env``.
"""

from __future__ import annotations

import math
from collections.abc import Mapping
from dataclasses import dataclass
from enum import Enum

from .models import Side
from .session_hours import ConfigurationError


class Market(str, Enum):
    KOSPI = "KOSPI"
    KOSDAQ = "KOSDAQ"
    KONEX = "KONEX"
    NASD = "NASD"
    NYSE = "NYSE"
    AMEX = "AMEX"


US_MARKETS: frozenset[Market] = frozenset({Market.NASD, Market.NYSE, Market.AMEX})


# Conservative cap. KRX retail fee + FSC tax never realistically reach
# 5%; anything past this is treated as a typo (e.g. 20.0 instead of
# 0.20). Operators who legitimately need a higher cap should bump
# this constant in code via an ADR — env alone must not bypass it.
MAX_RATE: float = 0.05


@dataclass(frozen=True)
class FeeTaxModel:
    buy_commission_rate: float = 0.00015
    sell_commission_rate: float = 0.00015
    kospi_sell_tax_rate: float = 0.0020
    kosdaq_sell_tax_rate: float = 0.0020
    konex_sell_tax_rate: float = 0.0010
    # KIS overseas US online commission is account/promotion dependent.
    # The conservative default mirrors the public non-promo 0.25% shape
    # and is explicitly overrideable for accounts on a lower-rate event.
    us_buy_commission_rate: float = 0.0025
    us_sell_commission_rate: float = 0.0025
    # US SEC/TAF-like sell-side charges are small and broker-specific.
    # Keep default zero unless the operator has an audited rate sheet.
    us_sell_regulatory_fee_rate: float = 0.0
    # 환전 스프레드/수수료를 USD notional equivalent로 반영한다. KIS API가
    # 환전 주문 endpoint를 제공하지 않더라도 통합증거금/수동환전 경로에서
    # 비용은 발생할 수 있으므로, 계좌별 실제 우대율은 env로 보정한다.
    us_fx_conversion_fee_rate: float = 0.0

    def sell_tax_rate(self, market: Market) -> float:
        if market == Market.KOSPI:
            return self.kospi_sell_tax_rate
        if market == Market.KOSDAQ:
            return self.kosdaq_sell_tax_rate
        if market == Market.KONEX:
            return self.konex_sell_tax_rate
        if market in US_MARKETS:
            return self.us_sell_regulatory_fee_rate
        raise ValueError(f"unsupported market: {market}")

    def buy_commission_rate_for(self, market: Market) -> float:
        if market in US_MARKETS:
            return self.us_buy_commission_rate
        return self.buy_commission_rate

    def sell_commission_rate_for(self, market: Market) -> float:
        if market in US_MARKETS:
            return self.us_sell_commission_rate
        return self.sell_commission_rate

    def fx_conversion_fee_rate_for(self, market: Market) -> float:
        if market in US_MARKETS:
            return self.us_fx_conversion_fee_rate
        return 0.0

    @classmethod
    def from_env(cls, env: Mapping[str, str]) -> "FeeTaxModel":
        """Build a :class:`FeeTaxModel` from env, validating every override.

        Keys absent from ``env`` keep the dataclass default. Keys
        present go through :func:`_parse_rate`, which fails closed on
        non-numeric, NaN/Inf, negative, or above-cap values. The
        per-key default is read from a fresh ``cls()`` so the source
        of truth stays the dataclass field declarations above (no
        duplicated literal between this method and the dataclass).
        """

        defaults = cls()
        return cls(
            buy_commission_rate=_resolve_rate(
                env, "KIS_FEE_BUY_COMMISSION_RATE", defaults.buy_commission_rate
            ),
            sell_commission_rate=_resolve_rate(
                env, "KIS_FEE_SELL_COMMISSION_RATE", defaults.sell_commission_rate
            ),
            kospi_sell_tax_rate=_resolve_rate(
                env, "KIS_TAX_KOSPI_SELL_RATE", defaults.kospi_sell_tax_rate
            ),
            kosdaq_sell_tax_rate=_resolve_rate(
                env, "KIS_TAX_KOSDAQ_SELL_RATE", defaults.kosdaq_sell_tax_rate
            ),
            konex_sell_tax_rate=_resolve_rate(
                env, "KIS_TAX_KONEX_SELL_RATE", defaults.konex_sell_tax_rate
            ),
            us_buy_commission_rate=_resolve_rate(
                env,
                "KIS_US_FEE_BUY_COMMISSION_RATE",
                defaults.us_buy_commission_rate,
            ),
            us_sell_commission_rate=_resolve_rate(
                env,
                "KIS_US_FEE_SELL_COMMISSION_RATE",
                defaults.us_sell_commission_rate,
            ),
            us_sell_regulatory_fee_rate=_resolve_rate(
                env,
                "KIS_US_SELL_REGULATORY_FEE_RATE",
                defaults.us_sell_regulatory_fee_rate,
            ),
            us_fx_conversion_fee_rate=_resolve_rate(
                env,
                "KIS_US_FX_CONVERSION_FEE_RATE",
                defaults.us_fx_conversion_fee_rate,
            ),
        )


def _resolve_rate(env: Mapping[str, str], key: str, default: float) -> float:
    """Return the validated env override or ``default`` when absent."""

    raw = env.get(key)
    if raw is None:
        return default
    return _parse_rate(raw, key=key)


def _parse_rate(value: str, *, key: str) -> float:
    """Parse a fee/tax rate string with bounded validation.

    See module docstring for the three validation rules. The error
    messages include the env key so operators can fix a malformed
    ``.env`` without re-reading source.
    """

    stripped = value.strip()
    if stripped == "":
        raise ConfigurationError(f"{key} must be a numeric rate (got empty string)")
    try:
        rate = float(stripped)
    except ValueError as exc:
        raise ConfigurationError(
            f"{key} must be a numeric rate (got {value!r})"
        ) from exc
    if not math.isfinite(rate):
        raise ConfigurationError(f"{key} must be finite (got {value!r})")
    if rate < 0:
        raise ConfigurationError(f"{key} must be non-negative (got {value!r})")
    if rate > MAX_RATE:
        raise ConfigurationError(
            f"{key} must be <= {MAX_RATE} (got {value!r}); "
            "rates above this cap are treated as typos — verify the value "
            "is a decimal fraction, not a percentage"
        )
    return rate


@dataclass(frozen=True)
class TradeCost:
    commission: float
    tax: float
    fx_fee: float = 0.0

    @property
    def total(self) -> float:
        return self.commission + self.tax + self.fx_fee


def estimate_trade_cost(
    notional: float,
    side: Side,
    market: Market,
    model: FeeTaxModel | None = None,
) -> TradeCost:
    if notional < 0:
        raise ValueError("notional must be non-negative")
    active_model = model or FeeTaxModel()
    commission_rate = (
        active_model.buy_commission_rate_for(market)
        if side == Side.BUY
        else active_model.sell_commission_rate_for(market)
    )
    commission = notional * commission_rate
    fx_fee = notional * active_model.fx_conversion_fee_rate_for(market)
    tax = notional * active_model.sell_tax_rate(market) if side == Side.SELL else 0.0
    return TradeCost(commission=commission, tax=tax, fx_fee=fx_fee)


def break_even_sell_price(
    buy_price: float,
    quantity: int,
    market: Market,
    model: FeeTaxModel | None = None,
    min_profit_krw: float = 0.0,
    usd_krw_rate: float | None = None,
) -> float:
    if buy_price <= 0:
        raise ValueError("buy_price must be positive")
    if quantity <= 0:
        raise ValueError("quantity must be positive")
    active_model = model or FeeTaxModel()
    buy_notional = buy_price * quantity
    buy_cost = estimate_trade_cost(buy_notional, Side.BUY, market, active_model).total
    sell_rate = active_model.sell_commission_rate_for(
        market
    ) + active_model.sell_tax_rate(market) + active_model.fx_conversion_fee_rate_for(
        market
    )
    required_profit = _native_profit_floor(
        min_profit_krw,
        market=market,
        usd_krw_rate=usd_krw_rate,
    )
    required_sell_notional = buy_notional + buy_cost + required_profit
    return required_sell_notional / (quantity * (1 - sell_rate))


def _native_profit_floor(
    min_profit_krw: float,
    *,
    market: Market,
    usd_krw_rate: float | None,
) -> float:
    if min_profit_krw == 0:
        return 0.0
    if market not in US_MARKETS:
        return min_profit_krw
    if usd_krw_rate is None or usd_krw_rate <= 0:
        raise ValueError(
            "usd_krw_rate must be positive when applying a KRW profit "
            "floor to a US market"
        )
    return min_profit_krw / usd_krw_rate


def net_pnl(
    buy_price: float,
    sell_price: float,
    quantity: int,
    market: Market,
    model: FeeTaxModel | None = None,
) -> float:
    if quantity <= 0:
        raise ValueError("quantity must be positive")
    buy_notional = buy_price * quantity
    sell_notional = sell_price * quantity
    active_model = model or FeeTaxModel()
    costs = estimate_trade_cost(buy_notional, Side.BUY, market, active_model).total
    costs += estimate_trade_cost(sell_notional, Side.SELL, market, active_model).total
    return sell_notional - buy_notional - costs


__all__ = [
    "FeeTaxModel",
    "MAX_RATE",
    "Market",
    "TradeCost",
    "US_MARKETS",
    "break_even_sell_price",
    "estimate_trade_cost",
    "net_pnl",
]
