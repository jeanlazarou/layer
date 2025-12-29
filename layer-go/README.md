# Layer Go Implementation

**Native and WebAssembly interpreter for the Layer functional programming language**

---

## Quick Start

### Native CLI

```bash
# Build
go build ./cmd/layer

# Run a Layer program
./layer ../examples/calculator.layer

# REPL (coming soon)
./layer
```

### WebAssembly

```bash
# Build WASM module
./build-wasm.sh

# Deploy to visual editor
cp layer.wasm ../visual-editor/frontend/public/

# The visual editor will automatically use the WASM interpreter
```

---

## Features

### Complete Language Implementation

✅ **Lexer** - Tokenization with support for λ symbol, multi-character operators
✅ **Parser** - Recursive descent with precedence climbing
✅ **Interpreter** - Full evaluation with closures and pattern matching
✅ **Layer System** - Core, UI, Data, and Debug layers
✅ **WebAssembly** - Compile to WASM for browser execution

### Layer Functions

**Core Layer** (36 functions):
- Arithmetic: `+`, `-`, `*`, `/`, `%`
- Comparison: `==`, `!=`, `<`, `>`, `<=`, `>=`
- Lists: `map`, `filter`, `reduce`, `fold`, `head`, `tail`, `append`, `length`, `reverse`, `take`, `drop`, `range`
- Strings: `split`, `join`, `str_concat`, `str_length`
- Type conversion: `to_string`, `to_float`, `to_int`
- Functional: `id`, `const`, `compose`, `print`

**Data Layer** (21 functions):
- localStorage: `get`, `set`, `delete_key`, `keys`, `clear`
- Files: `read_file`, `write_file`, `exists`, `delete`, `list_files`
- JSON: `read_json`, `write_json`, `parse_json`, `to_json`
- Options: `Some`, `None`, `is_some`, `is_none`, `unwrap`, `map_option`, `flatmap_option`
- Storage control (WASM only): `request_file_access`, `set_storage_mode`, `get_storage_mode`

**Debug Layer** (8 functions):
- Logging: `log`, `inspect`
- Tracing: `trace`, `benchmark`
- Testing: `assert_equal`
- Utilities: `breakpoint`, `get_logs`, `clear_logs`

**UI Layer**:
- Components: `view`, `text`, `button`, `input`, `container`
- Renders as interactive Material UI components in the visual editor

---

## Architecture

### Project Structure

```
layer-go/
├── cmd/
│   ├── layer/          # Native CLI entry point
│   └── wasm/           # WASM entry point
├── pkg/
│   ├── ast/            # AST node definitions
│   ├── lexer/          # Tokenization
│   ├── parser/         # Parsing
│   ├── interpreter/    # Evaluation + Layer system
│   │   ├── interpreter.go
│   │   ├── core_layer.go
│   │   ├── data_layer_native.go   # Native file I/O
│   │   ├── data_layer_wasm.go     # Browser file I/O
│   │   ├── debug_layer.go
│   │   └── ui_layer.go
│   └── layers/         # Layer context management
├── build-wasm.sh       # WASM build script
└── layer.wasm          # Compiled WASM module (4.3 MB)
```

### Execution Pipeline

```
Source Code (.layer)
    ↓
Lexer → Tokens
    ↓
Parser → AST
    ↓
Interpreter + Layers → Result
```

### Build Tags

The implementation uses Go build tags for platform-specific code:

- **Native**: `data_layer_native.go` - Uses `os` package for real file system
- **WASM**: `data_layer_wasm.go` - Uses `syscall/js` for browser APIs

```go
// data_layer_native.go
//go:build !js && !wasm

// data_layer_wasm.go
//go:build js && wasm
```

---

## Testing

```bash
# Run all tests
go test ./pkg/...

# Run specific test
go test ./pkg/interpreter -run TestStringConcatenation -v

# With coverage
go test ./pkg/... -cover
```

**Current status**: 22/22 tests passing ✅

---

## WebAssembly Details

### WASM Compilation

```bash
./build-wasm.sh
```

This produces:
- `layer.wasm` (4.3 MB) - The interpreter compiled to WebAssembly
- Automatically deployed to `../visual-editor/frontend/public/`

### Browser Integration

The WASM module integrates with the browser through:

1. **browserFileSystem.ts** - Dual-mode file system:
   - Virtual mode: localStorage-based (all browsers)
   - Real mode: File System Access API (Chrome/Edge 86+)

