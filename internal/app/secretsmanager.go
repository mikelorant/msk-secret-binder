package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/smithy-go"
)

type SecretsManagerClientAPI interface {
	ListSecrets(context.Context, *secretsmanager.ListSecretsInput, ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error)
}

const (
	_filterValue = "AmazonMSK_"
)

func listSecrets(cl SecretsManagerClientAPI) (secrets []types.SecretListEntry, err error) {
	secrets = []types.SecretListEntry{}

	filter := types.Filter{
		Key:    types.FilterNameStringTypeName,
		Values: []string{_filterValue},
	}

	input := &secretsmanager.ListSecretsInput{
		Filters: []types.Filter{filter},
	}

	options := func(o *secretsmanager.ListSecretsPaginatorOptions) {
		o.Limit = 100
	}

	pagination := secretsmanager.NewListSecretsPaginator(cl, input, options)
	for pagination.HasMorePages() {
		output, err := pagination.NextPage(context.TODO())
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) {
				return secrets, fmt.Errorf("unable to list secrets: %v", apiErr.ErrorMessage())
			}
			return secrets, fmt.Errorf("unable to list secrets: %w", err)
		}
		secrets = append(secrets, output.SecretList...)
	}

	return secrets, nil
}
