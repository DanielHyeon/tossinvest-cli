import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# 페이지가 부재 신호를 두 번 묻는다(부작용 두 배).
sub(d + '/internal/console/strategy_runtime_multimarket.go',
    '\tif !absent {',
    '\tif !strategyprojection.StrategyRuntimeAbsent(c.opts.StrategyRuntime) {')
