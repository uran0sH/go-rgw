package allocator

import (
	"github.com/google/uuid"
	"go-rgw/connection"
	"strings"
)

// AllocateObjectID
func AllocateObjectID(bucketID, clusterID string) string {
	objectUUID := uuid.New()
	objectID := strings.Join([]string{clusterID, bucketID, objectUUID.String()}, ".")
	return objectID
}

// AllocateUUID
func AllocateUUID() string {
	id := uuid.New()
	return id.String()
}

func AllocateMetadataID(objectID string) string {
	return objectID + "-metadata"
}

func AllocateAclID(objectID string) string {
	return objectID + "-acl"
}

// GetClusterID
func GetClusterID() (string, error) {
	return connection.CephMgr.Ceph.Connection.GetFSID()
}
