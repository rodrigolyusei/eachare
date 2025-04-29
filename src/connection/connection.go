package connection

import (
	"bufio"
	"net"
	"strconv"
	"strings"
	"sync"

	"EACHare/src/clock"
	"EACHare/src/commands/message"
	"EACHare/src/commands/request"
	"EACHare/src/commands/response"
	"EACHare/src/logger"
	"EACHare/src/peers"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Função para lidar com a conexão recebida
func ReceiveMessage(conn net.Conn, knownPeers *sync.Map, requestClient request.RequestClient, waitingCli bool) {
	// defer(adia) a função de fechamento da conexão quando as operações terminarem
	defer conn.Close()

	// Se a CLI está esperando por uma entrada, imprime nova linha para formatação
	if waitingCli {
		logger.Std("\n\n")
	}

	// Lê a mensagem recebida no buffer até encontrar \n
	msg, err := bufio.NewReader(conn).ReadString('\n')
	check(err)

	// Recupera as partes da mensagem
	msg = strings.TrimSuffix(msg, "\n")
	msgParts := strings.Split(msg, " ")

	// Imprime a mensagem/resposta recebida e atualiza o clock
	if msgParts[2] == "PEERS_LIST" {
		logger.Info("\tResposta recebida: \"" + msg + "\"")
	} else {
		logger.Info("\tMensagem recebida: \"" + msg + "\"")
	}

	// Guarda o valor do clock da mensagem recebida
	receivedClock, err := strconv.Atoi(msgParts[1])
	check(err)

	// Atualiza o relógio local comparando o valor local e recebido
	clock.UpdateMaxClock(receivedClock)

	// Monta a mensagem recebida
	receivedMessage := message.BaseMessage{
		Origin:    msgParts[0],
		Clock:     receivedClock,
		Type:      message.GetMessageType(msgParts[2]),
		Arguments: msgParts[3:],
	}

	// Verifica se a mensagem recebida é de um peer conhecido
	neighbor, exists := knownPeers.Load(receivedMessage.Origin)
	if exists {
		neighborStatus := neighbor.(peers.Peer).Status
		neighborClock := neighbor.(peers.Peer).Clock

		// Atualiza o status do peer conhecido com o maior clock
		if receivedClock > neighborClock {
			knownPeers.Store(receivedMessage.Origin, peers.Peer{Status: peers.ONLINE, Clock: receivedClock})
		} else {
			knownPeers.Store(receivedMessage.Origin, peers.Peer{Status: peers.ONLINE, Clock: neighborClock})
		}

		// Mostra mensagem de atualização apenas se for de peer OFFLINE e não for uma mensagem de BYE
		if neighborStatus == peers.OFFLINE && receivedMessage.Type != message.BYE {
			logger.Info("\tAtualizando peer " + receivedMessage.Origin + " status " + peers.ONLINE.String())
		}
	} else {
		knownPeers.Store(receivedMessage.Origin, peers.Peer{Status: peers.ONLINE, Clock: receivedClock})
		logger.Info("\tAdicionando novo peer " + receivedMessage.Origin + " status " + peers.ONLINE.String())
	}

	// Lida o comando recebido de acordo com o tipo de mensagem
	neighbor, _ = knownPeers.Load(receivedMessage.Origin)
	switch receivedMessage.Type {
	case message.GET_PEERS:
		response.GetPeersResponse(receivedMessage, knownPeers, conn, requestClient)
	case message.PEERS_LIST:
		response.PeersListResponse(receivedMessage, knownPeers)
	case message.BYE:
		response.ByeResponse(receivedMessage, knownPeers, neighbor.(peers.Peer).Clock)
	}

	// Verifica se a CLI está esperando por uma entrada
	if waitingCli {
		logger.Std("\n> ")
	}
}
