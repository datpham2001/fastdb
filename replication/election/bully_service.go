package election

import "errors"

type BullyService struct {
	election *BullyElection
}

func NewBullyService(election *BullyElection) *BullyService {
	return &BullyService{election: election}
}

func (s *BullyService) HandleMessage(msg Message, response *RPCResponse) error {
	switch msg.Type {
	case MessageTypePing:
		return s.handlePingMessage(msg, response)
	case MessageTypeElectionInProgresss:
		return s.handleElectionInProgressMessage(msg, response)
	case MessageTypeElectionCompleted:
		return s.handleElectionCompletedMessage(msg, response)
	}

	return errors.New("invalid message type")
}

func (s *BullyService) handlePingMessage(_ Message, response *RPCResponse) error {
	response.Success = true
	response.NodeID = s.election.currentNode.ID
	return nil
}

func (s *BullyService) handleElectionInProgressMessage(msg Message, response *RPCResponse) error {
	if msg.SenderID < s.election.currentNode.ID {
		response.Success = false
		response.NodeID = s.election.currentNode.ID

		go s.election.startElection()
		return nil
	}

	response.Success = true
	response.NodeID = s.election.currentNode.ID
	return nil
}

func (s *BullyService) handleElectionCompletedMessage(msg Message, response *RPCResponse) error {
	s.election.mu.Lock()
	defer s.election.mu.Unlock()

	s.election.currentLeaderID = msg.SenderID
	s.election.currentNode.State = NodeStateLeader

	response.Success = true
	response.NodeID = s.election.currentNode.ID
	return nil
}
