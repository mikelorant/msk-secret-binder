package app

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

const (
	_filterKey   = "name"
	_filterValue = "AmazonMSK_"
)

func (svc *Service) listSecrets() error {
	filter := &secretsmanager.Filter{
		Key: aws.String(_filterKey),
		Values: []*string{
			aws.String(_filterValue),
		},
	}

	var nextToken *string
	for {
		output, err := svc.secretsmanager.ListSecrets(&secretsmanager.ListSecretsInput{
			Filters:   []*secretsmanager.Filter{filter},
			NextToken: nextToken,
		})
		if err != nil {
			return fmt.Errorf("unable to list secrets: %w", err)
		}

		svc.secrets = append(svc.secrets, output.SecretList...)
		nextToken = output.NextToken

		if nextToken == nil {
			break
		}
	}

	return nil
}
