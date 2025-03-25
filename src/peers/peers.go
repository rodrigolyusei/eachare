package peers

type Peer struct {
	Address string
	Port    string
	Status  bool
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
