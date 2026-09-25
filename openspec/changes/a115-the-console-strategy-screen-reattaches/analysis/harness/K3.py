import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# 화면 nil 판정 복귀 — 전략 페이지.
sub(d + '/internal/console/strategy_runtime_multimarket.go',
    '\tabsent := strategyprojection.StrategyRuntimeAbsent(c.opts.StrategyRuntime)',
    '\tabsent := c.opts.StrategyRuntime == nil')
