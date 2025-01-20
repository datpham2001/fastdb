package election

import (
	"fmt"
	"log"
	"net/rpc"
	"sync"
	"time"
)

type RPCResponse struct {
	Success bool
	NodeID  int
}

const (
	ElectionTimeout   = 3 * time.Second
	HeartbeatInterval = 1 * time.Second
)

type NodeState int

const (
	NodeStateFollower NodeState = iota
	NodeStateCandidate
	NodeStateLeader
)

type Node struct {
	ID      int
	Address string
	State   NodeState
	IsAlive bool
}

type BullyElection struct {
	// node election information
	currentNode *Node
	Peers       map[int]*Node

	// election information
	currentLeaderID int
	mu              sync.RWMutex

	// channels for coordination
	stopChan chan bool
}

func NewBullyElection(nodeID int, nodeAddr string, peerList map[int]string) *BullyElection {
	peers := make(map[int]*Node)
	for peerID, peerAddr := range peerList {
		peers[peerID] = &Node{
			ID:      peerID,
			Address: peerAddr,
			IsAlive: true,
		}
	}

	node := &Node{
		ID:      nodeID,
		Address: nodeAddr,
		State:   NodeStateFollower,
		IsAlive: true,
	}

	election := &BullyElection{
		currentNode:     node,
		Peers:           peers,
		currentLeaderID: -1,
		mu:              sync.RWMutex{},
		stopChan:        make(chan bool),
	}

	fmt.Println("election: ", election)

	go election.monitor()
	return election
}

func (b *BullyElection) monitor() {
	ticker := time.NewTicker(ElectionTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !b.hasLeader() {
				b.startElection()
			}
		case <-b.stopChan:
			return
		}
	}
}

func (b *BullyElection) GetCurrentLeader() *Node {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.Peers[b.currentLeaderID]
}

func (b *BullyElection) IsLeader() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.currentNode != nil && b.currentNode.State == NodeStateLeader
}

func (b *BullyElection) hasLeader() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.currentLeaderID != -1
}

func (b *BullyElection) startElection() {
	b.mu.Lock()
	b.currentNode.State = NodeStateCandidate
	b.mu.Unlock()

	higherNodes := b.getHigherPriorityNodes()
	if len(higherNodes) == 0 {
		b.becomeLeader()
		return
	}

	if !b.notifyElectionToHigherNodes(higherNodes) {
		b.becomeLeader()
	}
}

func (b *BullyElection) notifyElectionToHigherNodes(nodes []*Node) bool {
	responses := make(chan bool, len(nodes))

	for _, node := range nodes {
		go func(n *Node) {
			success := b.sendElectionMessage(n)
			responses <- success
		}(node)
	}

	ticker := time.NewTicker(ElectionTimeout)
	defer ticker.Stop()

	for range nodes {
		select {
		case response := <-responses:
			if response {
				return true
			}
		case <-ticker.C:
			return false
		}
	}

	return false
}

func (b *BullyElection) sendElectionMessage(node *Node) bool {
	client, err := rpc.DialHTTP("tcp", node.Address)
	if err != nil {
		log.Printf("Failed to connect to peer %d: %v", node.ID, err)
		return false
	}
	defer client.Close()

	msg := Message{
		Type:     MessageTypeElectionInProgresss,
		SenderID: b.currentNode.ID,
		OccurAt:  time.Now(),
	}

	var response RPCResponse
	err = client.Call("BullyElectionService.HandleMessage", msg, &response)
	if err != nil {
		log.Printf("Failed to handle message %+v from peer %d: %v", msg, node.ID, err)
		return false
	}

	return response.Success
}

func (b *BullyElection) getHigherPriorityNodes() []*Node {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var higherNodes []*Node
	for peerID, peer := range b.Peers {
		if b.currentNode.ID < peerID {
			higherNodes = append(higherNodes, peer)
		}
	}

	return higherNodes
}

func (b *BullyElection) becomeLeader() {
	b.mu.Lock()
	b.currentNode.State = NodeStateLeader
	b.currentLeaderID = b.currentNode.ID
	b.mu.Unlock()

	b.broadcastToPeers(Message{
		Type:     MessageTypeElectionCompleted,
		SenderID: b.currentNode.ID,
		OccurAt:  time.Now(),
	})

	go b.sendHeartbeats()
}

func (b *BullyElection) broadcastToPeers(msg Message) {
	log.Printf("Broadcasting message %+v to peers: %+v", msg, b.Peers)

	for _, peer := range b.Peers {
		go func(node *Node) {
			client, err := rpc.DialHTTP("tcp", node.Address)
			if err != nil {
				log.Printf("Failed to connect to peer %d: %v", node.ID, err)
				return
			}
			defer client.Close()

			var response RPCResponse
			err = client.Call("BullyElectionService.HandleMessage", msg, &response)
			if err != nil {
				log.Printf("Failed to handle message %+v from peer %d: %v", msg, node.ID, err)
			}
		}(peer)
	}
}

func (b *BullyElection) sendHeartbeats() {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.mu.RLock()
			isLeader := b.currentNode.State == NodeStateLeader
			b.mu.RUnlock()

			if !isLeader {
				return
			}

			b.broadcastToPeers(Message{
				Type:     MessageTypePing,
				SenderID: b.currentNode.ID,
				OccurAt:  time.Now(),
			})
		case <-b.stopChan:
			return
		}
	}
}
