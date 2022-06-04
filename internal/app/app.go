package app

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/aws/aws-sdk-go/service/secretsmanager"

	"github.com/theckman/yacspin"
)

const (
	_region = "ap-southeast-2"
)

func Run() error {
	fmt.Println("Bind secrets to AWS MSK clusters.")
	fmt.Println()

	config := aws.NewConfig().WithRegion(_region)
	session := session.Must(session.NewSession())

	svc := &Service{
		kafka:          kafka.New(session, config),
		secretsmanager: secretsmanager.New(session, config),
		clusters:       []*Cluster{},
	}

	spinner := newSpinner()

	spinner.Start()
	spinner.Message("kafka list clusters and secretsmanager list secrets")

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		svc.listClusters()
	}()
	go func() {
		defer wg.Done()
		svc.listSecrets()
	}()
	wg.Wait()

	for _, cluster := range svc.clusters {
		name := aws.StringValue(cluster.clusterInfo.ClusterName)
		spinner.Message(fmt.Sprintf("list scram secrets [%v]", name))
		if err := svc.listScramSecrets(cluster); err != nil {
			return fmt.Errorf("unable to list scramsecrets for %v: %w", name, err)
		}
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

func mapSecretsToClusters(cluster *Cluster, secrets []*secretsmanager.SecretListEntry) error {
	sl := []*string{}
	for _, secret := range secrets {
		if isCluster(cluster.clusterInfo.ClusterName, secret.Tags) {
			sl = append(sl, secret.ARN)
			continue
		}
	}
	cluster.secretArnList = sl

	return nil
}

func reconcileClusterSecrets(cluster *Cluster) error {
	cluster.secretArnChangeSet = &SecretChangeSet{}
	add := diff(cluster.secretArnList, cluster.assosciatedSecretArnList)
	remove := diff(cluster.assosciatedSecretArnList, cluster.secretArnList)
	cluster.secretArnChangeSet.add = append(cluster.secretArnChangeSet.add, add...)
	cluster.secretArnChangeSet.remove = append(cluster.secretArnChangeSet.remove, remove...)

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
