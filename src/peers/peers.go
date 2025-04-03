package peers

// Booleano para o status do peer
type PeerStatus bool

// Constantes para os status booleanos
const (
	ONLINE  PeerStatus = true
	OFFLINE PeerStatus = false
)

// Estrutura para armazenar informações do peer conhecido
type Peer struct {
	Address string
	Port    string
	Status  PeerStatus
}

// Função para obter o endereço completo a partir do Peer
func (peer Peer) FullAddress() string {
	return peer.Address + ":" + peer.Port
}

// Função para obter o estado do peer a partir do PeerStatus
func (status PeerStatus) String() string {
	if status {
		return "ONLINE"
	}
	return "OFFLINE"
}

// Função para obter o estado do peer a partir de uma string
func GetPeerStatus(status string) PeerStatus {
	if status == "ONLINE" {
		return ONLINE
	}
	return OFFLINE
}
