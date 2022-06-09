package app

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
)

type mockKafkaClientAPI struct {
	listClustersOutput     []*kafka.ListClustersOutput
	listScramSecretsOutput []*kafka.ListScramSecretsOutput
	err                    error
}

func (m mockKafkaClientAPI) ListClusters(ctx context.Context, params *kafka.ListClustersInput, optFns ...func(*kafka.Options)) (*kafka.ListClustersOutput, error) {
	var page int
	var nextToken *string

	if params.NextToken != nil {
		page, _ = strconv.Atoi(aws.ToString(params.NextToken))
	}

	if page < len(m.listClustersOutput)-1 {
		nextToken = aws.String(strconv.Itoa(page + 1))
	} else {
		nextToken = nil
	}

	return &kafka.ListClustersOutput{
		NextToken:       nextToken,
		ClusterInfoList: m.listClustersOutput[page].ClusterInfoList,
	}, m.err
}

func (m mockKafkaClientAPI) ListScramSecrets(ctx context.Context, params *kafka.ListScramSecretsInput, optFns ...func(*kafka.Options)) (*kafka.ListScramSecretsOutput, error) {
	var page int
	var nextToken *string

	if params.NextToken != nil {
		page, _ = strconv.Atoi(aws.ToString(params.NextToken))
	}

	if page < len(m.listScramSecretsOutput)-1 {
		nextToken = aws.String(strconv.Itoa(page + 1))
	} else {
		nextToken = nil
	}

	return &kafka.ListScramSecretsOutput{
		NextToken:     nextToken,
		SecretArnList: m.listScramSecretsOutput[page].SecretArnList,
	}, m.err
}

func (m mockKafkaClientAPI) BatchAssociateScramSecret(ctx context.Context, params *kafka.BatchAssociateScramSecretInput, optFns ...func(*kafka.Options)) (*kafka.BatchAssociateScramSecretOutput, error) {
	return nil, nil
}

func (m mockKafkaClientAPI) BatchDisassociateScramSecret(ctx context.Context, params *kafka.BatchDisassociateScramSecretInput, optFns ...func(*kafka.Options)) (*kafka.BatchDisassociateScramSecretOutput, error) {
	return nil, nil
}

func TestListClusters(t *testing.T) {
	tests := []struct {
		name string
		give []*kafka.ListClustersOutput
		want []types.ClusterInfo
		err  error
	}{
		{
			name: "one",
			give: []*kafka.ListClustersOutput{
				{
					ClusterInfoList: []types.ClusterInfo{
						{
							ClusterName: aws.String("example1"),
							ClusterArn:  aws.String("arn:aws:kafka:ap-southeast-2:123456789012:cluster/example1/1"),
						}, {
							ClusterName: aws.String("example2"),
							ClusterArn:  aws.String("arn:aws:kafka:ap-southeast-2:123456789012:cluster/example2/2"),
						},
					},
				},
			},
			want: []types.ClusterInfo{
				{
					ClusterName: aws.String("example1"),
					ClusterArn:  aws.String("arn:aws:kafka:ap-southeast-2:123456789012:cluster/example1/1"),
				}, {
					ClusterName: aws.String("example2"),
					ClusterArn:  aws.String("arn:aws:kafka:ap-southeast-2:123456789012:cluster/example2/2"),
				},
			},
		}, {
			name: "many",
			give: []*kafka.ListClustersOutput{
				{
					ClusterInfoList: []types.ClusterInfo{
						{
							ClusterName: aws.String("example1"),
							ClusterArn:  aws.String("arn:aws:kafka:ap-southeast-2:123456789012:cluster/example1/1"),
						}, {
							ClusterName: aws.String("example2"),
							ClusterArn:  aws.String("arn:aws:kafka:ap-southeast-2:123456789012:cluster/example2/2"),
						},
					},
				}, {
					ClusterInfoList: []types.ClusterInfo{
						{
							ClusterName: aws.String("example3"),
							ClusterArn:  aws.String("arn:aws:kafka:ap-southeast-2:123456789012:cluster/example3/3"),
						}, {
							ClusterName: aws.String("example4"),
							ClusterArn:  aws.String("arn:aws:kafka:ap-southeast-2:123456789012:cluster/example4/4"),
						},
					},
				},
			},
			want: []types.ClusterInfo{
				{
					ClusterName: aws.String("example1"),
					ClusterArn:  aws.String("arn:aws:kafka:ap-southeast-2:123456789012:cluster/example1/1"),
				}, {
					ClusterName: aws.String("example2"),
					ClusterArn:  aws.String("arn:aws:kafka:ap-southeast-2:123456789012:cluster/example2/2"),
				}, {
					ClusterName: aws.String("example3"),
					ClusterArn:  aws.String("arn:aws:kafka:ap-southeast-2:123456789012:cluster/example3/3"),
				}, {
					ClusterName: aws.String("example4"),
					ClusterArn:  aws.String("arn:aws:kafka:ap-southeast-2:123456789012:cluster/example4/4"),
				},
			},
		}, {
			name: "none",
			give: []*kafka.ListClustersOutput{
				{
					ClusterInfoList: []types.ClusterInfo{},
				},
			},
			want: []types.ClusterInfo{},
		}, {
			name: "error",
			give: []*kafka.ListClustersOutput{
				{
					ClusterInfoList: []types.ClusterInfo{},
				},
			},
			want: []types.ClusterInfo{},
			err:  errors.New("the security token included in the request is invalid"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := func() *mockKafkaClientAPI {
				return &mockKafkaClientAPI{
					listClustersOutput: tt.give,
					err:                tt.err,
				}
			}()

			got, err := listClusters(cl)
			if !errors.Is(err, tt.err) {
				t.Errorf("got '%v', want '%v'", err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestListScramSecrets(t *testing.T) {
	tests := []struct {
		name string
		give []*kafka.ListScramSecretsOutput
		want []string
		err  error
	}{
		{
			name: "one",
			give: []*kafka.ListScramSecretsOutput{
				{
					SecretArnList: []string{
						"arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example1",
						"arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example2",
					},
				},
			},
			want: []string{
				"arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example1",
				"arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example2",
			},
		}, {
			name: "many",
			give: []*kafka.ListScramSecretsOutput{
				{
					SecretArnList: []string{
						"arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example1",
						"arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example2",
					},
				}, {
					SecretArnList: []string{
						"arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example3",
						"arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example4",
					},
				},
			},
			want: []string{
				"arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example1",
				"arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example2",
				"arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example3",
				"arn:aws:secretsmanager:ap-southeast-2:123456789012:secret:AmazonMSK_example4",
			},
		}, {
			name: "none",
			give: []*kafka.ListScramSecretsOutput{
				{
					SecretArnList: []string{},
				},
			},
			want: []string{},
		}, {
			name: "error",
			give: []*kafka.ListScramSecretsOutput{
				{
					SecretArnList: []string{},
				},
			},
			want: []string{},
			err:  errors.New("the security token included in the request is invalid"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := func() *mockKafkaClientAPI {
				return &mockKafkaClientAPI{
					listScramSecretsOutput: tt.give,
					err:                    tt.err,
				}
			}()

			arn := aws.String("arn:aws:kafka:ap-southeast-2:123456789012:cluster/example1/1")

			got, err := listScramSecrets(cl, arn)
			if !errors.Is(err, tt.err) {
				t.Errorf("got '%v', want '%v'", err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
