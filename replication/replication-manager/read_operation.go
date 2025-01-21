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

	if rm.Election.NodeID == rm.Election.CoordinatorID {
		return rm.getLocal(key)
	}

	return rm.getFromLeader(key)
}

func (rm *ReplicationManager) getLocal(key int) (*GetResult, error) {
	value, ok := rm.db.Get(rm.bucket, key)
	if !ok {
		return nil, errors.New("failed to get value from local database")
	}

	return &GetResult{
		Value:     value,
		Found:     true,
		Source:    fmt.Sprintf("Node-%d", rm.nodeID),
		Timestamp: time.Now(),
	}, nil
}

func (rm *ReplicationManager) getFromLeader(key int) (*GetResult, error) {
	leaderID := rm.Election.CoordinatorID
	if leaderID == -1 {
		return nil, fmt.Errorf("no leader available")
	}

	leaderAddr, ok := rm.Election.Peers[leaderID]
	if !ok {
		return nil, fmt.Errorf("leader node %d not found in peers list", leaderID)
	}

	client, err := rpc.Dial("tcp", leaderAddr)
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
	if rm.Election.NodeID != rm.Election.CoordinatorID {
		return fmt.Errorf("not the leader")
	}

	localResult, err := rm.getLocal(key)
	if err != nil {
		return err
	}

	*result = *localResult
	return nil
}
