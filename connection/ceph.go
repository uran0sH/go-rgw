package connection

import (
	"fmt"
	"github.com/ceph/go-ceph/rados"
)

type Ceph struct {
	Connection *rados.Conn
}


func NewCeph() (*Ceph, error) {
	conn, err := rados.NewConn()
	if err != nil {
		return nil, err
	}
	ceph := &Ceph{
		Connection: conn,
	}
	return ceph, nil
}

func (c *Ceph) InitDefault() error {
	err := c.Connection.ReadDefaultConfigFile()
	if err != nil {
		return err
	}
	err = c.Connection.Connect()
	if err != nil {
		return err
	}
	fmt.Println("connect ceph cluster successfully")
	return nil
}

func (c *Ceph) Shutdown() {
	if c.Connection != nil {
		c.Connection.Shutdown()
	}
}

func (c *Ceph) WriteObject(pool string, oid string, data []byte, offset uint64) error {
	ioctx, err := c.Connection.OpenIOContext(pool)
	if err != nil {
		return err
	}
	defer ioctx.Destroy()
	err = ioctx.Write(oid, data, offset)
	if err != nil {
		return err
	}
	return nil
}


func (c *Ceph) ReadObject(pool string, oid string, data []byte, offset uint64) (int, error) {
	ioctx, err := c.Connection.OpenIOContext(pool)
	if err != nil {
		return 0, err
	}
	defer ioctx.Destroy()
	num, err := ioctx.Read(oid, data, offset)
	if err != nil {
		return 0, err
	}
	return num, nil
}

func (c *Ceph) DeleteObject(pool string, oid string) error {
	ioctx, err := c.Connection.OpenIOContext(pool)
	if err != nil {
		return err
	}
	defer ioctx.Destroy()
	return ioctx.Delete(oid)
}