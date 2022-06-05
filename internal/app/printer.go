package app

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/fatih/color"
	"github.com/rodaine/table"
)

func (svc *Service) printOverview() error {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()

	tbl := table.New("Cluster Name", "Version", "Assosciated Secrets", "Additions", "Removals")
	tbl.WithHeaderFormatter(headerFmt)

	for _, cluster := range svc.clusters {
		tbl.AddRow(
			aws.ToString(cluster.clusterInfo.ClusterName),
			aws.ToString(cluster.clusterInfo.CurrentBrokerSoftwareInfo.KafkaVersion),
			len(cluster.assosciatedSecretArnList),
			len(cluster.secretArnChangeSet.add),
			len(cluster.secretArnChangeSet.remove),
		)
	}
	tbl.Print()

	fmt.Println()

	return nil
}

func (svc *Service) printChangeSet() error {
	for _, cluster := range svc.clusters {
		c := len(cluster.secretArnChangeSet.add) + len(cluster.secretArnChangeSet.remove)
		if c > 0 {
			fmt.Println(aws.ToString(cluster.clusterInfo.ClusterName))
			fmt.Print(cluster.secretArnChangeSet)
			fmt.Println()
		}
	}

	return nil
}
