package internal_test

import (
	"testing"

	. "github.com/whatedcgveg/v2ray-core/app/log/internal"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
)

func TestAccessLog(t *testing.T) {
	assert := assert.On(t)

	entry := &AccessLog{
		From:   "test_from",
		To:     "test_to",
		Status: "Accepted",
		Reason: "test_reason",
	}

	entryStr := entry.String()
	assert.String(entryStr).Contains("test_from")
	assert.String(entryStr).Contains("test_to")
	assert.String(entryStr).Contains("test_reason")
	assert.String(entryStr).Contains("Accepted")
}
