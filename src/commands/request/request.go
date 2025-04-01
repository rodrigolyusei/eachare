package request

// Pacotes nativos de go e pacotes internos
import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"EACHare/src/clock"
	"EACHare/src/commands/message"
	"EACHare/src/peers"
)

// Variável para o endereço do peer próprio
var Address string = "localhost"

// Função para enviar mensagem
func sendMessage(connection net.Conn, message message.BaseMessage, receiverAddress string) error {
	// Atualiza o clock antes de enviar mensagem
	message.Clock = clock.UpdateClock()

	// Cria a string do argumento da mensagem enviada
	arguments := ""
	if message.Arguments != nil {
		arguments = " " + strings.Join(message.Arguments, " ")
	}

	// Cria a string da mensagem inteira e imprime o encaminhamento
	messageStr := fmt.Sprintf("%s %d %s%s", Address, message.Clock, message.Type.String(), arguments)
	fmt.Printf("\tEncaminhando mensagem \"%s\" para %s\n", messageStr, receiverAddress)

	// Se a conexão é nula retorna um erro
	if connection == nil {
		return errors.New("connection is nil")
	}

	// Envia a mensagem pela conexão
	_, err := connection.Write([]byte(messageStr))
	return err
}

// Função para criar a mensagem GET_PEERS
func GetPeersRequest(knownPeers map[string]peers.PeerStatus) []net.Conn {
	// Cria um slice de conexões e a estrutura da mensagem GET_PEERS
	connections := make([]net.Conn, 0)
	baseMessage := message.BaseMessage{Clock: 0, Type: message.GET_PEERS, Arguments: nil}

	// Itera sobre os peers conhecidos
	for addressPort := range knownPeers {
		// Tenta estabelecer uma conexão com o peer
		conn, _ := net.Dial("tcp", addressPort)
		if conn != nil {
			connections = append(connections, conn)
			conn.SetDeadline(time.Now().Add(2 * time.Second))
		}

		// Envia a mensagem GET_PEERS para e considera diferentes casos
		err := sendMessage(conn, baseMessage, addressPort)
		if err != nil {
			// Se a conexão falhar e o peer estiver ONLINE, atualiza o status para OFFLINE
			if knownPeers[addressPort] == peers.ONLINE {
				fmt.Println("\tAtualizando peer " + addressPort + " status OFFLINE")
				knownPeers[addressPort] = peers.OFFLINE
			}
		} else {
			// Se a conexão for bem-sucedida e o peer estiver OFFLINE, atualiza o status para ONLINE
			if knownPeers[addressPort] == peers.OFFLINE {
				fmt.Println("\tAtualizando peer " + addressPort + " status ONLINE")
				knownPeers[addressPort] = peers.ONLINE
			}
		}
	}

	// Retorna as conexões estabelecidas para criar um receiver para cada um
	return connections
}

func ByeRequest(knownPeers map[string]peers.PeerStatus) {
	fmt.Println("Saindo...")
	baseMessage := message.BaseMessage{Origin: Address, Clock: 0, Type: message.BYE, Arguments: nil}

	for addressPort := range knownPeers {
		conn, err := net.Dial("tcp", addressPort)
		if err == nil {
			conn.SetDeadline(time.Now().Add(2 * time.Second))
			defer conn.Close()
		}
		err = sendMessage(conn, baseMessage, addressPort)
		if err != nil {
			continue
		}
	}

	os.Exit(0)
}

func PeerListRequest(conn net.Conn, receivedMessage message.BaseMessage, knownPeers map[string]peers.PeerStatus) {
	peers := []string{}

	size := 0
	for addressPort, peerStatus := range knownPeers {
		if addressPort == receivedMessage.Origin {
			continue
		}
		size++
		peers = append(peers, addressPort+":"+peerStatus.String()+":"+"0")
	}

	arguments := append([]string{strconv.Itoa(size)}, peers...)

	dropMessage := message.BaseMessage{Origin: Address, Clock: 0, Type: message.PEERS_LIST, Arguments: arguments}

	sendMessage(conn, dropMessage, receivedMessage.Origin)
}

func HelloRequest(receiverAddress string) peers.PeerStatus {
	baseMessage := message.BaseMessage{Clock: 0, Type: message.HELLO, Arguments: nil}
	conn, _ := net.Dial("tcp", receiverAddress)
	if conn != nil {
		defer conn.Close()
	}
	err := sendMessage(conn, baseMessage, receiverAddress)
	if err != nil {
		return peers.OFFLINE
	}
	return peers.ONLINE
}
