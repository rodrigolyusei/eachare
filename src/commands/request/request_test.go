package request

import (
	"EACHare/src/commands/message"
	"EACHare/src/logger"
	"EACHare/src/peers"
	"bytes"
	"strings"
	"sync"
	"testing"
)

var requestClient RequestClient

func init() {
	requestClient = RequestClient{Address: "localhost"}
}

func TestSendMessageArgumentsNilOK(t *testing.T) {
	conn := &mockConn{}
	message := message.BaseMessage{
		Origin:    requestClient.Address,
		Clock:     0,
		Type:      message.UNKNOWN,
		Arguments: nil,
	}
	var knownPeers sync.Map
	knownPeers.Store("127.0.0.1:9001", peers.Peer{Status: peers.ONLINE, Clock: 0})

	requestClient.sendMessage(conn, message, "127.0.0.1:9001", &knownPeers)

	if string(conn.data) != "localhost 1 UNKNOWN\n" {
		t.Fatalf("Expected %s, got %s", "localhost 1 UNKNOWN\n", string(conn.data))
	}
}

func TestSendMessageConnNil(t *testing.T) {
	message := message.BaseMessage{
		Origin:    requestClient.Address,
		Clock:     0,
		Type:      message.UNKNOWN,
		Arguments: nil,
	}
	var knownPeers sync.Map
	knownPeers.Store("127.0.0.1:9001", peers.Peer{Status: peers.ONLINE, Clock: 0})

	requestClient.sendMessage(nil, message, "127.0.0.1:9001", &knownPeers)

	neighbor, _ := knownPeers.Load("127.0.0.1:9001")
	neighborStatus := neighbor.(peers.Peer).Status
	if neighborStatus != peers.OFFLINE {
		t.Fatalf("Expected peer status to be OFFLINE, got %s", neighborStatus.String())
	}
}

func TestGetPeersRequest(t *testing.T) {
	var knownPeers sync.Map
	knownPeers.Store("127.0.0.1:9001", peers.Peer{Status: peers.ONLINE, Clock: 0})
	knownPeers.Store("127.0.0.2:9002", peers.Peer{Status: peers.OFFLINE, Clock: 0})

	conns := requestClient.GetPeersRequest(&knownPeers)

	if len(conns) != 0 {
		t.Fatalf("Expected no connections, got %d", len(conns))
	}

	knownPeers.Range(func(key, value any) bool {
		peerStatus := value.(peers.PeerStatus)
		if peerStatus {
			t.Errorf("Expected peer status to be false, got true")
		}

		return true
	})
}

func TestByeRequest(t *testing.T) {
	var knownPeers sync.Map
	knownPeers.Store("127.0.0.1:9001", peers.Peer{Status: peers.ONLINE, Clock: 0})
	knownPeers.Store("127.0.0.2:9002", peers.Peer{Status: peers.OFFLINE, Clock: 0})

	var buffer bytes.Buffer

	logger.SetOutput(&buffer)

	requestClient.ByeRequest(&knownPeers)

	out := buffer.String()
	expected := `Saindo...
	=> Atualizando relogio para 1
	Encaminhando mensagem "localhost 1 BYE" para 127.0.0.1:9001
	=> Atualizando relogio para 2
	Encaminhando mensagem "localhost 2 BYE" para 127.0.0.2:9002
	`
	if strings.TrimSpace(expected) != strings.TrimSpace(out) {
		t.Errorf("Expected %d \n%s, got %d \n%s", len(expected), expected, len(out), out)

	}
}

func TestPeerListRequest(t *testing.T) {
	var knownPeers sync.Map
	knownPeers.Store("127.0.0.2:9002", peers.Peer{Status: peers.OFFLINE, Clock: 0})

	receivedMessage := message.BaseMessage{
		Origin:    "127.0.0.1:9001",
		Clock:     1,
		Type:      message.GET_PEERS,
		Arguments: nil,
	}
	mockConn := &mockConn{}

	requestClient.PeersListRequest(mockConn, receivedMessage, &knownPeers)

	output := string(mockConn.data)
	expected := `localhost 1 PEERS_LIST 1 127.0.0.2:9002:OFFLINE:0`

	if strings.TrimSpace(output) != strings.TrimSpace(expected) {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, output)
	}
}

func TestHelloRequestOffline(t *testing.T) {
	var knownPeers sync.Map
	knownPeers.Store("127.0.0.2:9002", peers.Peer{Status: peers.OFFLINE, Clock: 0})

	client := RequestClient{Address: "localhost"}
	receiverAddress := "invalid-address:9999"

	client.HelloRequest(receiverAddress, &knownPeers)
}
