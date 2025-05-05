package connection

// Pacotes nativos de go e pacotes internos
import (
	"bufio"
	"errors"
	"net"
	"strconv"
	"strings"

	"EACHare/src/clock"
	"EACHare/src/logger"
	"EACHare/src/message"
	"EACHare/src/peers"
)

// Função para verificar e imprimir mensagem de erro
func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Função para enviar mensagem
func SendMessage(knownPeers *peers.SafePeers, conn net.Conn, message message.BaseMessage, receiverAddress string) {
	// Atualiza o clock e mostra o encaminhamento
	message.Clock = clock.UpdateClock()
	logger.Info("Encaminhando mensagem \"" + message.String() + "\" para " + receiverAddress)

	// Tenta enviar a mensagem e verificar se há um erro
	var err error
	if conn == nil {
		err = errors.New("connection is nil")
	} else {
		_, err = conn.Write([]byte(message.String() + "\n"))
	}

	// Atualiza o peer e imprime mensagem apenas quando o status muda
	neighbor, _ := knownPeers.Get(receiverAddress)
	neighborStatus := neighbor.Status
	neighborClock := neighbor.Clock
	if err != nil && neighborStatus == peers.ONLINE {
		logger.Info("Atualizando peer " + receiverAddress + " status " + peers.OFFLINE.String())
		knownPeers.Add(peers.Peer{Address: receiverAddress, Status: peers.OFFLINE, Clock: neighborClock})
	} else if err == nil && neighborStatus == peers.OFFLINE {
		logger.Info("Atualizando peer " + receiverAddress + " status " + peers.ONLINE.String())
		knownPeers.Add(peers.Peer{Address: receiverAddress, Status: peers.ONLINE, Clock: neighborClock})
	}
}

// Função para lidar com a conexão recebida
func ReceiveMessage(knownPeers *peers.SafePeers, conn net.Conn) message.BaseMessage {
	// Lê a mensagem recebida no buffer até encontrar \n e constrói as partes da mensagem
	msg, err := bufio.NewReader(conn).ReadString('\n')
	check(err)
	msg = strings.TrimSuffix(msg, "\n")
	msgParts := strings.Split(msg, " ")

	// Cria variáveis para as partes da mensagem
	receivedAddress := msgParts[0]
	receivedClock, err := strconv.Atoi(msgParts[1])
	check(err)
	receivedMessageType := message.GetMessageType(msgParts[2])
	var receivedArguments []string
	if len(msgParts) > 3 {
		receivedArguments = msgParts[3:]
	}

	// Adiciona ou atualiza (apenas se for informação mais recente) o peer recebido
	neighbor, exists := knownPeers.Get(receivedAddress)
	if exists && neighbor.Clock > receivedClock {
		knownPeers.Add(peers.Peer{Address: receivedAddress, Status: peers.ONLINE, Clock: neighbor.Clock})
	} else {
		knownPeers.Add(peers.Peer{Address: receivedAddress, Status: peers.ONLINE, Clock: receivedClock})
	}

	// Retorna a mensagem recebida
	return message.BaseMessage{
		Origin:    receivedAddress,
		Clock:     receivedClock,
		Type:      receivedMessageType,
		Arguments: receivedArguments,
	}
}
