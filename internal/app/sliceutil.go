package app

func diff(slice1, slice2 []string) []string {
	diff := []string{}
	for _, s1 := range slice1 {
		found := false
	slice2:
		for _, s2 := range slice2 {
			if s1 == s2 {
				found = true
				break slice2
			}
		}
		if !found {
			diff = append(diff, s1)
		}
	}

	return diff
}
