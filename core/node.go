package core

import (
	"fmt"
	"github.com/goombaio/namegenerator"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

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
	Info       NodeInfo
	logger     *log.Entry
	actions    map[string]*Action
	entryPoint *Action
}

func NewNode() *Node {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		seed := time.Now().UTC().UnixNano()
		nameGenerator := namegenerator.NewNameGenerator(seed)
		nodeName = nameGenerator.Generate()
	}

	return &Node{
		Info: NodeInfo{
			Name: nodeName,
		},
		logger:  log.WithField("node", nodeName),
		actions: map[string]*Action{},
	}
}

func (n *Node) SetEntryPoint(action *Action) {
	n.entryPoint = action
}

func (n *Node) Start() {
	n.logger.Info("Starting node...")

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
		n.logger.Info("Stopping node...")
		done <- true
	}()

	<-done
}

// OnEventDo is used to register an action to execute when a given event is received.
func (n *Node) OnEventDo(eventName string, action *Action) {
	_, ok := n.actions[eventName]
	if ok {
		n.logger.Warnf("an action was already registered for the event %s, ignoring new assignation", eventName)
		return
	}

	n.actions[eventName] = action
	n.logger.Infof("action configured: %s -> %s", eventName, action.Name)
}

func (n *Node) handleEvent(event *Event) {
	if event.Emitter == n.Info.Name {
		return
	}

	action, ok := n.actions[event.Name]
	if !ok {
		n.logger.Debugf("no actions registered for event %s, ignoring", event.Name)
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

	_ = event

	// Delegating to the event manager
	// ToDo
}
