package allocator

import "github.com/google/uuid"

func AllocateID() string {
	oid := uuid.New()
	return oid.String()
}

func AllocateMetadataID() string {
	oid := uuid.New()
	return "metadata_" + oid.String()
}