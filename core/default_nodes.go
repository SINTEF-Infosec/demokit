package core

import (
	log "github.com/sirupsen/logrus"
)

// NewDefaultNode returns a Node with a default configuration. The only components available
// are the Logger (logrus.Entry), the EventNetwork and the router (gin.Engine).
func NewDefaultNode() *Node {
	info := NodeInfo{} // Will default to a NODE_NAME or to a random name
	logger := log.NewEntry(log.New())

	rabbitMQEventNetwork := NewRabbitMQEventNetwork(ConnexionDetails{
		Username: getFromEnvOrFail("RABBIT_MQ_USERNAME", info.Name),
		Password: getFromEnvOrFail("RABBIT_MQ_PASSWORD", info.Name),
		Host:     getFromEnvOrFail("RABBIT_MQ_HOST", info.Name),
		Port:     getFromEnvOrFail("RABBIT_MQ_PORT", info.Name),
	})

	return NewNode(info, DefaultNodeConfig(), logger, NewDefaultRegistrationServer(), rabbitMQEventNetwork, nil, nil)
}

func DefaultNodeConfig() NodeConfig {
	return NodeConfig{
		ExposeActions: true,
	}
}
