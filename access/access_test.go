package access

import (
	"bytes"
	"testing"
	"weshare/security"
	"weshare/transport"

	"github.com/stretchr/testify/assert"
)

func TestTopicCreation(t *testing.T) {

	self, err := security.NewIdentity("test")
	assert.NoErrorf(t, err, "cannot create identity")

	c, err := transport.ReadConfig("../../credentials/s3-2.yaml")
	assert.NoErrorf(t, err, "Cannot load S3 config: %v", err)

	tp, err := OpenTopic(self, "test.weshare.net/public", 0, []transport.Config{c})
	assert.NoErrorf(t, err, "Cannot create topic: %v", err)
	defer tp.Close()

	_, err = tp.Post("test.txt", bytes.NewBufferString("just a simple test"))
	assert.NoErrorf(t, err, "Cannot create post: %v", err)

}
