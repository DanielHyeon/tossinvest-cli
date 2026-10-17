# a140 CodeGraphContext supporting context

- **not-applicable**(proposal 단계): CodeGraphContext 는 advisory 이고 이 단계는 편집이 없다. 이 change 의 구현은 **새 파일만**
  (`internal/breakoutsignal/` · `tools/a140-breakout-signal-shadow/`)이라 기존 심볼의 누락 후보를 찾을 편집 표면이 없다.
  후보 확장은 CodeGraph `callers` 와 HEAD grep(`raw/grep-production-refs.txt`)으로 닫았다.
- 구현 착수(사람 결정 · READY 뒤)에서 새 패키지가 import 할 심볼(`officialbars.PollClosedBars` · `PollQuoteL1` ·
  `strategyevidence.Open` · `BarSeries` 읽기 · `breakoutlane.Evaluate`)을 대상으로 CGC 질의를 한 번 하고 이 파일을 갱신한다(tasks 0.5).
