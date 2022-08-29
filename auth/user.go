package auth

import "weshare/security"

type User struct {
	Identity security.Identity
	Active   bool
	Admin    bool
}
