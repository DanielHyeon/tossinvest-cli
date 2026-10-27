package console

import (
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

type verifyStepItem struct {
	ID       verifylive.StepID
	Preamble string
	Header   string
	Body     string
}

func (p verifyPage) StepItems() []verifyStepItem {
	steps := verifylive.Steps()
	if p.Steps == "" || len(steps) == 0 {
		return nil
	}

	starts := make([]int, len(steps))
	from := 0
	for i, step := range steps {
		boundary := "\n" + fmt.Sprintf("%-22s ", step.ID)
		at := strings.Index(p.Steps[from:], boundary)
		if at < 0 {
			return nil
		}
		starts[i] = from + at + 1
		from = starts[i] + len(boundary) - 1
	}

	items := make([]verifyStepItem, 0, len(steps))
	for i, step := range steps {
		end := len(p.Steps)
		if i+1 < len(starts) {
			end = starts[i+1]
		}
		chunk := p.Steps[starts[i]:end]
		lineEnd := strings.IndexByte(chunk, '\n')
		if lineEnd < 0 || !strings.HasPrefix(chunk, string(step.ID)) {
			return nil
		}
		item := verifyStepItem{
			ID:     step.ID,
			Header: chunk[:lineEnd],
			Body:   chunk[lineEnd+1:],
		}
		if i == 0 {
			item.Preamble = p.Steps[:starts[i]]
		}
		items = append(items, item)
	}
	return items
}
