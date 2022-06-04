package app

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type Service struct {
	kafka          *kafka.Kafka
	secretsmanager *secretsmanager.SecretsManager
	clusters       []*Cluster
	secrets        []*secretsmanager.SecretListEntry
}

type Cluster struct {
	clusterInfo              *kafka.ClusterInfo
	assosciatedSecretArnList []*string
	secretArnList            []*string
	secretArnChangeSet       *SecretChangeSet
}

type SecretChangeSet struct {
	add    []*string
	remove []*string
}

func (s SecretChangeSet) String() string {
	var str strings.Builder
	for _, v := range s.add {
		fmt.Fprintf(&str, "+%v\n", aws.StringValue(v))
	}
	for _, v := range s.remove {
		fmt.Fprintf(&str, "-%v\n", aws.StringValue(v))
	}
	return str.String()
}
