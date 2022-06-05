package app

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/kafka"
)

type KafkaAPI interface {
	ListClusters(ctx context.Context, params *kafka.ListClustersInput, optFns ...func(*kafka.Options)) (*kafka.ListClustersOutput, error)
	ListScramSecrets(ctx context.Context, params *kafka.ListScramSecretsInput, optFns ...func(*kafka.Options)) (*kafka.ListScramSecretsOutput, error)
	BatchAssociateScramSecret(ctx context.Context, params *kafka.BatchAssociateScramSecretInput, optFns ...func(*kafka.Options)) (*kafka.BatchAssociateScramSecretOutput, error)
	BatchDisassociateScramSecret(ctx context.Context, params *kafka.BatchDisassociateScramSecretInput, optFns ...func(*kafka.Options)) (*kafka.BatchDisassociateScramSecretOutput, error)
}

func listClusters(api KafkaAPI) (clusters []*Cluster, err error) {
	clusters = []*Cluster{}

	output, err := api.ListClusters(context.TODO(), &kafka.ListClustersInput{})
	if err != nil {
		return clusters, fmt.Errorf("unable to list clusters: %w", err)
	}

	for _, ci := range output.ClusterInfoList {
		ci := ci
		clusters = append(clusters, &Cluster{
			clusterInfo:              &ci,
			assosciatedSecretArnList: []string{},
			secretArnList:            []string{},
			secretArnChangeSet:       &SecretChangeSet{},
		})
	}

	return clusters, nil
}

func listScramSecrets(api KafkaAPI, cluster *Cluster) error {
	output, err := api.ListScramSecrets(context.TODO(), &kafka.ListScramSecretsInput{
		ClusterArn: cluster.clusterInfo.ClusterArn,
	})
	if err != nil {
		return fmt.Errorf("unable to list scram secrets for %v: %w", cluster.clusterInfo.ClusterName, err)
	}

	cluster.assosciatedSecretArnList = append(cluster.assosciatedSecretArnList, output.SecretArnList...)

	return nil
}

func associateSecrets(api KafkaAPI, cluster *Cluster) error {
	out, err := api.BatchAssociateScramSecret(context.TODO(), &kafka.BatchAssociateScramSecretInput{
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

func disassociateSecrets(api KafkaAPI, cluster *Cluster) error {
	out, err := api.BatchDisassociateScramSecret(context.TODO(), &kafka.BatchDisassociateScramSecretInput{
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
