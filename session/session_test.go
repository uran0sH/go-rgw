package session

import (
	"bytes"
	"fmt"
	"go-rgw/connection"
	"go-rgw/gc"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

var mysql = connection.NewMySQL("root", "root", "127.0.0.1:3306", "ceph", "utf8mb4")

func TestMain(m *testing.M) {
	fmt.Println("init...")
	_ = mysql.Init()
	connection.InitMySQLManager(mysql)
	ceph, _ := connection.NewCeph()
	_ = ceph.InitDefault()
	connection.InitCephManager(ceph)
	gc.Init()
	exitCode := m.Run()
	os.Exit(exitCode)
}

//func TestAbortMultipartUpload(t *testing.T) {
//	type args struct {
//		bucketName string
//		objectName string
//		uploadID   string
//	}
//	tests := []struct {
//		name    string
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if err := AbortMultipartUpload(tt.args.bucketName, tt.args.objectName, tt.args.uploadID); (err != nil) != tt.wantErr {
//				t.Errorf("AbortMultipartUpload() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}

//func TestCompleteMultipartUpload(t *testing.T) {
//	type args struct {
//		bucketName string
//		objectName string
//		uploadID   string
//		partIDs    []string
//	}
//	tests := []struct {
//		name    string
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if err := CompleteMultipartUpload(tt.args.bucketName, tt.args.objectName, tt.args.uploadID, tt.args.partIDs); (err != nil) != tt.wantErr {
//				t.Errorf("CompleteMultipartUpload() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}

func TestCreateBucket(t *testing.T) {
	type args struct {
		userId     string
		bucketName string
		acl        string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "createBucket1",
			args: args{
				userId:     "root",
				bucketName: "bucket1",
				acl:        "PUBLIC_READ",
			},
			wantErr: false,
		},
	}
	for i, tt := range tests {
		fmt.Println(i, tt.name, tt.args.userId, tt.args.bucketName, tt.args.acl)
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateBucket(tt.args.userId, tt.args.bucketName, tt.args.acl); (err != nil) != tt.wantErr {
				t.Errorf("CreateBucket() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//func TestCreateMultipartUpload(t *testing.T) {
//	type args struct {
//		objectName string
//		bucketName string
//		metadata   string
//		acl        string
//	}
//	tests := []struct {
//		name    string
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if err := CreateMultipartUpload(tt.args.objectName, tt.args.bucketName, tt.args.metadata, tt.args.acl); (err != nil) != tt.wantErr {
//				t.Errorf("CreateMultipartUpload() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}

//func TestGetObject(t *testing.T) {
//	type args struct {
//		bucketName string
//		objectName string
//	}
//	tests := []struct {
//		name     string
//		args     args
//		wantData []byte
//		wantErr  bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			gotData, err := GetObject(tt.args.bucketName, tt.args.objectName)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("GetObject() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(gotData, tt.wantData) {
//				t.Errorf("GetObject() gotData = %v, want %v", gotData, tt.wantData)
//			}
//		})
//	}
//}

func TestSaveObject(t *testing.T) {
	type args struct {
		objectName string
		bucketName string
		object     io.ReadCloser
		hash       string
		metadataM  map[string][]string
		acl        string
	}
	data, _ := ioutil.ReadFile("../testdata/flowers.png")
	reader := bytes.NewReader(data)
	obj := ioutil.NopCloser(reader)
	metadata := make(map[string][]string, 10)
	metadata["suffix"] = []string{".png"}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "saveObject1",
			args: args{
				objectName: "testObj",
				bucketName: "test1",
				object:     obj,
				hash:       "bK5dxPxwc7aYTbylAizGTg==",
				metadataM:  metadata,
				acl:        "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SaveObject(tt.args.objectName, tt.args.bucketName, tt.args.object, tt.args.hash, tt.args.metadataM, tt.args.acl); (err != nil) != tt.wantErr {
				t.Errorf("SaveObject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//func TestSaveObjectPart(t *testing.T) {
//	type args struct {
//		objectName string
//		bucketName string
//		partID     string
//		uploadID   string
//		hash       string
//		object     io.ReadCloser
//		metadataM  map[string][]string
//	}
//	tests := []struct {
//		name    string
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if err := SaveObjectPart(tt.args.objectName, tt.args.bucketName, tt.args.partID, tt.args.uploadID, tt.args.hash, tt.args.object, tt.args.metadataM); (err != nil) != tt.wantErr {
//				t.Errorf("SaveObjectPart() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}

//func Test_readMultipartObject(t *testing.T) {
//	type args struct {
//		oid string
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    []byte
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := readMultipartObject(tt.args.oid)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("readMultipartObject() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("readMultipartObject() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

//func Test_readOneObject(t *testing.T) {
//	type args struct {
//		oid string
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    []byte
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := readOneObject(tt.args.oid)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("readOneObject() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("readOneObject() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

//func Test_rollbackSaveObject(t *testing.T) {
//	type args struct {
//		id string
//	}
//	tests := []struct {
//		name string
//		args args
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//		})
//	}
//}
