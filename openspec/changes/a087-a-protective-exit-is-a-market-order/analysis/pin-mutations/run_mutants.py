#!/usr/bin/env python3
"""a087 tasks 4.1·4.2·4.3·4.5 고정 시험의 변이 하네스.

공유 워킹트리를 건드리지 않으려고 `go test -overlay` 로만 변이함 — 변이본은 임시 디렉터리에 쓰고, overlay 가
컴파일러에게만 그 파일을 대신 보여 줌(디스크의 원본은 그대로). 각 시험마다 먼저 「무변이 대조군」(원본과 같은 바이트를
overlay 로 넣음)이 PASS 여야 하고, 그다음 변이마다 FAIL 이어야 CAUGHT 임. 대조군이 PASS 가 아니면 멈춤.

사용 (저장소 루트에서):
    python3 -I openspec/changes/a087-a-protective-exit-is-a-market-order/analysis/pin-mutations/run_mutants.py
"""

from __future__ import annotations

import json
import subprocess
import sys
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[5]

# (시험 패키지, 시험 이름, 대상 파일, 원문, 변이문, 설명)
MUTANTS = [
    ("./internal/flatten/", "TestA087FlattenStillSellsAtAnAggressiveLimit", "internal/flatten/liquidate.go",
     'OrderType:    "limit",', 'OrderType:    "market",', "4.1 M1 flatten 청산 매도를 MARKET 으로"),
    ("./internal/flatten/", "TestA087FlattenStillSellsAtAnAggressiveLimit", "internal/flatten/liquidate.go",
     "err == nil && limits.LowerLimit > 0 {", "err == nil && limits.LowerLimit > 1e18 {",
     "4.1 M2 거래소 하한가 1순위를 건너뜀"),
    ("./internal/verifylive/", "TestA087VerifyPlacesOnlyFarLimitOrders", "internal/verifylive/mutate.go",
     'OrderType:     "limit",', 'OrderType:     "market",', "4.2 M3 검증 주문을 MARKET 으로"),
    ("./internal/verifylive/", "TestA087VerifyPlacesOnlyFarLimitOrders", "internal/verifylive/mutate.go",
     "Price:         s.Price,", "Price:         s.Price * 2,", "4.2 M4 가격을 시장 쪽으로 넘김(매수가 최근가 위)"),
    ("./internal/trading/", "TestA087ConditionalMarketStopBypassesThePlacePredicateUnchanged", "internal/trading/conditional.go",
     "return s.conditional.CreateConditionalOrder(ctx, intent)",
     'intent.OrderType = "LIMIT"\n\treturn s.conditional.CreateConditionalOrder(ctx, intent)',
     "4.3 M5 조건주문 의도를 변형(MARKET→LIMIT)"),
    ("./internal/trading/", "TestA087ConditionalMarketStopBypassesThePlacePredicateUnchanged", "internal/trading/conditional.go",
     "return s.conditional.CreateConditionalOrder(ctx, intent)",
     'if intent.OrderType == "MARKET" && !placeIntentSupported(orderintent.PlaceIntent{Market: "kr", Side: "sell", '
     'OrderType: "market", Quantity: intent.Quantity, CurrencyMode: "KRW"}) {\n'
     "\t\treturn domain.ConditionalOrderRef{}, ErrPlaceUnsupported\n\t}\n"
     "\treturn s.conditional.CreateConditionalOrder(ctx, intent)",
     "4.3 M6 조건주문이 일반 주문 관문 ③ 을 거치게 함"),
    ("./internal/execgw/", "TestA087KRMarketSellWireBodyCarriesNoPriceOrTIF", "internal/execgw/wirebody.go",
     'if body.OrderType == "LIMIT" {', 'if body.OrderType == "LIMIT" || intent.Price > 0 {',
     "4.5 M7 사본 렌더러가 시장가에 남은 가격을 실음"),
    ("./internal/execgw/", "TestA087KRMarketSellWireBodyCarriesNoPriceOrTIF", "internal/official/orders_write.go",
     'if orderType == "LIMIT" {', 'if orderType == "LIMIT" || intent.Price > 0 {',
     "4.5 M8 실제 빌더가 시장가에 남은 가격을 실음"),
    ("./internal/execgw/", "TestA087KRMarketSellWireBodyCarriesNoPriceOrTIF", "internal/execgw/wirebody.go",
     'if body.OrderType == "LIMIT" {', 'body.TimeInForce = "DAY"\n\tif body.OrderType == "LIMIT" {',
     "4.5 M9 사본 렌더러가 시장가에 TIF 를 실음"),
]


