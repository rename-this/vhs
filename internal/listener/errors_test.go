package listener

import (
	"errors"
	"testing"

	"gotest.tools/v3/assert"
)

func TestError(t *testing.T) {
	err := &Error{}

	err.Append(NewInterfaceError("111", errors.New("failed")))
	err.Append(NewInterfaceError("222", errors.New("failed again")))
	err.Append(NewInterfaceError("333", nil))

	expectedErr := "failed to listen on 111: failed\nfailed to listen on 222: failed again\n"
	assert.Equal(t, err.Error(), expectedErr)
}
