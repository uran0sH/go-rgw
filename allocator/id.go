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

// AllocateBucketID
func AllocateBucketID() string {
	bucketID := uuid.New()
	return bucketID.String()
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
