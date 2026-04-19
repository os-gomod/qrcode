// Package testing provides contract tests and test utilities for the QR code library.
//
// The [GeneratorContractTest] function validates that any [qrcode.Generator]
// implementation satisfies the core contract (Generate, GenerateWithOptions,
// GenerateToWriter, Close). Assertion helpers like [AssertNoError], [AssertEquals],
// [AssertTrue], and [AssertFalse] simplify test expectations.
//
//	func TestMyGenerator(t *testing.T) {
//	    gen := my.NewGenerator()
//	    testing.GeneratorContractTest(t, gen)
//	}
package testing

import (
	"math/rand/v2"
	"reflect"
	"testing"
)

// AssertNoError fails the test immediately if err is non-nil, printing
// the error value.
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// AssertEquals fails the test immediately if expected != actual using
// [reflect.DeepEqual].
func AssertEquals(t *testing.T, expected, actual any) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected %v (%T), got %v (%T)", expected, expected, actual, actual)
	}
}

// AssertTrue fails the test immediately if condition is false, printing
// the provided message.
func AssertTrue(t *testing.T, condition bool, msg string) {
	t.Helper()
	if !condition {
		t.Fatalf("expected true, got false: %s", msg)
	}
}

// AssertFalse fails the test immediately if condition is true, printing
// the provided message.
func AssertFalse(t *testing.T, condition bool, msg string) {
	t.Helper()
	if condition {
		t.Fatalf("expected false, got true: %s", msg)
	}
}

// RandomString generates a random alphanumeric string of length n
// using lowercase, uppercase, and digit characters.
func RandomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))]
	}
	return string(b)
}
