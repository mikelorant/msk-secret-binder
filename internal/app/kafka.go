package app

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/kafka"
)

func (svc *Service) listClusters() error {
	output, err := svc.kafka.ListClusters(&kafka.ListClustersInput{})
	if err != nil {
		return fmt.Errorf("unable to list clusters: %w", err)
	}

	for _, ci := range output.ClusterInfoList {
		svc.clusters = append(svc.clusters, &Cluster{
			clusterInfo:              ci,
			assosciatedSecretArnList: []*string{},
		})
	}

	return nil
}

func (svc *Service) listScramSecrets(cluster *Cluster) error {
	output, err := svc.kafka.ListScramSecrets(&kafka.ListScramSecretsInput{
		ClusterArn: cluster.clusterInfo.ClusterArn,
	})
	if err != nil {
		return fmt.Errorf("unable to list scram secrets for %v: %w", cluster.clusterInfo.ClusterName, err)
	}

	cluster.assosciatedSecretArnList = append(cluster.assosciatedSecretArnList, output.SecretArnList...)

	return nil
}

func (svc *Service) associateSecrets(cluster *Cluster) error {
	out, err := svc.kafka.BatchAssociateScramSecret(&kafka.BatchAssociateScramSecretInput{
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

func (svc *Service) disassociateSecrets(cluster *Cluster) error {
	out, err := svc.kafka.BatchDisassociateScramSecret(&kafka.BatchDisassociateScramSecretInput{
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
