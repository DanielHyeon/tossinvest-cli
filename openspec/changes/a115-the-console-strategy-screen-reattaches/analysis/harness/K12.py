import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# 펌프가 콘솔 ctx 종료를 무시한다(goroutine 누수).
sub(d + '/cmd/tossctl/console_strategy_attach.go',
    '\t\tcase <-attachment.ctx.Done():\n\t\t\treturn\n', '')
