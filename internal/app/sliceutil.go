package app

func diff(src, cmp []string) []string {
	diff := []string{}
	for _, s := range src {
		found := false
	cmp:
		for _, c := range cmp {
			if s == c {
				found = true
				break cmp
			}
		}
		if !found {
			diff = append(diff, s)
		}
	}

	return diff
}
