package app

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/kafka"
	kafkatypes "github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	secretsmanagertypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

type Service struct {
	kafka          *kafka.Client
	secretsmanager *secretsmanager.Client
	clusters       []*Cluster
	secrets        []secretsmanagertypes.SecretListEntry
}

type Cluster struct {
	clusterInfo              *kafkatypes.ClusterInfo
	assosciatedSecretArnList []string
	secretArnList            []string
	secretArnChangeSet       *SecretChangeSet
}

type SecretChangeSet struct {
	add    []string
	remove []string
}

func (s SecretChangeSet) String() string {
	var str strings.Builder
	for _, v := range s.add {
		fmt.Fprintf(&str, "+%v\n", v)
	}
	for _, v := range s.remove {
		fmt.Fprintf(&str, "-%v\n", v)
	}
	return str.String()
}
