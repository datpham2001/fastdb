package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"strconv"

	"github.com/marcelloh/fastdb"
	"github.com/marcelloh/fastdb/replication/election"
	replicationmanager "github.com/marcelloh/fastdb/replication/replication-manager"
	"github.com/marcelloh/fastdb/service"
)

const (
	syncTime = 1000
)

var db *fastdb.DB

type KVStoreService interface {
	Set(args [2]interface{}, reply *string) error
	Get(args [1]interface{}, reply *interface{}) error
}

type KeyValueStoreImpl struct {
	service KVStoreService
}

func (k *KeyValueStoreImpl) Set(args [2]interface{}, reply *string) error {
	return k.service.Set(args, reply)
}

func (k *KeyValueStoreImpl) Get(args [1]interface{}, reply *interface{}) error {
	return k.service.Get(args, reply)
}

func initDB() error {
	if db != nil {
		return nil
	}

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}
	dbPath := filepath.Join(workDir, "data", "fastdb.db")

	db, err = fastdb.Open(dbPath, syncTime)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	return nil
}

func init() {
	if err := initDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
}

func getNode() (int, string) {
	nodeID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Error converting node ID to int: %v", err)
	}

	return nodeID, os.Args[2]
}

func main() {
	if err := initDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	nodeID, nodePort := getNode()
	nodeAddress := fmt.Sprintf("localhost:%s", nodePort)

	peers := map[int]string{
		1: "localhost:8080",
		2: "localhost:8081",
		3: "localhost:8082",
		4: "localhost:8083",
	}
	delete(peers, nodeID)

	election := election.NewBullyElection(nodeID, nodeAddress, peers)
	replicationManager := replicationmanager.NewReplicationManager(nodeID, db, election)
	keyValueStoreService := service.NewKeyValueStoreService(replicationManager)
	keyValueStoreImpl := &KeyValueStoreImpl{keyValueStoreService}
	if err := rpc.RegisterName("KeyValueStoreService", keyValueStoreImpl); err != nil {
		log.Fatalf("Error registering service: %v", err)
	}

	listener, err := net.Listen("tcp", ":"+nodePort)
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}

	fmt.Println("Server is running on port", nodePort)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error accepting connection: %v", err)
		}

		go rpc.ServeConn(conn)
	}
}
