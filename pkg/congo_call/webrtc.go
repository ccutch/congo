package congo_call

import "github.com/ccutch/congo/pkg/congo_auth"

type Peer struct {
	identity   *congo_auth.Identity
	room       *Call
	offers     chan *CallOffer
	answers    chan *CallAnswer
	candidates chan string
}

type CallOffer struct {
	Sdp  string `json:"sdp"`
	Type string `json:"type"`
}

type CallAnswer struct {
	Sdp  string `json:"sdp"`
	Type string `json:"type"`
}

func (room *Call) Register(ident *congo_auth.Identity, flusher func(label string, data any)) *Peer {
	peer := &Peer{identity: ident, room: room}
	room.peers = append(room.peers, peer)
	flusher("peer-joined", peer)
	return peer
}

func (room *Call) SendCallOffer(sdp, call_type string) *CallOffer {
	offer := &CallOffer{Sdp: sdp, Type: call_type}
	for _, peer := range room.peers {
		peer.offers <- offer
	}
	return offer
}

func (room *Call) SendCallAnswer(sdp, call_type string) *CallAnswer {
	answer := &CallAnswer{Sdp: sdp, Type: call_type}
	for _, peer := range room.peers {
		peer.answers <- answer
	}
	return answer
}

func (room *Call) SendCandidate(info string) {
	for _, peer := range room.peers {
		peer.candidates <- info
	}
}
