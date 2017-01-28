package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func EqualUint(t *testing.T, expected uint, actual uint, msgAndArgs ...interface{}) bool {
	return assert.Equal(t, expected, actual, msgAndArgs)
}

func EqualUint8(t *testing.T, expected uint8, actual uint8, msgAndArgs ...interface{}) bool {
	return assert.Equal(t, expected, actual, msgAndArgs)
}

func EqualUint16(t *testing.T, expected uint16, actual uint16, msgAndArgs ...interface{}) bool {
	return assert.Equal(t, expected, actual, msgAndArgs)
}

func EqualUint32(t *testing.T, expected uint32, actual uint32, msgAndArgs ...interface{}) bool {
	return assert.Equal(t, expected, actual, msgAndArgs)
}

func EqualUint64(t *testing.T, expected uint64, actual uint64, msgAndArgs ...interface{}) bool {
	return assert.Equal(t, expected, actual, msgAndArgs)
}
