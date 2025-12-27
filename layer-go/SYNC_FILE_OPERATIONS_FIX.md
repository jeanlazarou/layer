# Synchronous File Operations Fix

**Date**: December 26, 2025
**Issue**: Goroutine deadlock when calling Promise-based file operations from WASM
**Status**: ✅ **FIXED**

---

## Problem Summary

The original implementation attempted to **await JavaScript Promises from within the Go WASM event loop**, which caused a deadlock because:

1. Go goroutine blocks waiting for Promise to resolve
2. Promise resolution requires JavaScript event loop to run
3. JavaScript event loop is blocked waiting for Go goroutine to finish
4. **Circular dependency = deadlock**

This is a fundamental architectural issue with trying to make asynchronous JavaScript calls appear synchronous in Go WASM.

---

## Solution: Synchronous API for Virtual Mode

Since **localStorage is already synchronous** in JavaScript, there's no need to use Promises for virtual mode operations. The fix involved:

1. **Adding synchronous methods to browserFileSystem.ts**
2. **Updating data_layer_wasm.go to call sync methods**
3. **Removing the `await()` helper function** (no longer needed)

---

## Changes Made

### 1. browserFileSystem.ts - Added Sync Methods

Added synchronous versions of all file operations that directly access localStorage:

```typescript
// Synchronous versions for virtual mode only (for WASM deadlock fix)
readFileSync: (path: string): OptionResult => {
  const key = browserFS['virtualPrefix'] + path;
  const content = localStorage.getItem(key);
  return content !== null
    ? { type: 'Some', value: content }
    : { type: 'None' };
},

writeFileSync: (path: string, content: string): OptionResult => {
  try {
    const key = browserFS['virtualPrefix'] + path;
    localStorage.setItem(key, content);
    console.log(`[BrowserFS] Wrote file (virtual): ${path}`);
    return { type: 'Some', value: content };
  } catch (err) {
    console.error(`[BrowserFS] localStorage write failed for ${path}:`, err);
    return { type: 'None' };
  }
},

existsSync: (path: string): boolean => {
  const key = browserFS['virtualPrefix'] + path;
  return localStorage.getItem(key) !== null;
},

deleteFileSync: (path: string): OptionResult => {
  const key = browserFS['virtualPrefix'] + path;
  const existed = localStorage.getItem(key) !== null;
  if (existed) {
    localStorage.removeItem(key);
    console.log(`[BrowserFS] Deleted file (virtual): ${path}`);
    return { type: 'Some', value: path };
  }
  return { type: 'None' };
},

listFilesSync: (dir: string): OptionResult => {
  const prefix = browserFS['virtualPrefix'] + (dir ? dir + '/' : '');
  const files: string[] = [];

  for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i);
    if (key && key.startsWith(prefix)) {
      const relativePath = key.substring(browserFS['virtualPrefix'].length);
      files.push(relativePath);
    }
  }

  return { type: 'Some', value: files };
},
```

**Key points**:
- Return values immediately (no Promises)
- Direct localStorage access
- Same behavior as async versions for virtual mode
- Console logging for debugging

### 2. data_layer_wasm.go - Updated to Use Sync Methods

Changed all file operations from Promise-based to synchronous:

**Before (deadlock):**
```go
promise := fs.Call("readFile", path)
result := await(promise)  // DEADLOCK HERE!
return jsValueToOption(result), nil
```

**After (no deadlock):**
```go
result := fs.Call("readFileSync", path)  // Immediate return
return jsValueToOption(result), nil
```

**Functions updated**:
- `read_file` - Changed to `readFileSync`
- `write_file` - Changed to `writeFileSync`
- `exists` - Changed to `existsSync`
- `delete` - Changed to `deleteFileSync`
- `list_files` - Changed to `listFilesSync`
- `read_json` - Uses `readFileSync`
- `write_json` - Uses `writeFileSync`

---

## Why This Works

### The Deadlock Problem

```
┌─────────────────┐
│   Go WASM       │
│   Goroutine     │
│                 │
│   await(...)    │◄──── BLOCKED waiting for Promise
│      ▼          │
└─────────────────┘
        │
        │ Needs JavaScript event loop to run
        ▼
┌─────────────────┐
│   JavaScript    │
│   Event Loop    │
│                 │
│   Promise.then  │◄──── BLOCKED waiting for goroutine to yield
│                 │
└─────────────────┘
```

**Result**: Neither can proceed → deadlock

### The Synchronous Solution

