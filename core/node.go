package core

import (
	"fmt"
	"github.com/SINTEF-Infosec/demokit/hardware"
	"github.com/SINTEF-Infosec/demokit/media"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const APIAddr = ":8081"

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
	Info            NodeInfo
	Logger          *log.Entry
	actions         map[string]*Action
	entryPoint      *Action
	EventNetwork    EventNetwork
	Router          *gin.Engine
	Hardware        hardware.Hal
	MediaController media.MediaController
}

func newNode(info NodeInfo,
	logger *log.Entry,
	network EventNetwork,
	router *gin.Engine,
	mediaController media.MediaController,
	hal hardware.Hal) *Node {

	node := &Node{
		Info:            info,
		Logger:          logger,
		actions:         map[string]*Action{},
		EventNetwork:    network,
		Router:          router,
		Hardware:        hal,
		MediaController: mediaController,
	}

	// Ensuring required components are set
	if node.EventNetwork == nil {
		node.Logger.Fatalf("the event network is a mandatory component, but is nil")
	}

	if node.Router == nil {
		node.Logger.Fatalf("the router is a mandatory component, but is nil")
	}

	if node.Hardware == nil {
		node.Logger.Info("hardware not configured, using virtual hardware layer")
		node.Hardware = hardware.NewVirtualHardwareLayer()
	}

	if node.MediaController == nil {
		node.Logger.Info("media controller not configured, using virtual media controller instead")
		node.MediaController = media.NewVirtualMediaController()
	}

	// Bindings
	node.EventNetwork.SetReceivedEventCallback(node.handleEvent)

	return node
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

	n.StartAPIServer()
	n.EventNetwork.StartListeningForEvents()

	go func() {
		if n.entryPoint != nil {
			n.ExecuteAction(n.entryPoint, nil)
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

	n.Logger.Info("Node ready!")
	<-done
}

func (n *Node) StartAPIServer() {
	n.Logger.Info("Starting API Server...")
	go func() {
		if err := n.Router.Run(APIAddr); err != nil {
			n.Logger.Errorf("could not start API: %v", err)
		}
	}()
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

	n.ExecuteAction(action, event)
}

func (n *Node) ExecuteAction(action *Action, event *Event) {
	if action == nil {
		return
	}

	n.Logger.Debugf("Start executing action: %s", action.Name)
	if action.Do == nil {
		return
	}

	if action.DoDelay > 0 {
		n.Logger.Infof("Sleeping %d ms before pursuing", action.DoDelay)
		time.Sleep(time.Duration(action.DoDelay) * time.Millisecond)
	}

	if action.DoCondition != nil {
		if action.DoCondition(event) {
			action.Do(event)
		}
	} else {
		n.Logger.Debugf("Executing action: %s", action.Name)
		action.Do(event)
		n.Logger.Debugf("Action executed")
	}

	if action.Then != nil {
		n.Logger.Debugf("Execution next action: %s", action.Then.Name)
		n.ExecuteAction(action.Then, event)
	}
}

func (n *Node) BroadcastEvent(eventName, payload string) {
	event := &Event{
		Name:    eventName,
		Emitter: n.Info.Name,
		Payload: payload,
	}
	n.EventNetwork.BroadcastEvent(event)
}

func (n *Node) SendEventTo(receiver string, eventName, payload string) {
	event := &Event{
		Name:    eventName,
		Emitter: n.Info.Name,
		Payload: payload,
	}
	n.EventNetwork.SendEventTo(receiver, event)
}

func (n *Node) ServeState(state interface{}, allowEdit bool) {
	n.Router.GET("/state", func(c *gin.Context) {
		c.JSON(http.StatusOK, state)
	})

	if allowEdit {
		n.Router.PUT("/state", func(c *gin.Context) {
			if err := c.ShouldBind(state); err != nil {
				n.Logger.Errorf("error while binding state: %v", err)
				c.String(http.StatusBadRequest, "could not bind state")
				c.Abort()
				return
			}
			c.JSON(http.StatusOK, state)
		})
	}

	n.Logger.Infof("Node configured to serve its state on %s/state", APIAddr)
}
