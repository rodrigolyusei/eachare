package peers

// Booleano para o status do peer
type PeerStatus bool

// Constantes para os status booleanos
const (
	ONLINE  PeerStatus = true
	OFFLINE PeerStatus = false
)

// Estrutura para armazenar informações do peer conhecidos
type Peer struct {
	Address string
	Port    string
	Status  PeerStatus
}

// Função para obter o endereço completo do peer, diferente do peer próprio
func (peer Peer) FullAddress() string {
	return peer.Address + ":" + peer.Port
}

// Função para obter o estado do peer a partir do PeerStatus
func (s PeerStatus) String() string {
	if s {
		return "ONLINE"
	}
	return "OFFLINE"
}

// Função para obter o estado do peer a partir de uma string
func GetPeerStatus(s string) PeerStatus {
	switch s {
	case "ONLINE":
		return ONLINE
	}
	return OFFLINE
}
