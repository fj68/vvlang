# vv Language Design Decisions

This document outlines the key design decisions for the `vv` programming language and the underlying philosophy behind its syntax and runtime behavior.

---

## 1. Numeric Types and Explicit Division
`vv` strictly distinguishes between `int` (64-bit integer) and `float` (64-bit floating-point) to eliminate "implicit magic" during arithmetic operations.

* **Separation of Division Operators (`/` and `/:`)**: 
    * The standard division operator `/` always returns a `float`, even if both operands are integers.
    * For integer division (truncation), the explicit `/:` operator must be used.
    * **Design Intent**: This prevents the instability found in some dynamic languages where the result type fluctuates based on input values, forcing users to be intentional about whether they need precision or a whole number.
* **Mixed-Type Arithmetic**: If an operation involves at least one `float`, the result is promoted to a `float`. Operations between two integers (excluding `/`) maintain the `int` type.

## 2. Strings as Character Lists
In `vv`, strings are not an opaque, independent type; they are designed as "lists of characters".

* **String = `[char]`**: A string is treated exactly like a list where each element is a character (e.g., `'A'`).
* **Design Intent**: 
    * **Unified Concepts**: Knowledge gained about lists (indexing, slicing, `len()`, and higher-order functions) applies directly to strings, minimizing the learning curve.
    * **Minimalism**: By limiting core data structures to "Lists" and "Records," the language avoids a bloat of string-specific methods in the core engine.

## 3. Explicit Recursion (`fun rec` / `also`)
`vv` requires users to explicitly declare when a function is intended to be recursive.

* **`fun rec` and `also`**: Recursive functions must use the `rec` keyword, and mutual recursion must be linked using the `also` keyword.
* **Design Intent**: 
    * **Avoiding Ambiguity**: Automatic recursion detection can lead to behavior that is cryptic for beginners. Requiring an explicit declaration prevents "accidental" recursion and unexpected stack consumption.
    * **Performance Awareness**: It encourages users to recognize recursive logic, making them more mindful of stack overflow risks and computational complexity.

## 4. Control Flow and Scope Safety
The language provides "guardrails" within its control structures to prevent common pitfalls for beginners.

* **Scope-Based `defer`**: Unlike languages where `defer` runs at the end of a function, in `vv`, it executes at the end of the current block, including each iteration of a `while` loop.
    * **Design Intent**: This structurally prevents resource leaks (such as hanging file handles) that occur when resources are initialized inside a loop but not released until the entire function exits.
* **Limited to `while` Loops**: `vv` does not feature a `for` loop. It provides only the basic `while` loop and higher-order list functions (like `map` or `each`).
    * **Design Intent**: This avoids the need to introduce the abstract concept of "Iterators" to beginners. Users either control the counter manually or use clear functional patterns.
* **Unified `end` Keyword**: `if`, `while`, `fun`, `test`, and `begin` are all terminated with the same `end` keyword.
    * **Design Intent**: This reduces the mental overhead of memorizing different closing keywords for different blocks, ensuring a consistent syntactic rhythm.

## 5. Test-First Language Design
`vv` treats testing as a first-class citizen of the language rather than an external library.

* **Built-in `test` Blocks**: Test code is written using the `test "description" ... end` syntax, typically within the same file as the implementation.
    * **Design Intent**: 
        * **Code as Specification**: Placing tests near the implementation allows them to serve as "executable documentation," lowering the barrier to writing and maintaining tests.
        * **Core Philosophy**: It reflects the belief that tests are not an "appendix" to a program but an integral part of its definition and functionality.
* **Minimalist `assert` Syntax**: Logic is verified using a simple `assert expr`.
    * **Design Intent**: Rather than bloating the language with multiple assertion functions (like `assert_eq`), `vv` uses a single keyword. This maintains a small surface area for the language while allowing for future improvements in error reporting without changing the syntax.
* **Bundled Test Runner**: The interpreter includes a `vv test` command out of the box.
    * **Design Intent**: By providing the tooling immediately, `vv` fosters a culture of testing from day one without requiring complex environment setup.

## 6. Error Handling and Interoperability
* **Record-Based `!` Operator**: The postfix `!` operator relies on a specific record structure (`{ type = "ok", value = ... }`) defined in `std/result.vv`.
    * **Design Intent**: By using the "Record" as a protocol for error handling instead of a dedicated "Result Type" in the core, the language remains small and flexible while providing powerful error propagation.
* **`VUserData` and `extern`**: External host resources (like Go objects) are wrapped in a `VUserData` black box.
    * **Design Intent**: This prevents beginners from accidentally manipulating system-level resources (like raw file descriptors) as integers, which could crash the environment. It also ensures that the `vv` code remains portable across different host environments (e.g., "native" vs. "js").

### 7. Canonical Comparison and Omission of `>`, `>=`, and `!=`
`vv` deliberately limits its comparison operators to `==`, `<`, and `<=`. The operators `>`, `>=`, and `!=` are excluded to enforce a single, clear way of expressing logic.

* **Omission of `>` and `>=`**:
    * **Mathematical Consistency**: Any "greater than" check can be expressed as a "less than" check by swapping operands (e.g., `a > b` becomes `b < a`). 
    * **Visual Flow**: This forces a consistent "smaller-to-larger" visual pattern across the entire codebase, reducing the cognitive load required to parse conditional logic.

* **Omission of `!=` (The Choice of `not()`)**:
    * **Avoiding Ambiguity**: While some languages use `!=` or `/=`, these symbols can be non-intuitive or easily confused. Specifically, `/=` is often used as a shorthand for "divide and assign" in other languages, which could lead to syntax errors or confusion in `vv`.
    * **Forced Explicit Scope**: By using `not(a == b)`, the language requires the use of parentheses. This clearly defines the scope of the negation, preventing common mistakes related to operator precedence.
    * **Syntactic Minimalism**: Relying on the built-in `not()` syntax ensures that the language core remains small, treating "not equal" as a logical modification of "equal" rather than a separate primitive.
