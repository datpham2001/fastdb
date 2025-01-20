package service

// func setupTestDB(t *testing.T) *fastdb.DB {
// 	workDir, err := os.Getwd()
// 	if err != nil {
// 		t.Fatalf("failed to get working directory: %v", err)
// 	}

// 	parentDir := filepath.Join(workDir, "..")
// 	dbPath := filepath.Join(parentDir, "data", "unit_test.db")

// 	db, err := fastdb.Open(dbPath, 1000)
// 	if err != nil {
// 		t.Fatalf("Failed to open test database: %v", err)
// 	}

// 	return db
// }

// func TestKeyValueStoreService_Set(t *testing.T) {
// 	db := setupTestDB(t)
// 	defer db.Close()

// 	replicationManager := replicationmanager.NewReplicationManager(db)
// 	service := NewKeyValueStoreService(replicationManager)

// 	tests := []struct {
// 		name    string
// 		args    [2]interface{}
// 		wantErr bool
// 	}{
// 		{
// 			name:    "Valid key-value pair",
// 			args:    [2]interface{}{1, "test value"},
// 			wantErr: false,
// 		},
// 		{
// 			name:    "Nil key",
// 			args:    [2]interface{}{nil, "test value"},
// 			wantErr: true,
// 		},
// 		{
// 			name:    "Negative key",
// 			args:    [2]interface{}{-1, "test value"},
// 			wantErr: true,
// 		},
// 		{
// 			name:    "Nil value",
// 			args:    [2]interface{}{1, nil},
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			var reply string
// 			err := service.Set(tt.args, &reply)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if err == nil && reply != SetSuccess {
// 				t.Errorf("Set() reply = %v, want %v", reply, SetSuccess)
// 			}
// 		})
// 	}
// }

// func TestKeyValueStoreService_Get(t *testing.T) {
// 	db := setupTestDB(t)
// 	defer db.Close()

// 	service := NewKeyValueStoreService(db)

// 	// Setup test data
// 	setupArgs := [2]interface{}{1, "test value"}
// 	var setupReply string
// 	if err := service.Set(setupArgs, &setupReply); err != nil {
// 		t.Fatalf("Failed to setup test data: %v", err)
// 	}

// 	tests := []struct {
// 		name    string
// 		args    [1]interface{}
// 		want    string
// 		wantErr bool
// 	}{
// 		{
// 			name:    "Existing key",
// 			args:    [1]interface{}{1},
// 			want:    "test value",
// 			wantErr: false,
// 		},
// 		{
// 			name:    "Non-existing key",
// 			args:    [1]interface{}{2},
// 			want:    "",
// 			wantErr: true,
// 		},
// 		{
// 			name:    "Nil key",
// 			args:    [1]interface{}{nil},
// 			want:    "",
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			var reply interface{}
// 			err := service.Get(tt.args, &reply)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if err == nil && reply.(string) != tt.want {
// 				t.Errorf("Get() reply = %v, want %v", reply, tt.want)
// 			}
// 		})
// 	}
// }
