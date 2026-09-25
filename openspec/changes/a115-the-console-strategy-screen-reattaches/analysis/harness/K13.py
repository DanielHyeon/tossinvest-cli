import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# 판정 두 벌: httpapi 가 위임 대신 자기 사본을 둔다(오늘은 같은 답).
sub(d + '/internal/httpapi/strategy_runtime.go',
    '\treturn strategyprojection.StrategyRuntimeAbsent(reader)',
    '\tif reader == nil {\n\t\treturn true\n\t}\n\tif presence, ok := reader.(StrategyRuntimePresence); ok {\n\t\treturn !presence.StrategyRuntimeConfigured()\n\t}\n\treturn false')
