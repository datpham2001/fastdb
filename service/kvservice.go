package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/marcelloh/fastdb"
)

const (
	KeyBucket  = "kvstore"
	SetSuccess = "Set key successfully"
)

type KeyValueStoreService struct {
	db     *fastdb.DB
	bucket string
}

func NewKeyValueStoreService(db *fastdb.DB) *KeyValueStoreService {
	return &KeyValueStoreService{db: db, bucket: KeyBucket}
}

func (s *KeyValueStoreService) Set(args [2]interface{}, reply *string) error {
	if args[0] == nil || args[1] == nil {
		return errors.New("set->key or value is nil")
	}

	key, err := parseKey(args[0])
	if err != nil {
		return fmt.Errorf("set->parse key error: %w", err)
	}
	if key == nil {
		return errors.New("set->key is nil")
	}

	value, err := parseValue(args[1])
	if err != nil {
		return fmt.Errorf("set->parse value error: %w", err)
	}
	if len(value) == 0 {
		return errors.New("set->value is nil")
	}

	if err := s.db.Set(s.bucket, *key, value); err != nil {
		return err
	}

	*reply = SetSuccess
	return nil
}

func (s *KeyValueStoreService) Get(args [1]interface{}, reply *interface{}) error {
	if args[0] == nil {
		return errors.New("get->key is nil")
	}

	key, err := parseKey(args[0])
	if err != nil {
		return fmt.Errorf("get->parse key error: %w", err)
	}
	if key == nil {
		return errors.New("get->key is nil")
	}

	value, ok := s.db.Get(s.bucket, *key)
	if !ok {
		return errors.New("get->key not found")
	}

	json.Unmarshal(value, &reply)
	return nil
}

func parseKey(key interface{}) (*int, error) {
	keyValue, ok := key.(int)
	if !ok {
		return nil, fmt.Errorf("key=%+v, key is not an integer", key)
	}

	if keyValue < 0 {
		return nil, fmt.Errorf("key=%+v, key should be positive", key)
	}

	return &keyValue, nil
}

func parseValue(value interface{}) ([]byte, error) {
	byteValue, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("value=%+v, marshal value error: %w", value, err)
	}

	return byteValue, nil
}
