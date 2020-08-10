package session

const (
	read        = "READ"
	write       = "WRITE"
	fullControl = "FULL_CONTROL"
	// default acl
	private         = "PRIVATE"
	publicRead      = "PUBLIC_READ"
	publicReadWrite = "PUBLIC_READ_WRITE"
)

type Acl struct {
	UserId            string  `json:"owner"`
	DefaultAcl        string  `json:"default"`
	AccessControlList []Grant `json:"accessControlList"`
}

type Grant struct {
	UserId     string `json:"userId"`
	Permission string `json:"permission"`
}

func newAcl(userId, defaultAcl string) *Acl {
	acl := &Acl{
		UserId:     userId,
		DefaultAcl: defaultAcl,
	}
	return acl
}