def go_test(package: str, test: str, overlay: dict[str, str], workdir: Path) -> bool:
    """overlay 를 걸고 한 시험을 돌려 PASS 여부를 돌려줌. 컴파일 실패도 FAIL 로 셈(출력은 보존)."""
    overlay_file = workdir / "overlay.json"
    overlay_file.write_text(json.dumps({"Replace": overlay}), encoding="utf-8")
    proc = subprocess.run(
        ["go", "test", "-count=1", f"-overlay={overlay_file}", "-run", f"^{test}$", package],
        cwd=ROOT, capture_output=True, text=True,
    )
    lines = (proc.stdout + proc.stderr).strip().splitlines()
    # 실패 사유가 단언(…_test.go:N)인지 컴파일 오류인지 가를 수 있게 마지막 사유 줄을 보여 줌 — 컴파일 실패는 CAUGHT 가 아님.
    # (마지막인 이유: 시험의 t.Logf 표본 줄이 단언보다 먼저 찍힘)
    reason = next((line.strip() for line in reversed(lines) if ".go:" in line), "")
    print("      " + (reason or lines[-1] if lines else "(출력 없음)"))
    if proc.returncode != 0 and "_test.go:" not in reason:
        print("      [경고] 실패 사유가 시험 단언이 아님 — 컴파일 오류일 수 있음")
    return proc.returncode == 0


def main() -> int:
    with tempfile.TemporaryDirectory(prefix="a087-pin-mutants-") as tmp:
        workdir = Path(tmp)
        controls: dict[tuple[str, str], bool] = {}
        caught = 0
        for index, (package, test, target, original, mutated, label) in enumerate(MUTANTS, 1):
            source_path = ROOT / target
            source = source_path.read_text(encoding="utf-8")
            if source.count(original) != 1:
                print(f"[STOP] {label}: 원문이 {source.count(original)}번 나옴 — 변이 자리가 모호함")
                return 2
            key = (package, test)
            if key not in controls:
                # 무변이 대조군: 원본과 같은 바이트를 overlay 로 넣어도 PASS 여야 overlay 경로가 정상임.
                same = workdir / f"control-{len(controls)}.go"
                same.write_text(source, encoding="utf-8")
                print(f"[control] {test} ({target} 원본 overlay)")
                controls[key] = go_test(package, test, {str(source_path): str(same)}, workdir)
                if not controls[key]:
                    print("[STOP] 무변이 대조군이 PASS 가 아님 — 변이 결과를 믿을 수 없음")
                    return 3
            mutant = workdir / f"mutant-{index}.go"
            mutant.write_text(source.replace(original, mutated, 1), encoding="utf-8")
            print(f"[M{index}] {label}")
            passed = go_test(package, test, {str(source_path): str(mutant)}, workdir)
            verdict = "SURVIVED" if passed else "CAUGHT"
            caught += 0 if passed else 1
            print(f"      → {verdict}")
        print(f"합계: CAUGHT {caught}/{len(MUTANTS)} · 대조군 PASS {sum(controls.values())}/{len(controls)}")
        return 0 if caught == len(MUTANTS) else 1


if __name__ == "__main__":
    sys.exit(main())
