package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type RegistrationServer struct {
	Addr string
}

func NewDefaultRegistrationServer() *RegistrationServer {
	return &RegistrationServer{
		Addr: "192.168.0.102:4000",
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
