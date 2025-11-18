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

---

# ANNEX: Deep Dive on Side Effects and Monads

## The Fundamental Problem

Side effects violate **referential transparency** - the property that you can replace a function call with its result:

```layer
# Pure function - referentially transparent
fn double(x) = x * 2

let a = double(5) in       # Can replace with 10
let b = 10 in              # Same behavior!
  a + b

# Impure function - NOT referentially transparent
fn read_file(path) = ...   # Returns different values over time!

let a = read_file("data.txt") in   # Might be "hello"
let b = read_file("data.txt") in   # Might be "goodbye"
  a == b  # FALSE! Same call, different results
```

**The goal of functional programming:** Given the same parameters, always return the same result.

**The reality of useful programs:** We need to read files, write to databases, interact with users - all things that change!

**The question:** How do we reconcile purity with the need for change?

---

## How Monads Solve This: The Key Insight

**Monads don't make side effects pure.** Instead, they make a crucial distinction:

### Building vs. Executing

#### Pure Part: Building Effect Descriptions

```haskell
-- This is a PURE VALUE describing an action
readFileAction :: IO String
readFileAction = readFile "data.txt"

-- These two are IDENTICAL values (referentially transparent!)
let action1 = readFile "data.txt"
let action2 = readFile "data.txt"

-- action1 == action2 is TRUE!
-- They're the same DESCRIPTION of reading the file
```

#### Impure Part: Runtime Execution

```haskell
-- The Haskell runtime executes the description (outside your code)
main :: IO ()
main = readFileAction  -- Runtime performs the actual I/O
```

### The Mental Model

**Without Monads (Direct Effects):**
```javascript
// You press "execute" on this function
function process() {
  let content = readFile("data.txt")  // 💥 Effect happens NOW
  return content.toUpperCase()
}
```

**With Monads (Effect Descriptions):**
```haskell
-- You build a "recipe" or "program"
process :: IO String
process = do
  content <- readFile "data.txt"  -- DESCRIBING the read
  return (toUpperCase content)     -- DESCRIBING the transform

-- The recipe itself is PURE! Same inputs = same recipe
-- But RUNNING the recipe produces different results
```

---

## Concrete Analogy: Cooking Recipes

### Impure Approach (Direct Execution)

```
Step 1: Get flour from the cupboard
        ↓
        💥 Cupboard might be empty! Not deterministic!
```

Every time you "read" this instruction, you might get a different result (flour or no flour).

### Monad Approach (Recipe as Value)

```
Recipe (pure value):
  "Instructions: Get flour from cupboard → Mix with eggs → Bake at 350°F"

Recipe1 == Recipe2  ✓ (Same instructions)
```

The **recipe itself** is pure - two identical recipes are the same value, always.

