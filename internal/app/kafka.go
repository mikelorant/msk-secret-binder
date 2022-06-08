package app

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/aws/smithy-go"
)

type KafkaClientAPI interface {
	ListClusters(context.Context, *kafka.ListClustersInput, ...func(*kafka.Options)) (*kafka.ListClustersOutput, error)
	ListScramSecrets(context.Context, *kafka.ListScramSecretsInput, ...func(*kafka.Options)) (*kafka.ListScramSecretsOutput, error)
	BatchAssociateScramSecret(context.Context, *kafka.BatchAssociateScramSecretInput, ...func(*kafka.Options)) (*kafka.BatchAssociateScramSecretOutput, error)
	BatchDisassociateScramSecret(context.Context, *kafka.BatchDisassociateScramSecretInput, ...func(*kafka.Options)) (*kafka.BatchDisassociateScramSecretOutput, error)
}

func listClusters(cl KafkaClientAPI) (clusterInfo []types.ClusterInfo, err error) {
	clusterInfo = []types.ClusterInfo{}

	pagination := kafka.NewListClustersPaginator(cl, &kafka.ListClustersInput{})
	for pagination.HasMorePages() {
		output, err := pagination.NextPage(context.TODO())
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) {
				return clusterInfo, fmt.Errorf("unable to list clusters: %v", apiErr.ErrorMessage())
			}
			return clusterInfo, fmt.Errorf("unable to list clusters: %w", err)
		}
		clusterInfo = append(clusterInfo, output.ClusterInfoList...)
	}

	return clusterInfo, nil
}

func listScramSecrets(cl KafkaClientAPI, clusterArn *string) (secretArnList []string, err error) {
	secretArnList = []string{}

	options := func(o *kafka.ListScramSecretsPaginatorOptions) {
		o.Limit = 100
	}

	input := &kafka.ListScramSecretsInput{
		ClusterArn: clusterArn,
	}

	pagination := kafka.NewListScramSecretsPaginator(cl, input, options)
	for pagination.HasMorePages() {
		output, err := pagination.NextPage(context.TODO())
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) {
				return secretArnList, fmt.Errorf("unable to list scram secrets for %v: %v", aws.ToString(clusterArn), apiErr.ErrorMessage())
			}
			return secretArnList, fmt.Errorf("unable to list scram secrets: %w", err)
		}
		secretArnList = append(secretArnList, output.SecretArnList...)
	}

	return secretArnList, nil
}

func associateSecrets(cl KafkaClientAPI, cluster *Cluster) error {
	out, err := cl.BatchAssociateScramSecret(context.TODO(), &kafka.BatchAssociateScramSecretInput{
		ClusterArn:    cluster.clusterInfo.ClusterArn,
		SecretArnList: cluster.secretArnChangeSet.add,
	})
	if err != nil {
		return fmt.Errorf("unable to assosciate secrets: %w", err)
	}
	for _, v := range out.UnprocessedScramSecrets {
		log.Printf("unprocess scram secret: %v message: %v", v.SecretArn, v.ErrorMessage)
	}

	return nil
}

func disassociateSecrets(cl KafkaClientAPI, cluster *Cluster) error {
	out, err := cl.BatchDisassociateScramSecret(context.TODO(), &kafka.BatchDisassociateScramSecretInput{
		ClusterArn:    cluster.clusterInfo.ClusterArn,
		SecretArnList: cluster.secretArnChangeSet.add,
	})
	if err != nil {
		return fmt.Errorf("unable to disassosciate secrets: %w", err)
	}
	for _, v := range out.UnprocessedScramSecrets {
		log.Printf("unprocess scram secret: %v message: %v", v.SecretArn, v.ErrorMessage)
	}

	return nil
}
