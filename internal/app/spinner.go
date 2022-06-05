package app

import (
	"fmt"
	"time"

	"github.com/theckman/yacspin"
)

func newSpinner() *yacspin.Spinner {
	cfg := yacspin.Config{
		Frequency:       50 * time.Millisecond,
		CharSet:         yacspin.CharSets[14],
		Suffix:          " retrieving data",
		StopCharacter:   "✓",
		SuffixAutoColon: true,
		StopColors:      []string{"fgGreen"},
	}

	spinner, _ := yacspin.New(cfg)
	return spinner
}

func watchChan(msg chan string, format string, spinner *yacspin.Spinner) {
	i := 0
	for {
		select {
		case event := <-msg:
			i++
			spinner.Message(fmt.Sprintf(format, i, cap(msg), event))
			if i == cap(msg) {
				return
			}
		}
	}
}
