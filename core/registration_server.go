package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type RegisteredNode struct {
	NodeInfo        NodeInfo   `json:"info"`
	NodeStatus      NodeStatus `json:"status"`
	failUpdateCount int
}

type RegistrationServer struct {
	Addr string
}

func NewDefaultRegistrationServer(addr string) *RegistrationServer {
	return &RegistrationServer{
		Addr: addr,
	}
}

func (rs *RegistrationServer) RegisterNode(node *Node) error {
	// Preparing the data
	data, err := json.Marshal(node.Info)
	if err != nil {
		return fmt.Errorf("could not marshal node.Info: %v", err)
	}

	client := http.Client{
		Timeout: 2 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/register", rs.Addr), bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not do request: %v", err)
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("incorrect response code, expected 200, received %d", res.StatusCode)
	}

	return nil
}

func (rs *RegistrationServer) FetchNodesInfo() ([]RegisteredNode, error) {
	client := http.Client{
		Timeout: 2 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/nodes", rs.Addr), nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %v", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not do request: %v", err)
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("incorrect response code, expected 200, received %d", res.StatusCode)
	}

	// Getting the results
	// we expect a list of node info
	infos := make([]RegisteredNode, 0)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &infos); err != nil {
		return nil, err
	}

	return infos, nil
}