But **executing** the recipe can give different results (if you're out of flour).

**Key insight:** We've separated the **description** (pure) from the **execution** (impure).

---

## How This Helps Functional Programming

### 1. Referential Transparency of Descriptions

```haskell
-- These are the SAME VALUE (same description)
let io1 = readFile "data.txt" >>= processData
let io2 = readFile "data.txt" >>= processData

-- You can reason about them algebraically!
let io3 = let x = readFile "data.txt" in x >>= processData

-- io1 == io2 == io3  ✓ All identical descriptions
```

### 2. Composition Without Execution

```haskell
-- Pure composition of effects (no execution yet!)
let program = do
  content <- readFile "input.txt"
  let processed = toUpperCase content
  writeFile "output.txt" processed
  return processed

-- I can pass this around, transform it, combine it
-- WITHOUT executing it!
let bigger = program >> readFile "other.txt"
let logged = program >>= \result -> log result >> return result
```

### 3. Separation of Pure and Impure Code

```haskell
-- 100% PURE: business logic (easily testable)
processUser :: User -> ValidatedUser
processUser user =
  if user.age >= 18
    then ValidatedUser user
    else InvalidUser "Too young"

-- IMPURE BOUNDARY: effect description
loadAndProcess :: IO ValidatedUser
loadAndProcess = do
  user <- loadUserFromDB           -- Impure
  return (processUser user)        -- Pure function inside!
```

The pure logic (`processUser`) is completely separate from effects. You can test it without any I/O!

---

## Comparison with Layer Language Approaches

### Option A: Effect Descriptions (Monadic Style)

```layer
# Building phase (pure) - just creating descriptions
let readEffect = data.read_file("input.txt") in
let processEffect = data.map_effect(toUpperCase, readEffect) in
let writeEffect = data.write_file("output.txt", processEffect) in
writeEffect  # This is just a DESCRIPTION!

# Execution phase (impure) - runtime performs effects
execute(writeEffect)  # Only this line causes I/O
```

**Characteristics:**
- ✅ Effect descriptions are referentially transparent
- ✅ Can compose without executing
- ✅ Pure code testable without I/O
- ❌ More complex mental model
- ❌ Requires runtime support

### Option B: Direct Effects (Current Implementation)

```layer
# Combined build + execute (impure)
let content = data.read_file("input.txt") in  # 💥 Happens NOW
let processed = toUpperCase(content) in
data.write_file("output.txt", processed)      # 💥 Happens NOW
```

**Characteristics:**
- ❌ Not referentially transparent (effects happen immediately)
- ✅ Failure is explicit (Option type)
- ✅ Composable (via map_option, unwrap)
- ✅ Simple to understand
- ✅ Pragmatic for real work

---

## The Fundamental Trade-offs

### What Monads Give You

✅ **Referential transparency** of effect *descriptions*
- Same description = same value
- Can replace `let x = readFile "a.txt" in x >> x` with `readFile "a.txt" >> readFile "a.txt"`

✅ **Composability** without execution
- Build complex effect sequences without running them
- Pass effects as values to higher-order functions

✅ **Testability** without I/O
- Test with mock effect descriptions
- Verify that correct effects are built, without executing

✅ **Equational reasoning** about effect sequences
- Use algebraic laws to transform programs
- Optimize effect sequences

### What Monads DON'T Give You

❌ **Pure results from execution**
- Executing `IO String` still depends on external state
- Running the same `IO` action twice can produce different results

❌ **Determinism of effects**
- The *description* is deterministic
- The *execution* is not

❌ **Simplicity**
- Monadic code has a learning curve
- Mental overhead of "building vs executing"

---

## The Core Insight: Purity vs. Execution

```
┌─────────────────────────────────────────────────┐
│ PURE WORLD (Your Code)                          │
│                                                  │
│  readFile :: FilePath -> IO String              │
│  ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^                │
│  This is a PURE function!                       │
│  Same input = same IO action description        │
│                                                  │
│  Composition, transformation, reasoning - all pure│
└─────────────────────────────────────────────────┘
                       ↓
                  Boundary
                       ↓
┌─────────────────────────────────────────────────┐
│ IMPURE WORLD (Runtime/Interpreter)              │
│                                                  │
│  Actually reads the file                        │
│  Actually writes to disk                        │
│  Actually sends HTTP requests                   │
│                                                  │
│  This part is IMPURE and unrepeatable          │
└─────────────────────────────────────────────────┘
```

**The monad approach:** Keep your code pure by only building descriptions. Let the runtime/interpreter do the dirty work.

---

## Example: Comparing Three Worlds

### 1. Truly Impure (JavaScript/Python)

```javascript
function process() {
  let content = readFile("data.txt")  // Effect happens NOW
  let upper = content.toUpperCase()
  writeFile("out.txt", upper)          // Effect happens NOW
  return upper
}

// process() is completely impure
// Can't test without actual I/O
// Can't reason about it algebraically
```

### 2. Monadic (Haskell)

```haskell
process :: IO String
process = do
  content <- readFile "data.txt"      -- Building description
  let upper = toUpperCase content     -- Pure computation
  writeFile "out.txt" upper            -- Building description
  return upper                         -- Building description

-- process is a PURE VALUE (an IO action)
-- Can test by inspecting the action structure
-- Can compose: process >> process
```

### 3. Option Pattern (Layer - Option B)

```layer
fn process() =
  let content_opt = data.read_file("data.txt") in  -- Effect NOW
  let content = data.unwrap(content_opt, "") in
  let upper = toUpperCase(content) in
  data.write_file("out.txt", upper)                -- Effect NOW

-- Impure, but:
-- - Errors are explicit (Option type)
-- - Still composable via map_option
-- - Pragmatic and easy to understand
```

---

## Why Layer Chose Option B (For Now)

### Reasons for Direct Effects

1. **Pragmatism:** Get productive immediately without monadic complexity
2. **Consistency:** Matches the debug layer pattern
3. **Learning:** Understand requirements before adding abstraction
4. **Simplicity:** Lower barrier to entry for users

### What We Still Get

- ✅ **Explicit failure:** Option type makes errors visible
- ✅ **Composability:** `map_option`, `flatmap_option`, `unwrap`
- ✅ **Isolation:** Side effects contained in specific layers
- ✅ **Clear boundaries:** Core layer is pure, data layer is not

### Evolution Path

As the language matures, we can introduce Option A (effect descriptions) alongside Option B:

```layer
@layer data {
  # Pure API (new)
  fn read_pure(path) = ReadEffect(path)

  # Direct API (backward compatible)
  fn read(path) = execute(read_pure(path))
}

# Users can choose their style!
```

---

## Conclusion: What Monads Really Do

**Monads don't eliminate side effects.** They provide a way to:

1. **Describe** effects as pure values
2. **Compose** effect descriptions without executing them
3. **Defer** execution to the runtime/interpreter boundary

This gives you purity *in your code* while accepting that the runtime still has to interact with the messy, changing world.

It's a brilliant trick: **turn effects into data**, manipulate them purely, then let someone else handle the dirty work of execution.

For Layer, we're taking a pragmatic middle ground - acknowledging side effects exist while making them explicit and composable through the Option pattern and layer isolation.

---

**End of Annex**
