package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	rpcPort = "localhost:8082"
)

type GetResult struct {
	Value     []byte
	Found     bool
	Timestamp time.Time
	Source    string
}

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
	gob.Register([]byte{})
	gob.Register(string(""))
	gob.Register(int(0))
	gob.Register(float64(0))
	gob.Register(bool(false))
}

func main() {
	client, err := rpc.Dial("tcp", rpcPort)
	if err != nil {
		log.Fatalf("Error dialing connection to server: %v", err)
	}
	defer client.Close()

	reader := bufio.NewReader(os.Stdin)

	// Input for Set method
	var key int
	fmt.Print("Enter key (integer): ")
	fmt.Scanln(&key)

	fmt.Print("Enter value (any type): ")
	value, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading value: %v", err)
	}
	value = strings.TrimSpace(value)

	var valueType interface{}
	if intValue, err := strconv.Atoi(value); err == nil {
		valueType = intValue
	} else if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		valueType = floatValue
	} else if strings.ToLower(value) == "true" || strings.ToLower(value) == "false" {
		boolValue := strings.ToLower(value) == "true"
		valueType = boolValue
	} else {
		valueType = value
	}

	var setReply string
	setArgs := [2]interface{}{key, valueType}
	err = client.Call("KeyValueStore.Set", setArgs, &setReply)
	if err != nil {
		log.Fatalf("Error calling Set method: %v", err)
	}
	fmt.Println("Set key:", setArgs[0])
	fmt.Println("Set value:", setArgs[1])
	fmt.Println("Set reply:", setReply)

	// Input for Get method
	fmt.Print("\nEnter key to retrieve: ")
	fmt.Scan(&key)

	var getReply GetResult
	getArgs := [1]interface{}{key}
	err = client.Call("KeyValueStore.Get", getArgs, &getReply)
	if err != nil {
		log.Fatalf("Error calling Get method: %v", err)
	}

	fmt.Printf(
		"Value retrieved: [Value] = %s, [Found] = %t, [Timestamp] = %s, [Source] = %s\n",
		getReply.Value, getReply.Found, getReply.Timestamp, getReply.Source,
	)
}
