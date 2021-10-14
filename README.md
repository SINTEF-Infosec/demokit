# demokit

A kit to help creating demos, fast.

## Why the demokit?

The demokit has been created following the need for an easy way to create demonstrators out of projects.
IT projects are usually not visually attractive (who wants to see some log output from a console?). This kit provides
the building bricks to make demonstrators. It provides out of the box the developer with an event-based network and a 
basic node that interacts with it.

> :warning: this project is still at an early stage and is not stable. The API is often broken. 

## Example

Here is an example on how to use the demokit to create a simple node that will broadcast a "HELLO_WORLD" event every 2 seconds.

```go
package main

import (
	"github.com/SINTEF-Infosec/demokit/core"
	"github.com/sirupsen/logrus"
	"time"
)

func main() {
	node := NewHelloNode()
	node.Configure()
	node.Start()
}

type HelloNode struct {
	*core.Node
}

func NewHelloNode() *HelloNode {
	logrus.SetLevel(logrus.DebugLevel)
	return &HelloNode{
		Node: core.NewDefaultNode(),
	}
}

func (n *HelloNode) Configure() {
	n.SetEntryPoint(&core.Action{
		Name: "HelloWorld",
		Do: n.HelloWorld,

	})
}

func (n *HelloNode) HelloWorld(_ *core.Event) {
	for {
		n.Logger.Info("Broadcasting hello world...")
		n.BroadcastEvent("HELLO_WORLD", "")
		time.Sleep(2 * time.Second)
	}
}
```

More examples are available [here](https://github.com/SINTEF-Infosec/demokit-examples).

## Contributing

