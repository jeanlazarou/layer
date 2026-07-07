# Layer and Its Relatives

How Layer's ideas relate to prior art — what we share, what we borrow, and
what is genuinely new here. Knowing the neighbors sharpens the design and
tells us where to steal good answers to problems they hit first.

---

## Context-Oriented Programming (the closest relative)

**ContextL (Hirschfeld, Costanza, Nierstrasz, 2005+), JCop, ContextJS.**

COP is the closest existing body of work — its central construct is even
called a *layer*: a named group of partial method definitions that can be
dynamically activated and deactivated, changing program behavior per
context (`with-active-layers` in ContextL).

Shared with Layer:

- Named layers containing function/method definitions
- Layer activation changes dispatch (our `@layers [...]` list, their
  dynamic activation scopes)
- Conditional activation (our `debug if env == "development"`)
- The motivating examples are nearly identical: logging, debugging,
  environment-specific behavior

Where Layer differs:

- **COP layers modify classes/methods; Layer layers *are* the program.**
  In COP a layer is an aspect bolted onto an OO base program. In Layer the
  entire program is a stack of layers over a purely functional core — there
  is no "base program" outside the stack.
- **Activation is (mostly) static and visible.** COP emphasizes dynamic
  scoping of activation, which makes behavior hard to predict. Layer's
  stack is declared once in `@layers [...]` and shown permanently in the
  visual editor. Conditional activation is evaluated at program start, not
  per call frame.
- **Blend modes** (below) have no COP equivalent — COP layers combine only
  by override-with-`proceed`.

*What to study from them*: the COP literature spent a decade on the
"unwanted interception" problem (a layer activated for one purpose changes
behavior somewhere unrelated). Their answers — activation scopes, layer
dependencies, reflection on active layers — map directly onto our
introspection features (`@which`, `@inspect`, `@diff`).

## Aspect-Oriented Programming

**AspectJ, pointcuts/advice.**

AOP's `around` advice wrapping a join point is exactly our wrapping
pattern: `@layer debug { fn fib(n) = ...; core.fib(n) }` is `around`
advice with `core.fib` playing the role of `proceed()`.

Key difference: AOP selects interception points with a **pointcut
language** (quantification over the program: "all methods named set*").
Layer intercepts by **name shadowing only** — a layer overrides `fib` by
defining `fib`. This is far less powerful and far more predictable; you
can always see what a layer overrides by reading it (or asking `@diff`).
We consider the loss of quantification a feature, but AOP's literature on
*aspect interference* (two aspects wrapping the same point) is required
reading for blend-mode design.

## Mixins, CLOS method combination, and inheritance linearization

Layer's stack is a **linearization**: name resolution walks layers top to
bottom, first match wins — structurally the same as Python's MRO or
mixin/trait linearization, but *explicit and programmer-ordered* rather
than computed from an inheritance graph. `core.fib(n)` is a qualified
`super` call to a *specific* ancestor.

CLOS **method combination** is the closest precedent for blend modes:
`:before`/`:after`/`:around` methods, and operator combinations like
`(+ ...)` or `(list ...)` that call *all* applicable methods and merge
results — that is our `collect`. CLOS proves the concept is programmable
and useful; its lesson is that custom combinations are powerful but must
be easy to inspect (hence our introspection ops).

**Open recursion** is inherited from OO virtual dispatch: unqualified
recursive calls re-enter the top of the stack (a debug layer traces every
recursion level), and qualified calls (`core.fib`) opt out — the same
distinction as `this.method()` vs `super.method()`. This is now specified
and test-locked (see LANGUAGE_SPEC.md, "Recursion Across Layers").

## Effect systems and monads (Haskell, Koka, algebraic effects)

Layer and Haskell answer the same question — *how does pure code coexist
with effects?* — with different geometry:

- **Haskell**: effects are in the *types*; `IO a` propagates through every
  signature that touches it. Precise, compiler-checked, and famously
  steep.
- **Layer**: effects are in the *space* — code lives in a pure or impure
  layer, and the runtime enforces that pure layers cannot reach impure
  functions. Coarser than a type-level effect system, but visible at a
  glance and enforceable without type inference.

Algebraic effect handlers (Koka, Eff, OCaml 5) are the modern middle
ground, and `fallback` blend semantics (try handlers in order) rhymes with
handler stacks. If Layer ever wants fine-grained effects (this function
does I/O, that one only reads a clock), effect handlers are the design to
borrow, layered — literally — on top of the current pure/impure split.

The pure/impure layer split with runtime enforcement is implemented; see
LANGUAGE_SPEC.md "Layer Purity System".

## Middleware stacks (Rack, Express, WSGI)

The pragmatic cousin: an ordered stack where each element wraps the next,
composition is by position, and everyone understands it in five minutes.
Layer generalizes middleware from one pipeline of request handlers to
*every function in the program*, and blend modes generalize the single
"call next" protocol to override / chain / collect / fallback.

---

## What is actually novel in Layer

Being honest about the delta:

1. **Blend modes as declarative, per-stack resolution strategies.**
   Override is everywhere; chain is middleware; collect exists in CLOS.
   But making the combination strategy a *declared property of the layer
   stack* (`@layers [...] blend: fallback`), orthogonal to the functions
   themselves, appears to be new. The function doesn't know how it will be
   combined; the composition does. This is the feature to develop and the
   one worth writing up.

2. **The layer stack as the program's visible architecture.** COP and AOP
   hide their composition in activation scopes and pointcuts. Layer's
   editor renders the stack as the primary UI — toggling a layer is
   toggling behavior, and `@which`/`@inspect`/`@diff` make resolution
   queryable from *inside the language*. The Photoshop metaphor is not
   marketing; it is the interaction model.

3. **Purity as a spatial property enforced along the stack**, including
   its interplay with open recursion: pure code may not *request* effects,
   but higher impure layers may *observe* it by interception. That rule
   (see the purity spec) has no direct precedent we know of; the closest
   discussion is COP work on layer-aware reasoning.

## What we should adopt from the neighbors, eventually

- **From COP**: scoped activation (`with_layers([...]) do ...`) — already
  sketched in IDEAS.md — and the interference analyses.
- **From CLOS**: user-defined blend modes (a blend mode is just a fold
  over the stack's definitions; `vote` from IDEAS.md would be the first).
- **From effect systems**: finer-than-layer effect granularity if the
  pure/impure split proves too coarse in practice.
- **From AOP**: nothing syntactic — but read their failure modes before
  adding any form of quantified override.
