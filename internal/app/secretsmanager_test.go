package app

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

type mockSecretsManagerClientAPI struct {
	listSecretsOutput []*secretsmanager.ListSecretsOutput
	err               error
}

func (m mockSecretsManagerClientAPI) ListSecrets(ctx context.Context, input *secretsmanager.ListSecretsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error) {
	var page int
	var nextToken *string

	if input.NextToken != nil {
		page, _ = strconv.Atoi(aws.ToString(input.NextToken))
	}

	if page < len(m.listSecretsOutput)-1 {
		nextToken = aws.String(strconv.Itoa(page + 1))
	} else {
		nextToken = nil
	}

	return &secretsmanager.ListSecretsOutput{
		NextToken:  nextToken,
		SecretList: m.listSecretsOutput[page].SecretList,
	}, m.err
}

func TestListSecrets(t *testing.T) {
	tests := []struct {
		name string
		give []*secretsmanager.ListSecretsOutput
		want []types.SecretListEntry
		err  error
	}{
		{
			name: "one",
			give: []*secretsmanager.ListSecretsOutput{
				{
					SecretList: []types.SecretListEntry{
						{ARN: aws.String("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-123456")},
						{ARN: aws.String("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-234567")},
					},
				},
			},
			want: []types.SecretListEntry{
				{ARN: aws.String("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-123456")},
				{ARN: aws.String("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-234567")},
			},
		}, {
			name: "many",
			give: []*secretsmanager.ListSecretsOutput{
				{
					SecretList: []types.SecretListEntry{
						{ARN: aws.String("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-123456")},
						{ARN: aws.String("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-234567")},
					},
				}, {
					SecretList: []types.SecretListEntry{
						{ARN: aws.String("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-345678")},
						{ARN: aws.String("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-456789")},
					},
				},
			},
			want: []types.SecretListEntry{
				{ARN: aws.String("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-123456")},
				{ARN: aws.String("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-234567")},
				{ARN: aws.String("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-345678")},
				{ARN: aws.String("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-456789")},
			},
		}, {
			name: "none",
			give: []*secretsmanager.ListSecretsOutput{
				{
					SecretList: []types.SecretListEntry{},
				},
			},
			want: []types.SecretListEntry{},
		}, {
			name: "error",
			give: []*secretsmanager.ListSecretsOutput{
				{
					SecretList: []types.SecretListEntry{},
				},
			},
			want: []types.SecretListEntry{},
			err:  errors.New("The security token included in the request is invalid."),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := func() *mockSecretsManagerClientAPI {
				return &mockSecretsManagerClientAPI{
					listSecretsOutput: tt.give,
					err:               tt.err,
				}
			}()

			got, err := listSecrets(cl)
			if !errors.Is(err, tt.err) {
				t.Errorf("got '%v', want '%v'", err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
