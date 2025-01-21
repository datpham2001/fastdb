package main

import (
	"encoding/gob"
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
	Get(args [1]interface{}, reply *replicationmanager.GetResult) error
}

type KeyValueStoreImpl struct {
	service KVStoreService
}

func (k *KeyValueStoreImpl) Set(args [2]interface{}, reply *string) error {
	return k.service.Set(args, reply)
}

func (k *KeyValueStoreImpl) Get(args [1]interface{}, reply *replicationmanager.GetResult) error {
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
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
	gob.Register([]byte{})
	gob.Register(string(""))
	gob.Register(int(0))
	gob.Register(float64(0))
	gob.Register(bool(false))

	if err := initDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
}

func main() {
	if err := initDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	var (
		nodeID int            = 1
		coorID int            = 3
		peers  map[int]string = map[int]string{
			1: "localhost:8080",
			2: "localhost:8081",
			3: "localhost:8082",
		}
	)

	myID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Error converting nodeID to int: %v", err)
	}

	delete(peers, myID)
	bully := election.NewBullyAlgorithm(nodeID, coorID, peers)
	bully.NodeID = myID

	replicationManager := replicationmanager.NewReplicationManager(myID, db, bully)
	kvStore := service.NewKeyValueStoreService(replicationManager)
	kvStoreImp := &KeyValueStoreImpl{service: kvStore}

	myAddr := "localhost:" + os.Args[2]
	address, err := net.ResolveTCPAddr("tcp", myAddr)
	if err != nil {
		log.Fatalf("Error resolving TCP address: %v", err)
	}

	inbound, err := net.ListenTCP("tcp", address)
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}

	rpc.RegisterName("BullyAlgorithm", bully)
	rpc.RegisterName("ReplicationManager", replicationManager)
	rpc.RegisterName("KeyValueStore", kvStoreImp)
	fmt.Println("server is running with IP address and port number:", address)
	go rpc.Accept(inbound)

	reply := ""
	fmt.Printf("Is this node recovering from a crash?(y/n): ") // Recovery from crash.
	fmt.Scanf("%s", &reply)
	if reply == "y" {
		fmt.Println("Log: Invoking Elections")
		bully.StartElection()
	}

	random := ""
	for {
		fmt.Printf("Press enter for %d to communicate with coordinator.\n", bully.NodeID)
		fmt.Scanf("%s", &random)
		bully.CommunicateToCoordinator()
		fmt.Println("")
	}
}
