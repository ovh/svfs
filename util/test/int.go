package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func EqualInt(t *testing.T, expected int, actual int, msgAndArgs ...interface{}) bool {
	return assert.Equal(t, expected, actual, msgAndArgs)
}

func EqualInt8(t *testing.T, expected int8, actual int8, msgAndArgs ...interface{}) bool {
	return assert.Equal(t, expected, actual, msgAndArgs)
}

func EqualInt16(t *testing.T, expected int16, actual int16, msgAndArgs ...interface{}) bool {
	return assert.Equal(t, expected, actual, msgAndArgs)
}

func EqualInt32(t *testing.T, expected int32, actual int32, msgAndArgs ...interface{}) bool {
	return assert.Equal(t, expected, actual, msgAndArgs)
}

func EqualInt64(t *testing.T, expected int64, actual int64, msgAndArgs ...interface{}) bool {
	return assert.Equal(t, expected, actual, msgAndArgs)
}
