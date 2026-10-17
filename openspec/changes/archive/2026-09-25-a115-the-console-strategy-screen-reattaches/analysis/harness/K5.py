import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# 부재 신호 무시: 판정이 nil 만 본다.
sub(d + '/internal/strategyprojection/presence.go',
    '\tif presence, ok := reader.(StrategyRuntimePresence); ok {\n\t\treturn !presence.StrategyRuntimeConfigured()\n\t}',
    '')
