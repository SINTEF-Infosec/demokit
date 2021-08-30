package core

import (
	"github.com/goombaio/namegenerator"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

// NewDefaultNode returns a Node with a default configuration. The only components available
// are the Logger (logrus.Entry), the EventNetwork and the router (gin.Engine).
func NewDefaultNode() *Node {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		seed := time.Now().UTC().UnixNano()
		nameGenerator := namegenerator.NewNameGenerator(seed)
		nodeName = nameGenerator.Generate()
	}

	info := NodeInfo{
		Name: nodeName,
	}

	logger := log.WithField("node", info.Name)

	rabbitMQEventNetwork := NewRabbitMQEventNetwork(ConnexionDetails{
		Username: getFromEnvOrFail("RABBIT_MQ_USERNAME", info.Name),
		Password: getFromEnvOrFail("RABBIT_MQ_PASSWORD", info.Name),
		Host:     getFromEnvOrFail("RABBIT_MQ_HOST", info.Name),
		Port:     getFromEnvOrFail("RABBIT_MQ_PORT", info.Name),
	}, logger)

	router := NewNodeRouter(logger)

	return NewNode(info, DefaultNodeConfig(), logger, NewDefaultRegistrationServer(), rabbitMQEventNetwork, router, nil, nil)
}

func DefaultNodeConfig() NodeConfig {
	return NodeConfig{
		ExposeActions: true,
	}
}
