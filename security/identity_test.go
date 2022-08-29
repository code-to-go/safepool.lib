package security

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdentity(t *testing.T) {

	identity, err := NewIdentity("test")
	assert.NoErrorf(t, err, "cannot create identity")
	data, err := MarshalIdentity(identity, true)
	assert.NoErrorf(t, err, "cannot marshal private identity")

	println(base64.StdEncoding.EncodeToString(data))

	identity2, err := UnmarshalIdentity(data)
	assert.NoErrorf(t, err, "cannot unmarshal private identity")
	assert.Equal(t, identity, identity2)

	data, err = MarshalIdentity(identity, false)
	assert.NoErrorf(t, err, "cannot marshal public identity")

	println(base64.StdEncoding.EncodeToString(data))
	println(len(base64.StdEncoding.EncodeToString(data)))

	identity2, err = UnmarshalIdentity(data)
	assert.NoErrorf(t, err, "cannot unmarshal private identity")
	for id, key := range identity.Keys {
		assert.Equal(t, key.Public, identity2.Keys[id].Public)
	}
	assert.Equal(t, identity.Nick, identity2.Nick)

}