2. **Synchronous API** - To avoid deadlocks, file operations use synchronous JavaScript calls:
   ```javascript
   // Exposed to WASM
   window.layerFileSystem = {
       readFileSync: (path) => { /* localStorage.getItem */ },
       writeFileSync: (path, content) => { /* localStorage.setItem */ },
       // ...
   }
   ```

3. **Go-JS Bridge** - Using `syscall/js`:
   ```go
   fs := js.Global().Get("layerFileSystem")
   result := fs.Call("readFileSync", path)
   ```

### Performance

- **Load time**: < 1 second (4.3 MB gzipped)
- **Execution**: ~10x faster than Python implementation
- **File I/O** (virtual mode): < 10ms per operation
- **File I/O** (real mode): < 100ms per operation

---

## Known Limitations

### WASM-Specific

1. **File operations are synchronous only** - To avoid Go/JavaScript event loop deadlocks
2. **Real file mode requires Chrome/Edge** - File System Access API not in all browsers
3. **No native multithreading** - WASM runs on single thread (for now)

### Language Features

1. **Type inference not implemented** - Dynamic typing only
2. **No module system** - All code in single file
3. **No effect system** - Effects tracked by layer convention only

---

## Development Tips

### Adding a New Built-in Function

1. **Choose the layer** - Core, Data, Debug, or UI?
2. **Add to layer file**:
   ```go
   // In pkg/interpreter/core_layer.go
   lc.Define("myFunction", func(args ...interface{}) (interface{}, error) {
       // Implementation
       return result, nil
   })
   ```
3. **Add tests**:
   ```go
   // In pkg/interpreter/interpreter_test.go
   func TestMyFunction(t *testing.T) {
       // Test code
   }
   ```
4. **Update documentation** - Add to LANGUAGE_SPEC.md

### Debugging

**Native**:
```bash
# Use Go debugger
dlv debug ./cmd/layer -- ../examples/program.layer
```

**WASM**:
```javascript
// Browser console
console.log = console.log; // Enable logging
// WASM logs will appear in console
```

### Common Issues

**Issue**: "goroutine deadlock" in WASM
**Solution**: Use synchronous JavaScript APIs, never await Promises from Go
**Reference**: [SYNC_FILE_OPERATIONS_FIX.md](SYNC_FILE_OPERATIONS_FIX.md)

**Issue**: File not found in browser
**Solution**: Files are in virtual mode (localStorage), check with:
```javascript
localStorage.getItem('layerfs:filename.txt')
```

**Issue**: WASM not updating
**Solution**: Hard refresh browser (Ctrl+Shift+R / Cmd+Shift+R)

---

## Migration from Python

The Layer language was originally implemented in Python. The Go implementation is now the **primary and only maintained version**.

### Why Go?

1. **Performance**: 10x faster execution
2. **WebAssembly**: Can run in browser without server
3. **Type safety**: Easier to maintain and extend
4. **Deployment**: Single binary, no dependencies
5. **Concurrency**: Better support for future parallel execution

### Migration Status

✅ **Complete feature parity** with Python
✅ **All examples working**
✅ **WASM compilation successful**
✅ **Browser integration complete**

---

## Roadmap

### Short Term
- [ ] REPL implementation
- [ ] Better error messages (line/column numbers in WASM)
- [ ] Performance optimizations (bytecode compilation)

### Medium Term
- [ ] Module system
- [ ] Type inference (Hindley-Milner)
- [ ] LSP server for editor support

### Long Term
- [ ] Effect system (algebraic effects)
- [ ] Compile to native code (via LLVM)
- [ ] Package manager

---

## Resources

- **Language Spec**: [../LANGUAGE_SPEC.md](../LANGUAGE_SPEC.md)
- **Philosophy**: [../PHILOSOPHY.md](../PHILOSOPHY.md)
- **Examples**: [../examples/](../examples/)
- **Visual Editor**: [../visual-editor/](../visual-editor/)
- **Testing Guide**: [TESTING_GUIDE.md](TESTING_GUIDE.md)

---

## Contributing

Contributions welcome! The codebase is well-tested and documented.

**Getting started**:
1. Read [../PHILOSOPHY.md](../PHILOSOPHY.md) to understand the language design
2. Check [../LANGUAGE_SPEC.md](../LANGUAGE_SPEC.md) for syntax and semantics
3. Look at [../examples/](../examples/) for usage patterns
4. Run tests: `go test ./pkg/...`

**Areas for contribution**:
- Performance optimizations
- Better error messages
- REPL implementation
- Module system design
- Documentation improvements

---

**Layer Go Implementation**: Fast, portable, browser-ready functional programming.
