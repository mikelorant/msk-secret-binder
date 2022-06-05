package app

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

type SecretsManagerListSecretsAPI interface {
	ListSecrets(ctx context.Context,
		input *secretsmanager.ListSecretsInput,
		optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error)
}

const (
	_filterValue = "AmazonMSK_"
)

func listSecrets(api SecretsManagerListSecretsAPI) (secrets []types.SecretListEntry, err error) {
	secrets = []types.SecretListEntry{}

	filter := types.Filter{
		Key:    types.FilterNameStringTypeName,
		Values: []string{_filterValue},
	}

	input := &secretsmanager.ListSecretsInput{
		Filters:   []types.Filter{filter},
	}

	for {
		output, err := api.ListSecrets(context.TODO(), input)
		if err != nil {
			return nil, fmt.Errorf("unable to list secrets: %w", err)
		}

		secrets = append(secrets, output.SecretList...)
		if output.NextToken == nil {
			break
		}

		input.NextToken = output.NextToken
	}

	return secrets, nil
}
