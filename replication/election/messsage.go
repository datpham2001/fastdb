package election

import "time"

type MessageType string

const (
	MessageTypePing                MessageType = "ping"
	MessageTypeElectionInProgresss MessageType = "election_in_progress"
	MessageTypeElectionCompleted   MessageType = "election_completed"
)

type Message struct {
	SenderID int
	Type     MessageType
	OccurAt  time.Time
}
