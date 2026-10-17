import sys, os; sys.path.insert(0, os.path.dirname(__file__)); from apply import sub
d = sys.argv[1]
# 펌프가 시도 대상일 때만 깨운다(a109 G2 가 기각한 failed 게이트, freeze 리뷰 P1-2).
sub(d + '/cmd/tossctl/console_strategy_attach.go',
    '\t\tcase <-ticker.C:\n\t\t\tattachment.wake()',
    '\t\tcase <-ticker.C:\n\t\t\tif _, _, wanted := attachment.state(); wanted {\n\t\t\t\tattachment.wake()\n\t\t\t}')
