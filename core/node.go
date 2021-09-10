package core

import (
	"fmt"
	"github.com/SINTEF-Infosec/demokit/hardware"
	"github.com/SINTEF-Infosec/demokit/media"
	"github.com/gin-gonic/gin"
	"github.com/goombaio/namegenerator"
	log "github.com/sirupsen/logrus"
	"net"
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
	Name    string
	LocalIp string
}

type internalState struct {
	IsReady bool
}

type NodeCapabilities struct {
	MediaAvailable    bool `json:"media_available"`
	HardwareAvailable bool `json:"hardware_available"`
}

type NodeStatus struct {
	IsReady           bool                `json:"is_ready"`
	Capabilities      NodeCapabilities    `json:"capabilities"`
	RegisteredActions map[string][]string `json:"registered_actions"`
}

type NodeConfig struct {
	ExposeActions bool
}

// Node is the main component of the demokit. It aims to be a base for your own node and
// to provide an easy access to events happening in the network.
type Node struct {
	Info               NodeInfo
	Config             NodeConfig
	State              internalState
	Logger             *log.Entry
	actions            map[string]*Action
	entryPoint         *Action
	RegistrationServer *RegistrationServer
	EventNetwork       EventNetwork
	Router             *gin.Engine
	Hardware           hardware.Hal
	MediaController    media.MediaController
}

func NewNode(info NodeInfo,
	config NodeConfig,
	logger *log.Entry,
	rs *RegistrationServer,
	network EventNetwork,
	mediaController media.MediaController,
	hal hardware.Hal) *Node {

	node := &Node{
		Info:   info,
		Config: config,
		State: internalState{
			IsReady: false,
		},
		Logger:             logger,
		actions:            map[string]*Action{},
		RegistrationServer: rs,
		EventNetwork:       network,
		Router:             nil,
		Hardware:           hal,
		MediaController:    mediaController,
	}

	// If no node name is set, we check the env variables
	// If not set, we go for  random node name
	if node.Info.Name == "" {
		nodeName := os.Getenv("NODE_NAME")
		if nodeName == "" {
			seed := time.Now().UTC().UnixNano()
			nameGenerator := namegenerator.NewNameGenerator(seed)
			nodeName = nameGenerator.Generate()
		}
		node.Info.Name = nodeName
	}

	// Adding logger "node" field
	node.Logger = node.Logger.WithField("node", node.Info.Name)

	// Ensuring required components are set
	if node.EventNetwork == nil {
		node.Logger.Fatalf("the event network is a mandatory component, but is nil")
	}

	if node.Hardware == nil {
		node.Logger.Info("hardware not configured, using virtual hardware layer")
		node.Hardware = hardware.NewVirtualHardwareLayer()
	}

	if node.MediaController == nil {
		node.Logger.Info("media controller not configured, using virtual media controller instead")
		node.MediaController = media.NewVirtualMediaController()
	}

	node.Router = NewNodeRouter(node.Logger)

	// Setting logger for all the components
	node.EventNetwork.SetLogger(node.Logger)
	node.MediaController.SetLogger(node.Logger)
	node.Hardware.SetLogger(node.Logger)

	// Retrieving local info
	ip := node.RetrieveLocalIp()
	if ip != nil {
		node.Info.LocalIp = ip.String()
	}

	// Bindings
	node.Logger.Debug("Setting up event callback")
	node.EventNetwork.SetReceivedEventCallback(node.handleEvent)

	// Router configuration
	node.Logger.Debug("Enabling status")
	node.ServeStatus()

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
	n.Hardware.Init()

	n.StartAPIServer()
	n.EventNetwork.StartListeningForEvents()

	n.Register()

	n.Logger.Info("Node ready!")
	n.State.IsReady = true

	go func() {
		if n.entryPoint != nil {
			n.ExecuteAction(n.entryPoint, nil)
		}
	}()

	// Wait for exit will block until a SIGINT
	// or SIGTERM signal is received
	n.WaitForExit()
}

func (n *Node) WaitForExit() {
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

	n.Logger.Debugf("Start executing %s", action.Name)
	if action.Do == nil {
		return
	}

	if action.DoDelay > 0 {
		time.Sleep(time.Duration(action.DoDelay) * time.Millisecond)
	}

	if action.DoCondition != nil {
		if action.DoCondition(event) {
			action.Do(event)
		}
	} else {
		action.Do(event)
	}

	if action.Then != nil {
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

func (n *Node) ServeStatus() {
	n.Router.GET("/status", func(c *gin.Context) {
		var actions map[string][]string
		if n.Config.ExposeActions {
			actions = n.getRegisteredActions()
		}
		ns := NodeStatus{
			IsReady: n.State.IsReady,
			Capabilities: NodeCapabilities{
				HardwareAvailable: n.Hardware.IsAvailable(),
				MediaAvailable:    n.MediaController.IsAvailable(),
			},
			RegisteredActions: actions,
		}
		c.JSON(http.StatusOK, ns)
	})
}

func (n *Node) getRegisteredActions() map[string][]string {
	regActions := make(map[string][]string, len(n.actions))
	for event, action := range n.actions {
		regActions[event] = getActionsList(action, []string{})
	}
	return regActions
}

func getActionsList(action *Action, acc []string) []string {
	if action != nil {
		acc = append(acc, action.Name)
		acc = getActionsList(action.Then, acc)
	}
	return acc
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

func (n *Node) RetrieveLocalIp() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		n.Logger.Errorf("could not retrieve local IP: %v", err)
		return nil
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func (n *Node) Register() {
	if n.RegistrationServer == nil {
		n.Logger.Errorf("no registration server configured")
		return
	}
	if err := n.RegistrationServer.RegisterNode(n); err != nil {
		n.Logger.Errorf("could not register node: %v", err)
		return
	}
	log.Info("Successfully registered node against registration server")
}
