import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# 화면 nil 판정 복귀 — 설정 요약.
sub(d + '/internal/console/settings_tabs.go',
    '\tif strategyprojection.StrategyRuntimeAbsent(c.opts.StrategyRuntime) {',
    '\tif c.opts.StrategyRuntime == nil {')
