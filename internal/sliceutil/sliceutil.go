package sliceutil

func Diff(src, cmp []string) []string {
	diff := []string{}
	for _, s := range src {
		found := false
		for _, c := range cmp {
			if s == c {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, s)
		}
	}

	return diff
}
