package failure_test

import (
	"testing"

	"github.com/mohammadne/sanjagh/webhook/validation/failure"
	"github.com/stretchr/testify/assert"
)

func TestInvalidResponse(t *testing.T) {
	var f failure.Failure
	f.RegisterReason("invalid parameter1")
	f.RegisterReason("invalid parameter%d", 2)
	assert.False(t, f.IsAllowed())
	assert.Equal(t, "invalid parameter1,invalid parameter2", f.Reason())
	assert.Contains(t, f, "invalid parameter1")
	assert.Contains(t, f, "invalid parameter2")
}

func TestValidResponse(t *testing.T) {
	var f failure.Failure
	assert.True(t, f.IsAllowed())
	assert.Equal(t, "", f.Reason())
}
