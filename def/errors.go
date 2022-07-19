package def

import "fmt"

var ErrNoDriver = fmt.Errorf("no driver found for the provided configuration")
var ErrInvalidSignature = fmt.Errorf("signature does not match the user id")
