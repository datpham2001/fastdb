package main

import (
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"testing"

	"github.com/marcelloh/fastdb"
	"github.com/marcelloh/fastdb/service"
)

func setupTestServer(t *testing.T) (*rpc.Client, func()) {
	workDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	parentDir := filepath.Join(workDir, "..")
	dbPath := filepath.Join(parentDir, "data", "integration_test.db")

	testDB, err := fastdb.Open(dbPath, syncTime)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	keyValueStoreService := service.NewKeyValueStoreService(testDB)
	keyValueStoreImpl := &KeyValueStoreImpl{keyValueStoreService}
	if err := rpc.RegisterName("KeyValueStoreService", keyValueStoreImpl); err != nil {
		t.Fatalf("Error registering service: %v", err)
	}

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Error listening: %v", err)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go rpc.ServeConn(conn)
		}
	}()

	client, err := rpc.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Error dialing: %v", err)
	}

	cleanup := func() {
		client.Close()
		listener.Close()
		testDB.Close()
	}

	return client, cleanup
}

func TestRPCIntegration_SetAndGet(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()

	var setReply string
	setArgs := [2]interface{}{1, "test value"}
	err := client.Call("KeyValueStoreService.Set", setArgs, &setReply)
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}
	if setReply != service.SetSuccess {
		t.Errorf("Set reply = %v, want %v", setReply, service.SetSuccess)
	}

	var getReply interface{}
	getArgs := [1]interface{}{1}
	err = client.Call("KeyValueStoreService.Get", getArgs, &getReply)
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if getReply.(string) != "test value" {
		t.Errorf("Get reply = %v, want %v", getReply, "test value")
	}

	getArgs = [1]interface{}{999}
	err = client.Call("KeyValueStoreService.Get", getArgs, &getReply)
	if err == nil {
		t.Error("Expected error for non-existing key, got nil")
	}
}
