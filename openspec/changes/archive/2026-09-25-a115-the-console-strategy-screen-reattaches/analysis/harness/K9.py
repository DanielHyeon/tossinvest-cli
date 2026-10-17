import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# interval<=0 가드 제거.
sub(d + '/cmd/tossctl/console_strategy_attach.go',
    '\tif attachment.interval <= 0 {\n\t\treturn\n\t}\n', '')
