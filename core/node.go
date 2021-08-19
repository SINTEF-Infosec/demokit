package core

import (
	"encoding/json"
	"fmt"
	"github.com/goombaio/namegenerator"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const StateServerAddr = ":8081"

type NodeError string

func (ne *NodeError) Error() string {
	return fmt.Sprintf("node error: %s", ne)
}

type NodeInfo struct {
	Name string
}

// Node is the main component of the demokit. It aims to be a base for your own node and
// to provide an easy access to events happening in the network.
type Node struct {
	Info         NodeInfo
	Logger       *log.Entry
	actions      map[string]*Action
	entryPoint   *Action
	eventNetwork EventNetwork
}

func newNode(info NodeInfo, network EventNetwork, logger *log.Entry) *Node {
	node := &Node{
		Info:         info,
		Logger:       logger,
		actions:      map[string]*Action{},
		eventNetwork: network,
	}

	// Bindings
	node.eventNetwork.SetReceivedEventCallback(node.handleEvent)

	return node
}

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

	return newNode(info, rabbitMQEventNetwork, logger)
}

func getFromEnvOrFail(varName, nodeName string) string {
	envVar := os.Getenv(varName)
	if envVar == "" {
		log.WithField("node", nodeName).Fatalf("environment variable not set: %s", varName)
	}
	return envVar
}

func (n *Node) SetEntryPoint(action *Action) {
	n.entryPoint = action
}

func (n *Node) Start() {
	n.Logger.Info("Starting node...")

	n.eventNetwork.StartListeningForEvents()

	go func() {
		if n.entryPoint != nil {
			n.executeAction(n.entryPoint)
		}
	}()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		n.Logger.Info("Stopping node...")
		done <- true
	}()

	<-done
}

// OnEventDo is used to register an action to execute when a given event is received.
func (n *Node) OnEventDo(eventName string, action *Action) {
	_, ok := n.actions[eventName]
	if ok {
		n.Logger.Warnf("an action was already registered for the event %s, ignoring new assignation", eventName)
		return
	}

	n.actions[eventName] = action
	n.Logger.Infof("action configured: %s -> %s", eventName, action.Name)
}

func (n *Node) handleEvent(event *Event) {
	// Ignoring events sent by this node
	if event.Emitter == n.Info.Name {
		return
	}

	// Ignoring unicast events that are not for this node
	if event.Receiver != "*" && event.Receiver != n.Info.Name {
		return
	}

	action, ok := n.actions[event.Name]
	if !ok {
		n.Logger.Debugf("no actions registered for event %s, ignoring", event.Name)
		return
	}

	n.executeAction(action)
}

func (n *Node) executeAction(action *Action) {
	if action == nil {
		return
	}

	if action.Do == nil {
		return
	}

	if action.DoCondition != nil {
		if action.DoCondition() {
			action.Do()
		}
	} else {
		action.Do()
	}

	if action.Then != nil {
		n.executeAction(action.Then)
	}
}

func (n *Node) BroadcastEvent(eventName, payload string) {
	event := &Event{
		Name:    eventName,
		Emitter: n.Info.Name,
		Payload: payload,
	}
	n.eventNetwork.BroadcastEvent(event)
}

func (n *Node) SendEventTo(receiver string, eventName, payload string) {
	event := &Event{
		Name:    eventName,
		Emitter: n.Info.Name,
		Payload: payload,
	}
	n.eventNetwork.SendEventTo(receiver, event)
}

func (n *Node) ServeState(state interface{}) {
	http.HandleFunc("/state", func(w http.ResponseWriter, req *http.Request) {
		jsonData, err := json.Marshal(state)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - An error occurred while getting the node's state."))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	})
	go http.ListenAndServe(StateServerAddr, nil)
	n.Logger.Info("Serving state on")
}
