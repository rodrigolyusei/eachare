package request

import (
	"EACHare/src/clock"
	"EACHare/src/commands/message"
	"EACHare/src/logger"
	"EACHare/src/peers"
	"bytes"
	"strings"
	"testing"
)

var requestClient RequestClient

func init() {
	requestClient = RequestClient{Address: "localhost"}
}

func TestSendMessageArgumentsNilOK(t *testing.T) {
	clock.ResetClock()
	conn := &mockConn{}
	message := message.BaseMessage{
		Origin:    requestClient.Address,
		Clock:     0,
		Type:      message.UNKNOWN,
		Arguments: nil,
	}

	err := requestClient.sendMessage(conn, message, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := "localhost 1 UNKNOWN"
	if string(conn.data) != expected {
		t.Fatalf("Expected %s, got %s", expected, string(conn.data))
	}
}

func TestSendMessageConnNil(t *testing.T) {
	clock.ResetClock()
	message := message.BaseMessage{
		Origin:    requestClient.Address,
		Clock:     0,
		Type:      message.UNKNOWN,
		Arguments: nil,
	}

	err := requestClient.sendMessage(nil, message, "")
	if err.Error() != "connection is nil" {
		t.Fatalf("Expected error, got %v", err)
	}
}

func TestSendMessageWriteError(t *testing.T) {
	clock.ResetClock()
	conn := &mockConn{}
	message := message.BaseMessage{
		Origin:    requestClient.Address + "testingWriteError",
		Clock:     0,
		Type:      message.UNKNOWN,
		Arguments: nil,
	}

	err := requestClient.sendMessage(conn, message, "")
	if err.Error() != "write error" {
		t.Fatalf("Expected error, got %v", err)
	}
}

func TestGetPeersRequest(t *testing.T) {
	knowPeers := make(map[string]peers.PeerStatus)
	knowPeers["127.0.0.1:8080"] = peers.ONLINE
	knowPeers["127.0.0.2:8081"] = peers.OFFLINE

	conns := requestClient.GetPeersRequest(knowPeers)

	if len(conns) != 0 {
		t.Fatalf("Expected no connections, got %d", len(conns))
	}

	for _, peerStatus := range knowPeers {
		if peerStatus {
			t.Errorf("Expected peer status to be false, got true")
		}
	}
}

func TestByeRequest(t *testing.T) {
	knowPeers := make(map[string]peers.PeerStatus)
	knowPeers["127.0.0.1:8080"] = peers.ONLINE
	knowPeers["127.0.0.2:8081"] = peers.OFFLINE

	var buffer bytes.Buffer
	var exit bool = false

	logger.SetOutput(&buffer)

	requestClient.ByeRequest(knowPeers, &exit)

	out := buffer.String()
	expected := `	Saindo...
	=> Atualizando relogio para 1
	Encaminhando mensagem "localhost 1 BYE" para 127.0.0.1:8080
	=> Atualizando relogio para 2
	Encaminhando mensagem "localhost 2 BYE" para 127.0.0.2:8081
	`
	if strings.TrimSpace(expected) != strings.TrimSpace(out) {
		t.Errorf("Expected %d \n%s, got %d \n%s", len(expected), expected, len(out), out)

	}
}

func TestPeerListRequest(t *testing.T) {
	knownPeers := map[string]peers.PeerStatus{
		"127.0.0.1:9001": peers.ONLINE,
		"127.0.0.2:9002": peers.OFFLINE,
	}

	receivedMessage := message.BaseMessage{
		Origin:    "127.0.0.1:9001",
		Clock:     1,
		Type:      message.GET_PEERS,
		Arguments: nil,
	}
	mockConn := &mockConn{}

	requestClient.PeersListRequest(mockConn, receivedMessage, knownPeers)

	output := string(mockConn.data)
	expected := `localhost 1 PEERS_LIST 1 127.0.0.2:9002:OFFLINE:0`

	if strings.TrimSpace(output) != strings.TrimSpace(expected) {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, output)
	}
}

func TestHelloRequestOffline(t *testing.T) {
	client := RequestClient{Address: "localhost"}
	receiverAddress := "invalid-address:9999"

	status := client.HelloRequest(receiverAddress)

	if status != peers.OFFLINE {
		t.Errorf("Expected peer to be OFFLINE, got %v", status)
	}
}
