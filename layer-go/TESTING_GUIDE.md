# Layer Language Testing Guide

Quick reference for testing the complete Layer implementation.

---

## 🧪 Run All Tests

```bash
cd /home/lab/projects/layer/layer-go
go test ./pkg/interpreter -v
```

**Expected output**: 22/22 tests passing ✅

---

## 🌐 Browser Testing (Visual Editor)

### 1. Test Calculator (String Concatenation Fix)

**File**: `examples/calculator.layer`

**Steps**:
1. Open visual editor
2. Load `examples/calculator.layer`
3. Click digit buttons (7, 8, 9, etc.)
4. Click operation buttons (+, -, *, /)
5. Click = to see result

**Expected**: Display updates correctly without errors ✅

**What was fixed**: The `+` operator now handles string concatenation for display updates

---

### 2. Test Data Layer (Virtual Mode)

**File**: `examples/data_demo.layer`

**Steps**:
1. Open visual editor
2. Load `examples/data_demo.layer`
3. Run the program
4. Open browser DevTools → Console
5. Check localStorage: `localStorage.getItem('layerfs:test_output.txt')`

**Expected output**:
```
=== File I/O Demo ===
[INFO] Hello from Layer!

=== JSON Demo ===
[INFO] [{name: "Alice", age: 30}]

=== localStorage Demo ===
[INFO] alice
[INFO] All keys: [username, score, level]
```

**Storage**: Files stored in localStorage with `layerfs:` prefix ✅

---

### 3. Test Data Layer (Real Mode) - Chrome/Edge Only

**Layer code**:
```layer
@layers [data]

# Request real file system access
let has_access = data.request_file_access() in

match has_access with
| true ->
    # Switch to real mode
    data.set_storage_mode("real")

    # Write to real file system
    data.write_file("test.txt", "Real file on disk!")
    data.write_file("config.json", "{\"theme\": \"dark\"}")

    # Read back
    debug.log(data.read_file("test.txt"))
    debug.log("Storage mode: " ++ data.get_storage_mode())

| false ->
    debug.log("File system access not granted or not supported")
```

**Steps**:
1. Open visual editor in Chrome or Edge
2. Paste code above
3. Run program
4. Browser shows directory picker → select a folder
5. Check that folder on disk → files should exist!

**Expected**:
- Directory picker appears ✅
- Files written to selected directory ✅
- Files readable from disk via CLI ✅

**Verification**:
```bash
# Check files were created
ls ~/selected-folder/
# Should show: test.txt, config.json

# Read with CLI
cat ~/selected-folder/test.txt
# Output: Real file on disk!
```

---

### 4. Test Debug Layer (Tracing)

**File**: `examples/debug_trace.layer`

**Steps**:
1. Open visual editor
2. Load `examples/debug_trace.layer`
3. Run the program
4. Open browser DevTools → Console

**Expected output**:
```
[TRACE] power(2, 10) called
[TRACE] power(2, 9) called
[TRACE] power(2, 8) called
[TRACE] power(2, 7) called
[TRACE] power(2, 6) called
[TRACE] power(2, 5) called
[TRACE] power(2, 4) called
[TRACE] power(2, 3) called
[TRACE] power(2, 2) called
[TRACE] power(2, 1) called
[TRACE] power(2, 0) called
[TRACE] power(2, 0) = 1
[TRACE] power(2, 1) = 2
[TRACE] power(2, 2) = 4
[TRACE] power(2, 3) = 8
[TRACE] power(2, 4) = 16
[TRACE] power(2, 5) = 32
[TRACE] power(2, 6) = 64
[TRACE] power(2, 7) = 128
[TRACE] power(2, 8) = 256
[TRACE] power(2, 9) = 512
[TRACE] power(2, 10) = 1024
```

**What it shows**: Complete execution trace of recursive function ✅

---

## 🖥️ CLI Testing

### 1. Build Native CLI

```bash
cd /home/lab/projects/layer/layer-go
go build -o layer ./cmd/layer
```

### 2. Run Examples

**Calculator**:
```bash
./layer examples/calculator.layer
```

**Data Demo**:
```bash
./layer examples/data_demo.layer
```
Expected: Creates real files in current directory

**Debug Trace**:
```bash
./layer examples/debug_trace.layer
```
Expected: Shows trace output to stderr

---

## 🔄 CLI ↔ Browser Interoperability Test

This is the **killer feature** - same files work in both environments!

### Setup

1. Create a test directory:
```bash
mkdir ~/layer-test-project
cd ~/layer-test-project
```

2. Create a Layer program:
```bash
cat > main.layer << 'EOF'
@layers [core, data, debug]

debug.log("Writing data...")

let user_data = {
    name: "Alice",
    score: 100,
    level: 5
} in

# Write JSON file
data.write_json("user.json", user_data)

# Write text file
data.write_file("log.txt", "Program executed successfully!")

debug.log("Files created!")
EOF
```

### Test Workflow

**Step 1: Run in Browser**
1. Open visual editor in Chrome/Edge
2. Click "Request File Access"
3. Select `~/layer-test-project`
4. Load `main.layer`
5. Run program

**Step 2: Verify with CLI**
```bash
cd ~/layer-test-project

# Check files were created by browser
ls -la
# Should show: main.layer, user.json, log.txt

# Read files
cat user.json
cat log.txt
```

**Step 3: Modify in CLI**
```bash
# Edit with CLI
echo "Updated from CLI" > log.txt
```

