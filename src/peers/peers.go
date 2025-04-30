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
	Status PeerStatus
	Clock  int
}

// Função para obter o estado do peer a partir do PeerStatus
func (status PeerStatus) String() string {
	if status {
		return "ONLINE"
	}
	return "OFFLINE"
}

// Função para obter o estado do peer a partir de uma string
func GetStatus(status string) PeerStatus {
	if status == "ONLINE" {
		return ONLINE
	}
	return OFFLINE
}
