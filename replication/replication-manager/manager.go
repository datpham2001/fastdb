package replicationmanager

import (
	"time"

	"github.com/marcelloh/fastdb"
	"github.com/marcelloh/fastdb/replication/election"
)

type ReplicationRequest struct {
	Key      int
	Value    []byte
	OccurrAt time.Time
	LeaderID int
}

type ReplicationResponse struct {
	Success bool
}

const (
	KeyBucket = "kvstore"
)

type ReplicationManager struct {
	nodeID   int
	db       *fastdb.DB
	election *election.BullyElection
	bucket   string
}

func NewReplicationManager(
	nodeID int,
	db *fastdb.DB,
	election *election.BullyElection,
) *ReplicationManager {
	return &ReplicationManager{
		nodeID:   nodeID,
		db:       db,
		election: election,
		bucket:   KeyBucket,
	}
}
