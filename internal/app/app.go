package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	secretsmanagertypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"

	"github.com/theckman/yacspin"

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
	spinner.Message("kafka list clusters and secretsmanager list secrets")

	g := new(errgroup.Group)

	g.Go(func() error {
		svc.clusters, err = listClusters(svc.kafka)
		if err != nil {
			return fmt.Errorf("unable to list clusters: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		svc.secrets, err = listSecrets(svc.secretsmanager)
		if err != nil {
			return fmt.Errorf("unable to list secrets: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("unable to list clusters and secrets: %w", err)
	}

	total := len(svc.clusters)
	name := make(chan string, total)
	g.Go(func() error {
		watchListScramSecrets(name, spinner)
		return nil
	})

	for _, cluster := range svc.clusters {
		cluster := cluster
		g.Go(func() error {
			if err := listScramSecrets(svc.kafka, cluster); err != nil {
				return fmt.Errorf("unable to list scram secrets: %w", err)
			}
			name <- aws.ToString(cluster.clusterInfo.ClusterName)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("unable to list scram secrets: %w", err)
	}

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
	//
	// fmt.Scanln()
	//
	// spinner.Suffix(" modifying clusters")
	// spinner.Start()
	// for _, cluster := range svc.clusters {
	// 	name := aws.StringValue(cluster.clusterInfo.ClusterName)
	// 	spinner.Message(fmt.Sprintf("updating scram secrets [%v]", name))
	// 	if err := svc.associateSecrets(svc.kafka, cluster); err != nil {
	// 		return fmt.Errorf("unable to assosciate secrets for %v: %w", name, err)
	// 	}
	// 	if err := svc.disassociateSecrets(svc.kafka, cluster); err != nil {
	// 		return fmt.Errorf("unable to disassosciate secrets for %v: %w", name, err)
	// 	}
	// }
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

func newSpinner() *yacspin.Spinner {
	cfg := yacspin.Config{
		Frequency:       50 * time.Millisecond,
		CharSet:         yacspin.CharSets[14],
		Suffix:          " retrieving data",
		StopCharacter:   "✓",
		SuffixAutoColon: true,
		StopColors:      []string{"fgGreen"},
	}

	spinner, _ := yacspin.New(cfg)
	return spinner
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

func watchListScramSecrets(name chan string, spinner *yacspin.Spinner) {
	i := 0
	for {
		select {
		case completed := <- name:
			i++
			spinner.Message(fmt.Sprintf("list scram secrets [%v/%v] - %v", i, cap(name), completed))
			if i == cap(name) {
				return
			}
		}
	}
}
