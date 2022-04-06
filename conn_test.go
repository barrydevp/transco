package transco_test

import (
	"fmt"
	"testing"

	"github.com/barrydevp/transco"
)

const (
	TRANS_URI = "http://localhost:8001,localhost:8002,localhost:8003"
)

func TestConnectConn(t *testing.T) {
	conn, err := transco.NewConn(TRANS_URI)
	if err != nil {
		t.Error(err)
		return
	}

	err = conn.Connect()
	if err != nil {
		t.Error(err)
		return
	}

	nodes := conn.GetNodes()

	fmt.Println(conn.Leader().Ping())
	fmt.Println(len(nodes))

	for _, n := range nodes {
		fmt.Println(n.Ping())
	}

}
