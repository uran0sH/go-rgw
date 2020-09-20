package connection

import (
	"os"
	"strconv"
	"testing"
)

var mysql = NewMySQL("root", "root", "127.0.0.1:3306", "ceph", "utf8mb4")

func TestMain(m *testing.M) {
	_ = mysql.Init()
	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestMySQL_CreateObject(t *testing.T) {
	for i := 0; i < 10; i++ {
		if mysql.FindObject("test"+strconv.Itoa(i)).ObjectID == "" {
			err := mysql.CreateObject("test"+strconv.Itoa(i), strconv.Itoa(i), false)
			if err != nil {
				t.Error(err)
			}
		}
	}
}

func TestMySQL_SaveObjectTransaction(t *testing.T) {
	t.Parallel()
	for i := 0; i < 10; i++ {
		err := mysql.SaveObjectTransaction("test"+strconv.Itoa(i), strconv.Itoa(i+10), "abc", "abc", false)
		t.Logf("name: %s, id: %s", "test"+strconv.Itoa(i), strconv.Itoa(i+10))
		if err != nil {
			t.Fatal(err)
		}
	}
}

//func TestMySQL_UpdateObject(t *testing.T) {
//	t.Parallel()
//	for i := 0; i < 10; i++ {
//		mysql.UpdateObject("test"+strconv.Itoa(i), strconv.Itoa(i+10))
//		t.Logf("name: %s, id: %s", "test"+strconv.Itoa(i), strconv.Itoa(i+10))
//	}
//}

func TestMySQL_FindObject(t *testing.T) {
	t.Parallel()
	for i := 0; i < 10; i++ {
		object := mysql.FindObject("test" + strconv.Itoa(4))
		t.Logf("name: %s, id: %s", object.ObjectName, object.ObjectID)
	}
}

func TestMySQL_CreateBucketTransaction(t *testing.T) {
	err := mysql.CreateBucketTransaction("111", "111", "123", "123")
	if err != nil {
		t.Error(err)
	}
}
