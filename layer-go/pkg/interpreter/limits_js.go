//go:build js && wasm

package interpreter

// Under WebAssembly the browser's native call stack (about 1 MB) is far
// smaller than Go's. Go's wasm port unwinds that native stack whenever a
// goroutine parks, so deep Layer recursion hops to a fresh goroutine every
// stackHopEvery calls (see Closure.Call) and native stack use stays bounded.
// Go stacks live in linear memory, so the depth limit is lower than native
// to keep memory use reasonable.
const defaultMaxCallDepth = 20000

// stackHopEvery is a variable (not a constant) so the modulo in Closure.Call
// compiles when it is zero on other platforms.
var stackHopEvery = 200
