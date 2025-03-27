package peers

type PeerStatus bool

const (
	ONLINE  PeerStatus = true
	OFFLINE PeerStatus = false
)

type Peer struct {
	Address string
	Port    string
	Status  PeerStatus
}

func (peer Peer) FullAddress() string {
	return peer.Address + ":" + peer.Port
}

func (peer Peer) GetStatus() string {
	if peer.Status {
		return "Online"
	}
	return "Offline"
}

func (s PeerStatus) String() string {
	if s {
		return "ONLINE"
	}
	return "OFFLINE"
}

func GetPeerStatus(s string) PeerStatus {
	switch s {
	case "ONLINE":
		return ONLINE
	}
	return OFFLINE
}