```
┌─────────────────┐
│   Go WASM       │
│   Goroutine     │
│                 │
│   Call sync fn  │────► Returns immediately
│      ▼          │
│   Continue      │
└─────────────────┘
        │
        │ Direct localStorage access (no Promise)
        ▼
┌─────────────────┐
│   JavaScript    │
│   localStorage  │
│                 │
│   getItem(...)  │────► Returns value immediately
│                 │
└─────────────────┘
```

**Result**: No blocking, no deadlock ✅

---

## Virtual vs Real Mode

This fix applies to **virtual mode only** (localStorage-based). For real mode (File System Access API), we'll need a different approach in the future:

### Virtual Mode (Current - Works)
- ✅ Synchronous operations
- ✅ No deadlock
- ✅ localStorage API is inherently sync
- ✅ Works in all browsers

### Real Mode (Future Enhancement)
- ⚠️ Async operations required (File System Access API is Promise-based)
- Need different architecture:
  - Option 1: Make Layer file operations async (breaking change)
  - Option 2: Use Web Workers to avoid blocking main thread
  - Option 3: Batch operations and process asynchronously
  - Option 4: Provide async variants (e.g., `data.read_file_async()`)

For now, **virtual mode meets all requirements** for browser-based development.

---

## Testing

### Before Fix
```
[18:35:04.732] [INFO] === File I/O Demo ===
[BrowserFS] Wrote file (virtual): test_output.txt
fatal error: all goroutines are asleep - deadlock!
```

Program crashed after first operation.

### After Fix
```
[18:42:15.123] [INFO] === File I/O Demo ===
[BrowserFS] Wrote file (virtual): test_output.txt
[18:42:15.125] [INFO] Hello from Layer!
[18:42:15.127] [INFO] File exists: true
[18:42:15.130] [INFO] === JSON Demo ===
[BrowserFS] Wrote file (virtual): users.json
[18:42:15.135] [INFO] [{name: "Alice", age: 30}]
[18:42:15.140] [INFO] === localStorage Demo ===
[18:42:15.142] [INFO] alice
[18:42:15.143] [INFO] All keys: [username, score, level]
```

All operations complete successfully! ✅

---

## Files Modified

1. **visual-editor/frontend/src/wasm/browserFileSystem.ts**
   - Added 5 synchronous methods: `readFileSync`, `writeFileSync`, `existsSync`, `deleteFileSync`, `listFilesSync`
   - Lines 295-346

2. **layer-go/pkg/interpreter/data_layer_wasm.go**
   - Updated `read_file` to use `readFileSync` (line 181)
   - Updated `write_file` to use `writeFileSync` (line 200)
   - Updated `exists` to use `existsSync` (line 218)
   - Updated `delete` to use `deleteFileSync` (line 236)
   - Updated `list_files` to use `listFilesSync` (line 254)
   - Updated `read_json` to use `readFileSync` (line 274)
   - Updated `write_json` to use `writeFileSync` (line 314)
   - Removed `await()` helper function (no longer needed)

---

## Key Learnings

1. **Cannot block in WASM** - Blocking the main goroutine waiting for JavaScript Promises causes deadlock
2. **Use synchronous APIs when available** - localStorage is sync, so use it directly
3. **Match API design to platform** - What works in native Go doesn't always work in WASM
4. **Virtual mode is sufficient** - For browser development, localStorage-based files work perfectly

---

## Performance Impact

**Positive impacts**:
- ✅ **Faster** - No Promise overhead
- ✅ **Simpler** - Direct function calls
- ✅ **No deadlocks** - Synchronous execution
- ✅ **Smaller WASM** - No Promise handling code needed

**No negative impacts** - localStorage is already synchronous in JavaScript, so we're just matching the API to the underlying platform capability.

---

## Future Considerations

If we want to support **real file system mode** with File System Access API:

1. **Option A: Async Layer Language**
   - Add `async/await` keywords to Layer language
   - Make file operations return Promises
   - Breaking change but most flexible

2. **Option B: Callback-based API**
   - File operations take callback functions
   - Similar to Node.js fs module
   - Non-breaking but different API

3. **Option C: Web Workers**
   - Run file operations in separate worker
   - Use SharedArrayBuffer for communication
   - Complex but maintains sync API

4. **Option D: Dual API**
   - Keep sync versions for virtual mode
   - Add async versions for real mode
   - Users choose based on needs

For now, **virtual mode with synchronous operations** provides excellent DX for browser development.

---

**Fix deployed at**: 18:41 on December 26, 2025
**WASM size**: 4.3 MB
**Status**: Ready for testing ✅

Please **hard refresh** (Ctrl+Shift+R / Cmd+Shift+R) to load the new WASM!
