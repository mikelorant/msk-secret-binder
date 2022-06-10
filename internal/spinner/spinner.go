package spinner

import (
	"fmt"
	"time"

	"github.com/theckman/yacspin"
)

func NewSpinner() (s *yacspin.Spinner, err error) {
	return yacspin.New(yacspin.Config{
		Frequency:         50 * time.Millisecond,
		CharSet:           yacspin.CharSets[14],
		Suffix:            " retrieving data",
		StopCharacter:     "✓",
		StopFailCharacter: "✗",
		SuffixAutoColon:   true,
		StopColors:        []string{"fgGreen"},
	})
}

func WatchChan(s *yacspin.Spinner, msg chan string, format string) {
	i := 0
	for m := range msg {
		i++
		s.Message(fmt.Sprintf(format, i, cap(msg), m))
	}
}
