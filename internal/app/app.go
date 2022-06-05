package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/theckman/yacspin"

	secretsmanagertypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"

	"golang.org/x/sync/errgroup"
)

func Run() error {
	svc, err := newService()
	if err != nil {
		return fmt.Errorf("unable to create new service: %w", err)
	}
	spinner := newSpinner()

	fmt.Println("Bind secrets to AWS MSK clusters.")
	fmt.Println()

	spinner.Start()
	spinner.Message("list kafka clusters and secretsmanager secrets")
	listClustersSecrets(svc)

	spinner.Message("list scram secrets")
	listScramSecretsByCluster(svc, spinner)

	spinner.Suffix(" retrieved data")
	spinner.Stop()
	fmt.Println()

	for _, cluster := range svc.clusters {
		mapSecretsToClusters(cluster, svc.secrets)
		reconcileClusterSecrets(cluster)
	}

	svc.printOverview()
	svc.printChangeSet()

	// fmt.Println("Press enter to apply changes.")
	// fmt.Scanln()
	//
	// spinner.Suffix(" modifying clusters")
	// spinner.Start()
	// updateClustersSecrets(svc, spinner)
	// spinner.Stop()

	return nil
}

func newService() (svc *Service, err error) {
	config, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to create aws config: %w", err)
	}

	return &Service{
		kafka:          kafka.NewFromConfig(config),
		secretsmanager: secretsmanager.NewFromConfig(config),
		clusters:       []*Cluster{},
	}, nil
}

func listClustersSecrets(svc *Service) error {
	g := new(errgroup.Group)

	g.Go(func() error {
		clusters, err := listClusters(svc.kafka)
		if err != nil {
			return fmt.Errorf("unable to list clusters: %w", err)
		}
		svc.clusters = clusters
		return nil
	})

	g.Go(func() error {
		secrets, err := listSecrets(svc.secretsmanager)
		if err != nil {
			return fmt.Errorf("unable to list secrets: %w", err)
		}
		svc.secrets = secrets
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("unable to list clusters and secrets: %w", err)
	}

	return nil
}

func listScramSecretsByCluster(svc *Service, spinner *yacspin.Spinner) error {
	g := new(errgroup.Group)

	clusterName := make(chan string, len(svc.clusters))

	g.Go(func() error {
		format := "list scram secrets [%v/%v] - %v"
		watchChan(clusterName, format, spinner)
		return nil
	})

	for _, cluster := range svc.clusters {
		cluster := cluster
		g.Go(func() error {
			if err := listScramSecrets(svc.kafka, cluster); err != nil {
				return fmt.Errorf("unable to list scram secrets: %w", err)
			}
			clusterName <- aws.ToString(cluster.clusterInfo.ClusterName)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("unable to list scram secrets: %w", err)
	}

	return nil
}

func updateClustersSecrets(svc *Service, spinner *yacspin.Spinner) error {
	for _, cluster := range svc.clusters {
		name := aws.ToString(cluster.clusterInfo.ClusterName)
		spinner.Message(fmt.Sprintf("updating scram secrets [%v]", name))
		if err := associateSecrets(svc.kafka, cluster); err != nil {
			return fmt.Errorf("unable to assosciate secrets for %v: %w", name, err)
		}
		if err := disassociateSecrets(svc.kafka, cluster); err != nil {
			return fmt.Errorf("unable to disassosciate secrets for %v: %w", name, err)
		}
	}

	return nil
}

func mapSecretsToClusters(cluster *Cluster, secrets []secretsmanagertypes.SecretListEntry) error {
	for _, secret := range secrets {
		if isClusterSecret(cluster.clusterInfo.ClusterName, secret.Tags) {
			cluster.secretArnList = append(cluster.secretArnList, aws.ToString(secret.ARN))
			continue
		}
	}

	return nil
}

func reconcileClusterSecrets(cluster *Cluster) error {
	add := diff(cluster.secretArnList, cluster.assosciatedSecretArnList)
	remove := diff(cluster.assosciatedSecretArnList, cluster.secretArnList)
	cluster.secretArnChangeSet.add = append(cluster.secretArnChangeSet.add, add...)
	cluster.secretArnChangeSet.remove = append(cluster.secretArnChangeSet.remove, remove...)

	return nil
}

func isClusterSecret(name *string, tags []secretsmanagertypes.Tag) bool {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == "Cluster" {
			if strings.HasPrefix(aws.ToString(name), aws.ToString(tag.Value)) {
				return true
			}
		}
	}

	return false
}
