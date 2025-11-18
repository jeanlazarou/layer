# Data Layer & Side Effects Brainstorm

**Date:** 2025-11-18
**Topic:** Implementing the data/persistence layer and handling side effects in a functional language

## The Core Challenge

Functional languages prioritize pure functions and immutability, but real-world applications require side effects:
- **Data Layer:** File I/O, databases, persistence
- **UI Layer:** DOM updates, user interactions
- **Network Layer:** HTTP requests, WebSockets

The question is: **How do we handle these effects while maintaining functional purity?**

---

## Side Effects in Functional Languages: Survey of Approaches

### 1. IO Monad (Haskell)
```haskell
-- Pure function that returns a "recipe" for I/O
readFile :: FilePath -> IO String
writeFile :: FilePath -> String -> IO ()

-- Chain effects with bind (>>=)
program :: IO ()
program = readFile "input.txt" >>= writeFile "output.txt"
```

**Pros:** Effects are explicit in the type system
**Cons:** Requires monadic infrastructure, steep learning curve

### 2. Algebraic Effects (Koka, Eff)
```
effect File {
  read(path: string): string
  write(path: string, content: string): unit
}

fn process() {
  val content = File.read("input.txt")
  File.write("output.txt", content.uppercase())
}
```

**Pros:** Effects are tracked but more ergonomic than monads
**Cons:** Requires effect system in type checker

### 3. Command Pattern (Elm)
```elm
-- Pure function returns a description of effects
update : Msg -> Model -> (Model, Cmd Msg)
update msg model =
  case msg of
    SaveData -> (model, saveToDb model.data)
```

**Pros:** Pure core, runtime handles effects
**Cons:** Need runtime to interpret commands

### 4. Interpreter Pattern
Functions return "effect descriptions" that a separate interpreter executes.

---

## Our Current Approach: Debug Layer Analysis

The existing **debug layer** provides an interesting precedent:

```layer
@layer debug {
  fn log(value) = ...  # logs AND returns the value
  fn trace(label, fn) = ...  # wraps a function with tracing
}
```

**Key insight:** The debug layer **performs side effects** (logging) but maintains some level of composability by always returning values.

**Trade-off:** Functions are technically impure, but effects are isolated to the debug layer.

---

## Three Options for Data Layer

### Option A: Pure API, Runtime Execution
**Most "functionally pure" approach**

```layer
@layer data {
  # These return "effect descriptions", don't execute immediately
  fn read(path) = DataEffect("read", path)
  fn write(path, content) = DataEffect("write", path, content)
  fn query(table, filter) = DataEffect("query", table, filter)

  # Compose effects
  fn copy(from, to) =
    read(from) |> fn(content) -> write(to, content)
}

# At the top level, interpreter executes effects
execute(
  data.read("users.json")
  |> parse_json
  |> filter(fn(u) -> u.age > 18)
  |> fn(users) -> data.write("adults.json", users)
)
```

**Pros:**
- Functions stay pure (return descriptions)
- Core language remains functional
- Effects are explicit
- Composable
- Easy to test (inspect effect trees)

**Cons:**
- Need runtime support for executing effects
- More complex mental model
- Requires significant infrastructure

---

### Option B: Direct Effects (Like Debug Layer)
**Pragmatic approach, consistent with current design**

```layer
@layer data {
  # These actually perform I/O, but return values
  fn read(path) =
    # Actually reads file, returns content

  fn write(path, content) =
    # Actually writes file, returns content (for chaining)

  fn query(db, table, filter) =
    # Actually queries, returns results
}

# Usage
data.read("input.txt")
|> process
|> fn(result) -> data.write("output.txt", result)
```

**Pros:**
- Simple, straightforward
- Consistent with debug layer pattern
- Easy to understand
- Productive immediately

**Cons:**
- Functions are impure (but isolated to data layer)
- Order of evaluation matters
- Testing requires mocking or actual I/O

---

### Option C: Transaction-Based (Hybrid)
**Middle ground between purity and pragmatism**

```layer
@layer data {
  # Build a transaction (pure)
  fn transaction() = Transaction.new()
  fn add_read(tx, path) = Transaction.add(tx, ReadOp(path))
  fn add_write(tx, path, content) = Transaction.add(tx, WriteOp(path, content))

  # Execute transaction (impure, but contained)
  fn commit(tx) = execute_transaction(tx)
}

# Usage
transaction()
|> add_read("input.txt")
|> add_write("output.txt", processed_data)
|> commit()  # Only this line has effects
```

**Pros:**
- Clear boundary between pure (building) and impure (committing)
- Transactional semantics (atomic operations)
- Easy to test (inspect transaction before commit)
- Can add rollback support

**Cons:**
- More boilerplate
- Less flexible than direct approach
- Overhead for simple operations

---

## Decision: Start with Option B, Evolve to Option A

### Rationale:

1. **Consistency:** The debug layer already does direct effects with return values. The data layer should match this pattern initially.

2. **Pragmatism:** Option B is simple and gets us productive immediately. We can always add purity later.

3. **Evolution Path:** Once Option B is working, we can introduce effect descriptions (Option A) without breaking existing code.

4. **Learning:** We'll better understand the requirements after using Option B in practice.

---

