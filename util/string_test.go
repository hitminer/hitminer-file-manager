package util

import (
	"github.com/go-playground/assert/v2"
	"testing"
)

func TestCommonSuffix(t *testing.T) {
	assert.Equal(t, CommonSuffix("12312", "4312"), "312")
	assert.Equal(t, CommonSuffix("4312", "12312"), "312")
	assert.Equal(t, CommonSuffix("2341", "2341"), "2341")
	assert.Equal(t, CommonSuffix("11342134", "11342135"), "")
	assert.Equal(t, CommonSuffix("11342135", "11342134"), "")
	assert.Equal(t, CommonSuffix("344", "3424234"), "4")
	assert.Equal(t, CommonSuffix("3424234", "344"), "4")
	assert.Equal(t, CommonSuffix("", "3424234"), "")
	assert.Equal(t, CommonSuffix("3424234", ""), "")
	assert.Equal(t, CommonSuffix("", ""), "")
	assert.Equal(t, CommonSuffix("23241879413764", "23241879413764"), "23241879413764")
	assert.Equal(t, CommonSuffix("542354", "645425341"), "")
}
