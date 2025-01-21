package election

import (
	"errors"
	"fmt"
	"log"
	"net/rpc"
	"time"
)

type RPCResponse struct {
	Success bool
}

type BullyAlgorithm struct {
	NodeID        int
	CoordinatorID int
	Peers         map[int]string
}

var (
	noElectionInvoked    = true
	superiorNodeAvailabe = false
)

func NewBullyAlgorithm(nodeID int, coordinatorID int, peers map[int]string) *BullyAlgorithm {
	return &BullyAlgorithm{
		NodeID:        nodeID,
		CoordinatorID: coordinatorID,
		Peers:         peers,
	}
}

func (b *BullyAlgorithm) CommunicateToCoordinator() {
	coorID := b.CoordinatorID
	coorAddr := b.Peers[coorID]

	log.Printf("[Communication] Node-%d: communicate with coordinator Node-%d", b.NodeID, coorID)

	client, err := rpc.Dial("tcp", coorAddr)
	if err != nil {
		log.Printf(
			"[Communication] Node-%d: failed to connect to coordinator Node-%d: %v",
			b.NodeID, coorID, err,
		)

		log.Print("[Election] Start election")
		b.StartElection()
		return
	}

	var msg = Message{
		SenderID: b.NodeID,
		Type:     MessageTypePing,
		OccurAt:  time.Now(),
	}

	var reply RPCResponse
	err = client.Call("BullyAlgorithm.HandleMessage", msg, &reply)
	if err != nil || !reply.Success {
		log.Printf(
			"[Communication] Node-%d: failed to send PING message to coordinator Node-%d: %v",
			b.NodeID, coorID, err,
		)
		log.Print("[Election] Start election")
		b.StartElection()
		return
	}

	log.Printf("[Communication] Node-%d: received PONG message from coordinator Node-%d", b.NodeID, coorID)
}

func (b *BullyAlgorithm) StartElection() {
	for peerID, peerAddr := range b.Peers {
		if b.isItself(peerID) {
			continue
		}

		if b.isLessPriority(peerID) {
			log.Printf("[Election] Node-%d: send ELECTION message to Node-%d", b.NodeID, peerID)

			client, err := rpc.Dial("tcp", peerAddr)
			if err != nil {
				log.Printf("[Election] Node-%d: is not available, error: %v", peerID, err)
				continue
			}

			var msg = Message{
				SenderID: b.NodeID,
				Type:     MessageTypeElectionInProgress,
				OccurAt:  time.Now(),
			}

			var reply RPCResponse
			err = client.Call("BullyAlgorithm.HandleMessage", msg, &reply)
			if err != nil {
				log.Printf(
					"[Election] Node-%d: failed to send ELECTION message to Node-%d: %v",
					b.NodeID, peerID, err,
				)

				continue
			}

			if reply.Success {
				log.Printf("[Election] Node-%d: received ELECTION message from Node-%d", b.NodeID, peerID)
				superiorNodeAvailabe = true
			}
		}

		if !superiorNodeAvailabe {
			b.makeYourselfCoordinator()
		}

		superiorNodeAvailabe = false
		noElectionInvoked = true
	}
}

func (b *BullyAlgorithm) makeYourselfCoordinator() {
	for peerID, peerAddr := range b.Peers {
		log.Printf("[Victory] Node-%d: send VICTORY message to Node-%d", b.NodeID, peerID)

		client, err := rpc.Dial("tcp", peerAddr)
		if err != nil {
			continue
		}

		var msg = Message{
			SenderID: b.NodeID,
			Type:     MessageTypeElectionCompleted,
			OccurAt:  time.Now(),
		}

		var reply RPCResponse
		client.Call("BullyAlgorithm.HandleMessage", msg, &reply)
	}
}

func (b *BullyAlgorithm) HandleMessage(msg Message, reply *RPCResponse) error {
	switch msg.Type {
	case MessageTypePing:
		return b.handlePingMessage(msg, reply)
	case MessageTypeElectionInProgress:
		return b.handleElectionInProgressMessage(msg, reply)
	case MessageTypeElectionCompleted:
		return b.handleElectionCompletedMessage(msg, reply)
	}

	return errors.New("unknown message type")
}

func (b *BullyAlgorithm) handlePingMessage(msg Message, reply *RPCResponse) error {
	log.Printf(
		"[Communication] received PING message from Node-%d", msg.SenderID,
	)

	reply.Success = true
	return nil
}

func (b *BullyAlgorithm) handleElectionInProgressMessage(msg Message, reply *RPCResponse) error {
	fmt.Println("[Election]: Receiving election from", msg.SenderID)
	if msg.SenderID < b.NodeID {
		fmt.Println("[Election]: Sending OK to", msg.SenderID)
		reply.Success = true

		if noElectionInvoked {
			noElectionInvoked = false
			go b.StartElection()
		}
	}

	return nil
}

func (b *BullyAlgorithm) handleElectionCompletedMessage(msg Message, reply *RPCResponse) error {
	b.CoordinatorID = msg.SenderID
	log.Printf("[Victory] The Node-%d is the new coordinator\n", msg.SenderID)

	reply.Success = true
	return nil
}

func (b *BullyAlgorithm) isItself(peerID int) bool {
	return b.NodeID == peerID
}

func (b *BullyAlgorithm) isLessPriority(peerID int) bool {
	return b.NodeID < peerID
}
