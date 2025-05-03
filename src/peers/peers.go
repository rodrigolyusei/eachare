package peers

// Pacotes nativos de go
import (
	"sort"
	"sync"
)

// Booleano para o status do peer
type PeerStatus bool

// Constantes para os status booleanos
const (
	OFFLINE PeerStatus = false
	ONLINE  PeerStatus = true
)

// Estrutura para armazenar informações do peer conhecido
type Peer struct {
	Address string
	Status  PeerStatus
	Clock   int
}

// Estrutura para armazenar a lista de peers de forma segura
type SafePeers struct {
	mutex sync.RWMutex
	peers []Peer
}

// Função para obter o estado do peer a partir do PeerStatus
func (status PeerStatus) String() string {
	if status {
		return "ONLINE"
	}
	return "OFFLINE"
}

// Função para obter o estado do peer a partir da string
func GetStatus(status string) PeerStatus {
	if status == "ONLINE" {
		return ONLINE
	}
	return OFFLINE
}

// Função para adicinar um novo peer ao SafePeers
func (s *SafePeers) Add(peer Peer) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Se a lista estiver vazia, adiciona o peer diretamente
	if len(s.peers) == 0 {
		s.peers = append(s.peers, peer)
		return
	}

	// Encontra a posição correta para inserir o novo peer
	i := sort.Search(len(s.peers), func(i int) bool {
		return s.peers[i].Address >= peer.Address
	})

	if i < len(s.peers) && s.peers[i].Address == peer.Address {
		// Se o peer já existe, atualiza o status e o clock
		s.peers[i].Status = peer.Status
		s.peers[i].Clock = peer.Clock
	} else {
		// Se não, extende o slice, desloca os peers e adiciona na posição correta
		s.peers = append(s.peers, Peer{})
		copy(s.peers[i+1:], s.peers[i:])
		s.peers[i] = peer
	}
}

// Função para obter um peer do SafePeers pelo endereço
func (s *SafePeers) Get(address string) (Peer, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, peer := range s.peers {
		if peer.Address == address {
			return peer, true
		}
	}
	return Peer{}, false
}

// Função para obter uma cópia do SafePeers
func (s *SafePeers) GetAll() []Peer {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	copyPeers := make([]Peer, len(s.peers))
	copy(copyPeers, s.peers)
	return copyPeers
}

// Função para obter o tamanho do SafePeers
func (s *SafePeers) Len() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.peers)
}