## Proposed Data Layer API (Option B)

### File Operations
```layer
@layer data {
  fn read_file(path) = ...           # Returns Some(content) or None
  fn write_file(path, content) = ... # Returns Some(content) or None
  fn exists(path) = ...              # Returns boolean
  fn delete(path) = ...              # Returns Some(path) or None
  fn list_files(dir) = ...           # Returns Some([paths]) or None
}
```

### JSON Operations
```layer
@layer data {
  fn read_json(path) = ...           # Returns Some(parsed_object) or None
  fn write_json(path, data) = ...    # Returns Some(data) or None
}
```

### localStorage Operations
(For UI integration)
```layer
@layer data {
  fn get(key) = ...                  # Returns Some(value) or None
  fn set(key, value) = ...           # Returns Some(value) or None
  fn delete(key) = ...               # Returns Some(key) or None
  fn keys() = ...                    # Returns list of keys
  fn clear() = ...                   # Returns unit
}
```

### Pure Helper Functions
(For transforming data)
```layer
@layer data {
  fn map_rows(fn, data) = ...        # Map over rows
  fn filter_rows(predicate, data) = ... # Filter rows
  fn sort_rows(comparator, data) = ...  # Sort rows
}
```

---

## Error Handling: Option Pattern

**Decision:** Use option pattern (Some/None) for error handling.

```layer
# Success case
data.read_file("exists.txt")  # => Some("file contents")

# Error case
data.read_file("missing.txt") # => None

# Pattern matching for handling
match data.read_file(path) {
  Some(content) -> process(content),
  None -> "File not found"
}
```

**Why Option Pattern:**
- Explicit error handling at call site
- No exceptions (maintains functional style)
- Composable with map/flatMap
- Clear success vs. failure cases

**Future Enhancement:** Could add Result type with error messages:
```layer
Ok(value) | Error(message)
```

---

## Cross-Cutting Concern: UI + Data Integration

The UI layer is **declarative** (returns component trees), but will need data:

```layer
@layers [core, data, ui]

fn user_list() =
  let users = data.read_json("users.json") in
  match users {
    Some(user_list) -> column([
      user_list |> map(fn(u) -> text(u.name))
    ]),
    None -> text("Failed to load users")
  }
```

**Key insight:** Component functions can call data layer functions. The component tree is pure, but constructing it may cause effects.

---

## Implementation Decisions

### 1. Scope
- **Initial:** Files and localStorage
- **Future:** HTTP, databases, WebSockets

### 2. Error Handling
- **Pattern:** Option type (Some/None)
- **Future:** Result type (Ok/Error) with error messages

### 3. Type System
- **Current:** Dynamic typing
- **Future:** Possibly Hindley-Milner type system with effect tracking

### 4. Testing Strategy
- **TBD:** Need to explore mocking vs. actual I/O in tests
- **Consideration:** Could add a "mock" mode to data layer for testing

---

## Next Steps

1. ✅ Document this brainstorm
2. Implement core data layer module (`layer/layers/data.py`)
3. Add file operations with option pattern
4. Add JSON operations
5. Add localStorage operations (for web/UI context)
6. Create example programs demonstrating:
   - File I/O with error handling
   - JSON data processing pipeline
   - UI + data integration
7. Write tests for data layer
8. Update LANGUAGE_SPEC.md with data layer documentation

---

## Open Questions

1. **localStorage Context:** How do we handle localStorage in non-web contexts (CLI)?
   - **Option:** Make it conditional based on runtime environment
   - **Option:** Provide fallback to file-based storage

2. **Path Resolution:** Should paths be relative to script location or CWD?
   - **Recommendation:** Relative to CWD (standard behavior)

3. **Encoding:** What encoding for text files?
   - **Recommendation:** UTF-8 by default

4. **Binary Files:** Should we support binary file operations?
   - **Decision:** Start with text/JSON, add binary later if needed

---

## Philosophy: Pragmatic Functional Programming

This language takes a **pragmatic** approach to functional programming:

- **Core layer:** Pure functions, immutability enforced
- **Effect layers (debug, data, ui):** Explicit, isolated side effects
- **Layer composition:** Clear boundaries between pure and impure code

**Goal:** Get the benefits of functional programming (composability, testability, reasoning) without the dogmatism that makes real-world applications difficult.

**Inspiration:** Similar to how Elm handles effects (pure core, command pattern), but more flexible due to layer system.

---

## Future Evolution: Path to Purity (Option A)

Once Option B is stable, we could introduce effect descriptions:

```layer
@layer data {
  # Pure version - returns effect description
  fn read_pure(path) = ReadEffect(path)
  fn write_pure(path, content) = WriteEffect(path, content)

  # Impure version - executes immediately (backward compatibility)
  fn read(path) = execute(ReadEffect(path))
  fn write(path, content) = execute(WriteEffect(path, content))
}

# New pure composition
fn copy_pure(from, to) =
  read_pure(from)
  |> chain(fn(content) -> write_pure(to, content))

# Execute at top level
execute(copy_pure("in.txt", "out.txt"))
```

This gives users a choice: pragmatic (direct) or pure (effects as values).

---

**Conclusion:** Start with Option B for rapid development and practical learning, with a clear path to evolve toward Option A as we understand the requirements better.
