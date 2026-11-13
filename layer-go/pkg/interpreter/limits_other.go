//go:build !(js && wasm)

package interpreter

// Natively, Go's goroutine stacks grow to 1 GB, so no stack hopping is
// needed (see limits_js.go).
const defaultMaxCallDepth = 100000

var stackHopEvery = 0
