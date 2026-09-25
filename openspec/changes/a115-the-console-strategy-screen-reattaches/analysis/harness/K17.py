import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# wrapper 의 presence 메서드 이름이 바뀐다 — a115 의 컴파일 결속(var _)이 잡아야 한다(BROKEN 이 의도된 포획).
# httpapi 쪽 기존 결속은 지워서 a115 결속 하나만 남긴다.
p = d + '/cmd/tossctl/httpapi_strategy_attach.go'
sub(p, 'func (a *strategyRuntimeAttachment) StrategyRuntimeConfigured() bool {',
    'func (a *strategyRuntimeAttachment) StrategyRuntimeIsConfigured() bool {')
sub(p, 'var _ httpapi.StrategyRuntimePresence = (*strategyRuntimeAttachment)(nil)\n', '')
