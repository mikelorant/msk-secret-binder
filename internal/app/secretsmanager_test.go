package app

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

type mockListSecretsAPI func(ctx context.Context, input *secretsmanager.ListSecretsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error)

func (m mockListSecretsAPI) ListSecrets(ctx context.Context, input *secretsmanager.ListSecretsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error) {
	return m(ctx, input, optFns...)
}

func TestListSecrets(t *testing.T) {
	tests := []struct {
		name   string
		client func(t *testing.T) SecretsManagerListSecretsAPI
		want   []types.SecretListEntry
	}{
		{
			name: "one",
			client: func(t *testing.T) SecretsManagerListSecretsAPI {
				return mockListSecretsAPI(func(ctx context.Context, input *secretsmanager.ListSecretsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error) {
					t.Helper()
					return &secretsmanager.ListSecretsOutput{
						SecretList: []types.SecretListEntry{
							types.SecretListEntry{ARN: aws.String("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-123456")},
						},
					}, nil
				})
			},
			want: []types.SecretListEntry{
				types.SecretListEntry{ARN: aws.String("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-123456")},
			},
		}, {
			name: "many",
			client: func(t *testing.T) SecretsManagerListSecretsAPI {
				return mockListSecretsAPI(func(ctx context.Context, input *secretsmanager.ListSecretsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error) {
					t.Helper()
					return &secretsmanager.ListSecretsOutput{
						SecretList: []types.SecretListEntry{
							types.SecretListEntry{ARN: aws.String("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-123456")},
							types.SecretListEntry{ARN: aws.String("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-234567")},
							types.SecretListEntry{ARN: aws.String("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-345678")},
						},
					}, nil
				})
			},
			want: []types.SecretListEntry{
				types.SecretListEntry{ARN: aws.String("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-123456")},
				types.SecretListEntry{ARN: aws.String("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-234567")},
				types.SecretListEntry{ARN: aws.String("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-345678")},
			},
		}, {
			name: "pagination",
			client: func(t *testing.T) SecretsManagerListSecretsAPI {
				return mockListSecretsAPI(func(ctx context.Context, input *secretsmanager.ListSecretsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error) {
					t.Helper()
					sl := []types.SecretListEntry{}
					for i := 1; i < 100; i++ {
						sl = append(sl, types.SecretListEntry{
							ARN: aws.String(fmt.Sprintf("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-%v", i)),
						})
					}
					return &secretsmanager.ListSecretsOutput{
						SecretList: sl,
					}, nil
				})
			},
			want: func() []types.SecretListEntry {
				sl := []types.SecretListEntry{}
				for i := 1; i < 100; i++ {
					sl = append(sl, types.SecretListEntry{
						ARN: aws.String(fmt.Sprintf("arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example-%v", i)),
					})
				}
				return sl
			}(),
		}, {
			name: "none",
			client: func(t *testing.T) SecretsManagerListSecretsAPI {
				return mockListSecretsAPI(func(ctx context.Context, input *secretsmanager.ListSecretsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error) {
					t.Helper()
					return &secretsmanager.ListSecretsOutput{
						SecretList: []types.SecretListEntry{},
					}, nil
				})
			},
			want: []types.SecretListEntry{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := listSecrets(tt.client(t))
			if err != nil {
				t.Fatalf("expect no error, got %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
