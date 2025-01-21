package replicationmanager

import (
	"fmt"
	"log"
	"net/rpc"
	"sync"
	"time"
)

func (rm *ReplicationManager) Set(key int, value []byte) error {
	if rm.Election.NodeID != rm.Election.CoordinatorID {
		return fmt.Errorf("not the leader, current leader is Node-%d", rm.Election.CoordinatorID)
	}

	if err := rm.replicateToBackups(key, value); err != nil {
		return fmt.Errorf("failed to replicate to backups: %w", err)
	}

	if err := rm.db.Set(rm.bucket, key, value); err != nil {
		return fmt.Errorf("failed to set key in local db: %w", err)
	}

	return nil
}

func (rm *ReplicationManager) replicateToBackups(key int, value []byte) error {
	var peers []string
	for _, peer := range rm.Election.Peers {
		peers = append(peers, peer)
	}
	errors := make(chan error, len(peers))

	wg := sync.WaitGroup{}
	for _, peer := range peers {
		wg.Add(1)
		go func(peerAddr string) {
			defer wg.Done()
			if err := rm.sendReplication(peerAddr, key, value); err != nil {
				errors <- err
			}
		}(peer)
	}

	wg.Wait()
	close(errors)

	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("replication errors: %v", errs)
	}

	log.Println("Replication to backups completed successfully")
	return nil
}

func (rm *ReplicationManager) sendReplication(peerAddr string, key int, value []byte) error {
	client, err := rpc.Dial("tcp", peerAddr)
	if err != nil {
		return fmt.Errorf("failed to dial peer %s: %w", peerAddr, err)
	}
	defer client.Close()

	request := ReplicationRequest{
		Key:      key,
		Value:    value,
		OccurrAt: time.Now(),
		LeaderID: rm.nodeID,
	}

	var response ReplicationResponse
	if err := client.Call("ReplicationManager.HandleReplication", request, &response); err != nil {
		return fmt.Errorf("failed to replicate to peer %s: %w", peerAddr, err)
	}

	if !response.Success {
		return fmt.Errorf("replication rejected by peer %s", peerAddr)
	}

	return nil
}

func (rm *ReplicationManager) HandleReplication(
	request ReplicationRequest,
	response *ReplicationResponse,
) error {
	if request.LeaderID != rm.Election.CoordinatorID {
		response.Success = false
		return nil
	}

	if err := rm.db.Set(rm.bucket, request.Key, request.Value); err != nil {
		response.Success = false
		return fmt.Errorf("failed to set key in local db: %w", err)
	}

	response.Success = true
	return nil
}
