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

	kafkatypes "github.com/aws/aws-sdk-go-v2/service/kafka/types"
	secretsmanagertypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"

	"github.com/mikelorant/msk-secret-binder/internal/sliceutil"
	"github.com/mikelorant/msk-secret-binder/internal/spinner"

	"golang.org/x/sync/errgroup"
)

func Run() error {
	svc, err := newService()
	if err != nil {
		return fmt.Errorf("unable to create new service: %w", err)
	}

	spin, err := spinner.NewSpinner()
	if err != nil {
		return fmt.Errorf("unable to create spinner: %w", err)
	}

	fmt.Println("Bind secrets to AWS MSK clusters.")
	fmt.Println()

	spin.Start()
	spin.Message("list kafka clusters and secretsmanager secrets")
	if err := listClustersSecrets(svc); err != nil {
		spin.StopFail()
		return err
	}

	spin.Message("list scram secrets")
	if err := listScramSecretsByCluster(svc, spin); err != nil {
		spin.StopFail()
		return err
	}

	spin.Suffix(" retrieved data")
	spin.Stop()
	fmt.Println()

	for _, cluster := range svc.clusters {
		mapSecretsToClusters(cluster, svc.secrets)
		reconcileClusterSecrets(cluster)
	}

	printOverview(svc.clusters)
	printChangeSet(svc.clusters)

	fmt.Println("Press enter to apply changes.")
	fmt.Scanln()

	spin.Suffix(" modifying clusters")
	spin.Start()
	updateClustersSecrets(svc, spin)
	spin.Stop()

	return nil
}

func newService() (svc *Service, err error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to create aws config: %w", err)
	}

	return &Service{
		kafka:          kafka.NewFromConfig(cfg),
		secretsmanager: secretsmanager.NewFromConfig(cfg),
	}, nil
}

func listClustersSecrets(svc *Service) error {
	clusterInfo := make(chan []kafkatypes.ClusterInfo, 1)
	secretListEntry := make(chan []secretsmanagertypes.SecretListEntry, 1)

	g := new(errgroup.Group)

	g.Go(func() error {
		ci, err := listClusters(svc.kafka)
		if err != nil {
			return fmt.Errorf("unable to list clusters: %w", err)
		}
		clusterInfo <- ci
		return nil
	})

	g.Go(func() error {
		secrets, err := listSecrets(svc.secretsmanager)
		if err != nil {
			return fmt.Errorf("unable to list secrets: %w", err)
		}
		secretListEntry <- secrets
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("unable to list clusters and secrets: %w", err)
	}

	for _, ci := range <-clusterInfo {
		ci := ci
		svc.clusters = append(svc.clusters, &Cluster{
			clusterInfo: &ci,
		})
	}

	svc.secrets = <-secretListEntry

	return nil
}

func listScramSecretsByCluster(svc *Service, spin *yacspin.Spinner) error {
	g := new(errgroup.Group)

	clusterName := make(chan string, len(svc.clusters))

	go func() {
		format := "list scram secrets [%v/%v] - %v"
		spinner.WatchChan(spin, clusterName, format)
	}()

	for _, cluster := range svc.clusters {
		cluster := cluster
		g.Go(func() error {
			scramSecrets, err := listScramSecrets(svc.kafka, cluster.clusterInfo.ClusterArn)
			if err != nil {
				return fmt.Errorf("unable to list scram secrets: %w", err)
			}
			cluster.assosciatedSecretArnList = scramSecrets
			clusterName <- aws.ToString(cluster.clusterInfo.ClusterName)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("unable to list scram secrets: %w", err)
	}

	close(clusterName)

	return nil
}

func updateClustersSecrets(svc *Service, spin *yacspin.Spinner) error {
	for _, cluster := range svc.clusters {
		name := aws.ToString(cluster.clusterInfo.ClusterName)
		spin.Message(fmt.Sprintf("updating scram secrets [%v]", name))
		if err := associateSecrets(svc.kafka, cluster); err != nil {
			return fmt.Errorf("unable to assosciate secrets for %v: %w", name, err)
		}
		// if err := disassociateSecrets(svc.kafka, cluster); err != nil {
		// 	return fmt.Errorf("unable to disassosciate secrets for %v: %w", name, err)
		// }
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
	add := sliceutil.Diff(cluster.secretArnList, cluster.assosciatedSecretArnList)
	// remove := sliceutil.Diff(cluster.assosciatedSecretArnList, cluster.secretArnList)
	remove := []string{}

	cluster.secretArnChangeSet = &SecretChangeSet{
		add:    add,
		remove: remove,
	}

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