**Step 4: Verify in Browser**
1. Refresh browser
2. Re-open `~/layer-test-project`
3. Run: `data.read_file("log.txt")`
4. Expected: "Updated from CLI" ✅

**This proves true interoperability!** 🎉

---

## 🧪 Test Individual Functions

### Core Functions

```layer
@layers [core]

# String operations
debug.log("Hello" ++ " " ++ "World")  # String concatenation
debug.log(split("a,b,c", ","))        # ["a", "b", "c"]
debug.log(join(["a", "b"], "-"))      # "a-b"
debug.log(str_length("hello"))        # 5

# Type conversions
debug.log(to_string(42))              # "42"
debug.log(to_float("3.14"))           # 3.14
debug.log(to_int("42"))               # 42

# List operations
debug.log(take(3, [1,2,3,4,5]))       # [1,2,3]
debug.log(drop(2, [1,2,3,4,5]))       # [3,4,5]

# Functional utilities
let add_one = fn(x) = x + 1 in
debug.log(map(add_one, [1,2,3]))      # [2,3,4]

let double = fn(x) = x * 2 in
let add_ten = fn(x) = x + 10 in
let transform = compose(add_ten, double) in
debug.log(transform(5))               # 20 (5*2 + 10)
```

### Data Functions

```layer
@layers [data, debug]

# localStorage
data.set("key", "value")
debug.log(data.get("key"))            # Some("value")
debug.log(data.keys())                # ["key"]

# Files (virtual mode)
data.write_file("test.txt", "content")
debug.log(data.read_file("test.txt")) # Some("content")
debug.log(data.exists("test.txt"))    # true

# JSON
let obj = {name: "Bob", age: 25} in
data.write_json("data.json", obj)
debug.log(data.read_json("data.json"))  # Some({name: "Bob", age: 25})

# Options
let result = data.get("key") in
debug.log(data.map_option(result, fn(x) = x ++ "!"))  # Some("value!")
```

### Debug Functions

```layer
@layers [core, debug]

# Basic logging
debug.log("Hello")

# Inspection
debug.inspect([1, 2, 3])
debug.inspect({key: "value"})

# Assertions
debug.assert_equal(1 + 1, 2, "Math works")

# Tracing
let factorial = fn(n) =
    match n with
    | 0 -> 1
    | n -> n * factorial(n - 1)
in
let traced_factorial = debug.trace("factorial", factorial) in
traced_factorial(5)  # Shows all recursive calls

# Benchmarking
let slow_fn = fn(n) =
    fold(fn(acc, x) = acc + x, 0, range(0, n))
in
let benchmarked = debug.benchmark("slow_fn", slow_fn) in
benchmarked(10000)  # Shows execution time in ms
```

---

## 🐛 Common Issues

### Issue 1: "layerFileSystem not available"

**Cause**: Browser file system not loaded before WASM

**Fix**: Ensure `browserFileSystem.ts` is imported in `layerWasm.ts`
```typescript
import './browserFileSystem';  // Must be first!
```

**Verify**: Open browser console, type:
```javascript
window.layerFileSystem
// Should show object with functions
```

---

### Issue 2: "File not found" in virtual mode

**Cause**: Files only persist in localStorage

**Fix**: Files are prefixed with `layerfs:` in localStorage

**Verify**:
```javascript
// In browser console
localStorage.getItem('layerfs:your_file.txt')
```

---

### Issue 3: Real mode not working

**Cause**: Browser doesn't support File System Access API

**Check support**:
```javascript
'showDirectoryPicker' in window  // Should be true in Chrome/Edge
```

**Browsers**:
- ✅ Chrome 86+
- ✅ Edge 86+
- ❌ Safari (use virtual mode)
- ❌ Firefox (use virtual mode)

---

### Issue 4: Permission denied in real mode

**Cause**: User didn't grant directory access

**Fix**: Call `data.request_file_access()` and accept the dialog

**Re-request**:
```layer
# Request again
data.request_file_access()
```

---

## ✅ Success Criteria

All features working when:

- [ ] All 22 Go tests pass
- [ ] Calculator works without string errors
- [ ] data_demo.layer runs successfully
- [ ] debug_trace.layer shows trace output
- [ ] Virtual mode stores files in localStorage
- [ ] Real mode (Chrome/Edge) writes to disk
- [ ] Same files accessible from browser and CLI
- [ ] All 36 core functions work
- [ ] All 21 data functions work
- [ ] All 8 debug functions work

---

## 📊 Performance Benchmarks

Expected performance:

- **WASM load time**: < 1 second (4.3 MB)
- **Virtual mode I/O**: < 10ms per operation
- **Real mode I/O**: < 100ms per operation (disk access)
- **Function execution**: Comparable to native Go

Test with:
```layer
@layers [debug]

let test_fn = fn(n) = fold(fn(a,b) = a+b, 0, range(0, n)) in
let bench = debug.benchmark("sum", test_fn) in
bench(100000)
```

---

## 🎯 Next Steps After Testing

Once all tests pass:

1. **Build production apps** with Layer
2. **Share projects** via git (Layer files are text)
3. **Develop in browser**, deploy via CLI
4. **Create libraries** using Layer's module system
5. **Contribute examples** to the project

---

**Happy testing!** 🚀

For issues or questions, check the documentation:
- [IMPLEMENTATION_STATUS.md](IMPLEMENTATION_STATUS.md)
- [BROWSER_FILE_SYSTEM_IMPLEMENTED.md](BROWSER_FILE_SYSTEM_IMPLEMENTED.md)
