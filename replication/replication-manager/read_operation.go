package replicationmanager

import (
	"errors"
	"fmt"
	"net/rpc"
	"time"
)

type ReadPreference int

const (
	ReadFromLeader ReadPreference = iota
	ReadFromLocal
)

type GetResult struct {
	Value     []byte
	Found     bool
	Timestamp time.Time
	Source    string
}

func (rm *ReplicationManager) Get(key int, preference ReadPreference) (*GetResult, error) {
	if preference == ReadFromLocal {
		return rm.getLocal(key)
	}

	if rm.election.IsLeader() {
		return rm.getLocal(key)
	} else {
		return rm.getFromLeader(key)
	}
}

func (rm *ReplicationManager) getLocal(key int) (*GetResult, error) {
	value, ok := rm.db.Get(rm.bucket, key)
	if !ok {
		return nil, errors.New("failed to get value from local database")
	}

	return &GetResult{
		Value:     value,
		Found:     true,
		Source:    fmt.Sprintf("node-%d", rm.nodeID),
		Timestamp: time.Now(),
	}, nil
}

func (rm *ReplicationManager) getFromLeader(key int) (*GetResult, error) {
	leader := rm.election.GetCurrentLeader()
	if leader.ID == -1 {
		return nil, fmt.Errorf("no leader available")
	}

	// Get leader's address from peers map
	leader, exists := rm.election.Peers[leader.ID]
	if !exists {
		return nil, fmt.Errorf("leader node %d not found in peers list", leader.ID)
	}

	client, err := rpc.DialHTTP("tcp", leader.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to leader: %w", err)
	}
	defer client.Close()

	var result GetResult
	err = client.Call("ReplicationManager.HandleGet", key, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get value from leader: %w", err)
	}

	return &result, nil
}

func (rm *ReplicationManager) HandleGet(key int, result *GetResult) error {
	if !rm.election.IsLeader() {
		return fmt.Errorf("not the leader")
	}

	localResult, err := rm.getLocal(key)
	if err != nil {
		return err
	}

	*result = *localResult
	return nil
}
