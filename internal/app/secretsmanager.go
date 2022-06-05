package app

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

const (
	_filterKey   = "name"
	_filterValue = "AmazonMSK_"
)

func (svc *Service) listSecrets() error {
	filter := types.Filter{
		Key:    types.FilterNameStringTypeName,
		Values: []string{_filterValue},
	}

	var nextToken *string
	for {
		output, err := svc.secretsmanager.ListSecrets(context.TODO(), &secretsmanager.ListSecretsInput{
			Filters:   []types.Filter{filter},
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
