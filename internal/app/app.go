package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/aws/aws-sdk-go/service/secretsmanager"

	"github.com/theckman/yacspin"
)

func Run() error {
	fmt.Println("Bind secrets to AWS MSK clusters.")
	fmt.Println()

	config := aws.NewConfig().WithRegion("ap-southeast-2")
	session := session.Must(session.NewSession())

	svc := &Service{
		kafka:          kafka.New(session, config),
		secretsmanager: secretsmanager.New(session, config),
		clusters:       []*Cluster{},
	}

	spinner := newSpinner()

	spinner.Start()
	spinner.Message("kafka list clusters")
	if err := svc.listClusters(); err != nil {
		return fmt.Errorf("unable to list clusters: %w", err)
	}

	for _, cluster := range svc.clusters {
		name := aws.StringValue(cluster.clusterInfo.ClusterName)
		spinner.Message(fmt.Sprintf("list scram secrets [%v]", name))
		if err := svc.listScramSecrets(cluster); err != nil {
			return fmt.Errorf("unable to list scramsecrets for %v: %w", name, err)
		}
	}

	spinner.Message("secretsmanager list secrets")
	if err := svc.listSecrets(); err != nil {
		return fmt.Errorf("unable to list secrets: %w", err)
	}
	spinner.Suffix(" retrieved data")
	spinner.Stop()
	fmt.Println()

	svc.mapSecretsToClusters()
	svc.reconcileClusterSecrets()

	svc.printOverview()
	svc.printChangeSet()

	fmt.Println("Press enter to apply changes.")

	fmt.Scanln()

	// spinner.Suffix(" modifying clusters")
	// spinner.Start()
	// for _, cluster := range svc.clusters {
	// 	name := aws.StringValue(cluster.clusterInfo.ClusterName)
	// 	spinner.Message(fmt.Sprintf("updating scram secrets [%v]", name))
	// 	if err := svc.associateSecrets(cluster); err != nil {
	// 		return fmt.Errorf("unable to assosciate secrets for %v: %w", name, err)
	// 	}
	// 	if err := svc.disassociateSecrets(cluster); err != nil {
	// 		return fmt.Errorf("unable to disassosciate secrets for %v: %w", name, err)
	// 	}
	// }
	// spinner.Stop()

	return nil
}

func (svc *Service) mapSecretsToClusters() error {
	for _, cluster := range svc.clusters {
		sl := []*string{}
	secret:
		for _, secret := range svc.secrets {
			if isCluster(cluster.clusterInfo.ClusterName, secret.Tags) {
				sl = append(sl, secret.ARN)
				continue secret
			}
		}
		cluster.secretArnList = sl
	}

	return nil
}

func (svc *Service) reconcileClusterSecrets() error {
	for _, cluster := range svc.clusters {
		cluster.secretArnChangeSet = &SecretChangeSet{}
		add := diff(cluster.secretArnList, cluster.assosciatedSecretArnList)
		remove := diff(cluster.assosciatedSecretArnList, cluster.secretArnList)
		cluster.secretArnChangeSet.add = append(cluster.secretArnChangeSet.add, add...)
		cluster.secretArnChangeSet.remove = append(cluster.secretArnChangeSet.remove, remove...)
	}

	return nil
}

func isCluster(name *string, tags []*secretsmanager.Tag) bool {
	for _, tag := range tags {
		if aws.StringValue(tag.Key) == "Cluster" {
			if strings.HasPrefix(aws.StringValue(name), aws.StringValue(tag.Value)) {
				return true
			}
		}
	}

	return false
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
