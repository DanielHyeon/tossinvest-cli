import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# (리뷰 R1) 틱 = 간격(절반이 아님).
sub(d + '/cmd/tossctl/console_strategy_attach.go',
    'time.NewTicker(max(attachment.interval/2, time.Millisecond))',
    'time.NewTicker(max(attachment.interval, time.Millisecond))')
