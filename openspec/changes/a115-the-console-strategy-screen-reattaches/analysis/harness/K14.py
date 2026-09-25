import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# presence 인터페이스 두 벌: alias 대신 새 정의.
sub(d + '/internal/httpapi/strategy_runtime.go',
    'type StrategyRuntimePresence = strategyprojection.StrategyRuntimePresence',
    'type StrategyRuntimePresence interface{ StrategyRuntimeConfigured() bool }')
