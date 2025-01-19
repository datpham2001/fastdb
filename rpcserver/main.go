package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"path/filepath"

	"github.com/marcelloh/fastdb"
	"github.com/marcelloh/fastdb/service"
)

const (
	rpcPort  = ":8080"
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

func main() {
	if err := initDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	keyValueStoreService := service.NewKeyValueStoreService(db)
	keyValueStoreImpl := &KeyValueStoreImpl{keyValueStoreService}
	if err := rpc.RegisterName("KeyValueStoreService", keyValueStoreImpl); err != nil {
		log.Fatalf("Error registering service: %v", err)
	}

	listener, err := net.Listen("tcp", rpcPort)
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}

	fmt.Println("Server is running on port", rpcPort)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error accepting connection: %v", err)
		}

		go rpc.ServeConn(conn)
	}
}
