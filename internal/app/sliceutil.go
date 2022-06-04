package app

import "github.com/aws/aws-sdk-go/aws"

func diff(slice1, slice2 []*string) []*string {
	var diff []*string
	for _, s1 := range slice1 {
		found := false
	s2:
		for _, s2 := range slice2 {
			if aws.StringValue(s1) == aws.StringValue(s2) {
				found = true
				break s2
			}
		}
		if !found {
			diff = append(diff, s1)
		}
	}

	return diff
}
