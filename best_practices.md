## Go Best Practices

You almost never need a pointer to an interface. You should be passing interfaces as values—the underlying data can still be a pointer.
An interface is two fields:
1. A pointer to some type-specific information. You can think of this as "type."
2. Data pointer. If the data stored is a pointer, it’s stored directly. If the data stored is a value, then a pointer to the value is stored.
If you want interface methods to modify the underlying data, you must use a pointer.

Verify interface compliance at compile time where appropriate. This includes:
- Exported types that are required to implement specific interfaces as part of their API contract
- Exported or unexported types that are part of a collection of types implementing the same interface
- Other cases where violating an interface would break users

The statement `var _ http.Handler = (*Handler)(nil)` will fail to compile if `*Handler` ever stops matching the `http.Handler` interface.
The right hand side of the assignment should be the zero value of the asserted type. This is `nil` for pointer types (like `*Handler`), slices, and maps, and an empty struct for struct types.

Methods with value receivers can be called on pointers as well as values. Methods with pointer receivers can only be called on pointers or addressable values.

Similarly, an interface can be satisfied by a pointer, even if the method has a value receiver.

Effective Go has a good write up on Pointers vs. Values.

The zero-value of `sync.Mutex` and `sync.RWMutex` is valid, so you almost never need a pointer to a mutex.

If you use a struct by pointer, then the mutex should be a non-pointer field on it. Do not embed the mutex on the struct, even if the struct is not exported.

The `Mutex` field, and the `Lock` and `Unlock` methods are unintentionally part of the exported API of `SMap`.

The mutex and its methods are implementation details of `SMap` hidden from its callers.

Slices and maps contain pointers to the underlying data so be wary of scenarios when they need to be copied.

Keep in mind that users can modify a map or slice you received as an argument if you store a reference to it.

Similarly, be wary of user modifications to maps or slices exposing internal state.

Use defer to clean up resources such as files and locks.

Defer has an extremely small overhead and should be avoided only if you can prove that your function execution time is in the order of nanoseconds. The readability win of using defers is worth the miniscule cost of using them. This is especially true for larger methods that have more than simple memory accesses, where the other computations are more significant than the `defer`.

Channels should usually have a size of one or be unbuffered. By default, channels are unbuffered and have a size of zero. Any other size must be subject to a high level of scrutiny. Consider how the size is determined, what prevents the channel from filling up under load and blocking writers, and what happens when this occurs.

The standard way of introducing enumerations in Go is to declare a custom type and a `const` group with `iota`. Since variables have a 0 default value, you should usually start your enums on a non-zero value.

There are cases where using the zero value makes sense, for example when the zero value case is the desirable default behavior.

Time is complicated. Incorrect assumptions often made about time include the following.
1. A day has 24 hours
2. An hour has 60 minutes
3. A week has 7 days
4. A year has 365 days
For example, *1* means that adding 24 hours to a time instant will not always yield a new calendar day.
Therefore, always use the `"time"` package when dealing with time because it helps deal with these incorrect assumptions in a safer, more accurate manner.

Use `time.Time` when dealing with instants of time, and the methods on `time.Time` when comparing, adding, or subtracting time.

Use `time.Duration` when dealing with periods of time.

Going back to the example of adding 24 hours to a time instant, the method we use to add time depends on intent. If we want the same time of the day, but on the next calendar day, we should use `Time.AddDate`. However, if we want an instant of time guaranteed to be 24 hours after the previous time, we should use `Time.Add`.

Use `time.Duration` and `time.Time` in interactions with external systems when possible. For example:
- Command-line flags: `flag` supports `time.Duration` via `time.ParseDuration`
- JSON: `encoding/json` supports encoding `time.Time` as an RFC 3339 string via its `UnmarshalJSON` method
- SQL: `database/sql` supports converting `DATETIME` or `TIMESTAMP` columns into `time.Time` and back if the underlying driver supports it
- YAML: `gopkg.in/yaml.v2` supports `time.Time` as an RFC 3339 string, and `time.Duration` via `time.ParseDuration`.
When it is not possible to use `time.Duration` in these interactions, use `int` or `float64` and include the unit in the name of the field.
For example, since `encoding/json` does not support `time.Duration`, the unit is included in the name of the field.

When it is not possible to use `time.Time` in these interactions, unless an alternative is agreed upon, use `string` and format timestamps as defined in RFC 3339. This format is used by default by `Time.UnmarshalText` and is available for use in `Time.Format` and `time.Parse` via `time.RFC3339`.
Although this tends to not be a problem in practice, keep in mind that the `"time"` package does not support parsing timestamps with leap seconds (8728), nor does it account for leap seconds in calculations (15190). If you compare two instants of time, the difference will not include the leap seconds that may have occurred between those two instants.

There are few options for declaring errors. Consider the following before picking the option best suited for your use case.
- Does the caller need to match the error so that they can handle it? If yes, we must support the `errors.Is` or `errors.As` functions by declaring a top-level error variable or a custom type.
- Is the error message a static string, or is it a dynamic string that requires contextual information? For the former, we can use `errors.New`, but for the latter we must use `fmt.Errorf` or a custom error type.
For example, use `errors.New` for an error with a static string. Export this error as a variable to support matching it with `errors.Is` if the caller needs to match and handle this error.

For an error with a dynamic string, use `fmt.Errorf` if the caller does not need to match it, and a custom `error` if the caller does need to match it.

Note that if you export error variables or types from a package, they will become part of the public API of the package.

There are three main options for propagating errors if a call fails:
- return the original error as-is
- add context with `fmt.Errorf` and the `%w` verb
- add context with `fmt.Errorf` and the `%v` verb
Return the original error as-is if there is no additional context to add. This maintains the original error type and message. This is well suited for cases when the underlying error message has sufficient information to track down where it came from.
Otherwise, add context to the error message where possible so that instead of a vague error such as "connection refused", you get more useful errors such as "call service foo: connection refused".
Use `fmt.Errorf` to add context to your errors, picking between the `%w` or `%v` verbs based on whether the caller should be able to match and extract the underlying cause.
- Use `%w` if the caller should have access to the underlying error. This is a good default for most wrapped errors, but be aware that callers may begin to rely on this behavior. So for cases where the wrapped error is a known `var` or type, document and test it as part of your function's contract.
- Use `%v` to obfuscate the underlying error. Callers will be unable to match it, but you can switch to `%w` in the future if needed.
When adding context to returned errors, keep the context succinct by avoiding phrases like "failed to", which state the obvious and pile up as the error percolates up through the stack.
However once the error is sent to another system, it should be clear the message is an error (e.g. an `err` tag or "Failed" prefix in logs).

For error values stored as global variables, use the prefix `Err` or `err` depending on whether they're exported.
For custom error types, use the suffix `Error` instead.

When a caller receives an error from a callee, it can handle it in a variety of different ways depending on what it knows about the error.
These include, but not are limited to:
- if the callee contract defines specific errors, matching the error with `errors.Is` or `errors.As` and handling the branches differently
- if the error is recoverable, logging the error and degrading gracefully
- if the error represents a domain-specific failure condition, returning a well-defined error
- returning the error, either wrapped or verbatim
Regardless of how the caller handles the error, it should typically handle each error only once. The caller should not, for example, log the error and then return it, because *its* callers may handle the error as well.

For example, consider the following cases:
**Bad**: Log the error and return it. Callers further up the stack will likely take a similar action with the error. Doing so causing a lot of noise in the application logs for little value.
**Good**: Wrap the error and return it. Callers further up the stack will handle the error. Use of `%w` ensures they can match the error with `errors.Is` or `errors.As` if relevant.
**Good**: Log the error and degrade gracefully. If the operation isn't strictly necessary, we can provide a degraded but unbroken experience by recovering from it.
**Good**: Match the error and degrade gracefully. If the callee defines a specific error in its contract, and the failure is recoverable, match on that error case and degrade gracefully. For all other cases, wrap the error and return it. Callers further up the stack will handle other errors.

The single return value form of a type assertion will panic on an incorrect type. Therefore, always use the "comma ok" idiom.

Code running in production must avoid panics. Panics are a major source of cascading failures. If an error occurs, the function must return an error and allow the caller to decide how to handle it.

Panic/recover is not an error handling strategy. A program must panic only when something irrecoverable happens such as a nil dereference. An exception to this is program initialization: bad things at program startup that should abort the program may cause panic.

Even in tests, prefer `t.Fatal` or `t.FailNow` over panics to ensure that the test is marked as failed.

Atomic operations with the sync/atomic package operate on the raw types (`int32`, `int64`, etc.) so it is easy to forget to use the atomic operation to read or modify the variables.
go.uber.org/atomic adds type safety to these operations by hiding the underlying type. Additionally, it includes a convenient `atomic.Bool` type.

Avoid mutating global variables, instead opting for dependency injection. This applies to function pointers as well as other kinds of values.

These embedded types leak implementation details, inhibit type evolution, and obscure documentation. Assuming you have implemented a variety of list types using a shared `AbstractList`, avoid embedding the `AbstractList` in your concrete list implementations. Instead, hand-write only the methods to your concrete list that will delegate to the abstract list.

Go allows type embedding as a compromise between inheritance and composition. The outer type gets implicit copies of the embedded type's methods. These methods, by default, delegate to the same method of the embedded instance.
The struct also gains a field by the same name as the type. So, if the embedded type is public, the field is public. To maintain backward compatibility, every future version of the outer type must keep the embedded type.
An embedded type is rarely necessary. It is a convenience that helps you avoid writing tedious delegate methods.
Even embedding a compatible AbstractList *interface*, instead of the struct, would offer the developer more flexibility to change in the future, but still leak the detail that the concrete lists use an abstract implementation.

Either with an embedded struct or an embedded interface, the embedded type places limits on the evolution of the type.
- Adding methods to an embedded interface is a breaking change.
- Removing methods from an embedded struct is a breaking change.
- Removing the embedded type is a breaking change.
- Replacing the embedded type, even with an alternative that satisfies the same interface, is a breaking change.
Although writing these delegate methods is tedious, the additional effort hides an implementation detail, leaves more opportunities for change, and also eliminates indirection for discovering the full List interface in documentation.

The Go language specification outlines several built-in, predeclared identifiers that should not be used as names within Go programs.
Depending on context, reusing these identifiers as names will either shadow the original within the current lexical scope (and any nested scopes) or make affected code confusing. In the best case, the compiler will complain; in the worst case, such code may introduce latent, hard-to-grep bugs.

Note that the compiler will not generate errors when using predeclared identifiers, but tools such as `go vet` should correctly point out these and other cases of shadowing.

Avoid `init()` where possible. When `init()` is unavoidable or desirable, code should attempt to:
1. Be completely deterministic, regardless of program environment or invocation.
2. Avoid depending on the ordering or side-effects of other `init()` functions. While `init()` ordering is well-known, code can change, and thus relationships between `init()` functions can make code brittle and error-prone.
3. Avoid accessing or manipulating global or environment state, such as machine information, environment variables, working directory, program arguments/inputs, etc.
4. Avoid I/O, including both filesystem, network, and system calls.
Code that cannot satisfy these requirements likely belongs as a helper to be called as part of `main()` (or elsewhere in a program's lifecycle), or be written as part of `main()` itself. In particular, libraries that are intended to be used by other programs should take special care to be completely deterministic and not perform "init magic".

Considering the above, some situations in which `init()` may be preferable or necessary might include:
- Complex expressions that cannot be represented as single assignments.
- Pluggable hooks, such as `database/sql` dialects, encoding type registries, etc.
- Optimizations to Google Cloud Functions and other forms of deterministic precomputation.

Go programs use `os.Exit` to exit immediately. (Panicking is not a good way to exit programs, please don't panic.)
Call one of `os.Exit` or `log.Fatal*` **only in `main()`**. All other functions should return errors to signal failure.

Rationale: Programs with multiple functions that exit present a few issues:
- Non-obvious control flow: Any function can exit the program so it becomes difficult to reason about the control flow.
- Difficult to test: A function that exits the program will also exit the test calling it. This makes the function difficult to test and introduces risk of skipping other tests that have not yet been run by `go test`.
- Skipped cleanup: When a function exits the program, it skips function calls enqueued with `defer` statements. This adds risk of skipping important cleanup tasks.

If possible, prefer to call `os.Exit` or `log.Fatal` **at most once** in your `main()`. If there are multiple error scenarios that halt program execution, put that logic under a separate function and return errors from it.
This has the effect of shortening your `main()` function and putting all key business logic into a separate, testable function.

The example above uses `log.Fatal`, but the guidance also applies to `os.Exit` or any library code that calls `os.Exit`.

You may alter the signature of `run()` to fit your needs. For example, if your program must exit with specific exit codes for failures, `run()` may return the exit code instead of an error. This allows unit tests to verify this behavior directly as well.

More generally, note that the `run()` function used in these examples is not intended to be prescriptive. There's flexibility in the name, signature, and setup of the `run()` function. Among other things, you may:
- accept unparsed command line arguments (e.g., `run(os.Args[1:])`)
- parse command line arguments in `main()` and pass them onto `run`
- use a custom error type to carry the exit code back to `main()`
- put business logic in a different layer of abstraction from `package main`
This guidance only requires that there's a single place in your `main()` responsible for actually exiting the process.

Any struct field that is marshaled into JSON, YAML, or other formats that support tag-based field naming should be annotated with the relevant tag.

Rationale: The serialized form of the structure is a contract between different systems. Changes to the structure of the serialized form--including field names--break this contract. Specifying field names inside tags makes the contract explicit, and it guards against accidentally breaking the contract by refactoring or renaming fields.

Goroutines are lightweight, but they're not free: at minimum, they cost memory for their stack and CPU to be scheduled.
While these costs are small for typical uses of goroutines, they can cause significant performance issues when spawned in large numbers without controlled lifetimes.
Goroutines with unmanaged lifetimes can also cause other issues like preventing unused objects from being garbage collected and holding onto resources that are otherwise no longer used.
Therefore, do not leak goroutines in production code. Use go.uber.org/goleak to test for goroutine leaks inside packages that may spawn goroutines.
In general, every goroutine:
- must have a predictable time at which it will stop running; or
- there must be a way to signal to the goroutine that it should stop
In both cases, there must be a way code to block and wait for the goroutine to finish. For example:
Bad - There's no way to stop this goroutine. This will run until the application exits.
Good - This goroutine can be stopped with `close(stop)`, and we can wait for it to exit with `<-done`.

Given a goroutine spawned by the system, there must be a way to wait for the goroutine to exit. There are two popular ways to do this:
- Use a `sync.WaitGroup`. Do this if there are multiple goroutines that you want to wait for
- Add another `chan struct{}` that the goroutine closes when it's done. Do this if there's only one goroutine.

`init()` functions should not spawn goroutines.
If a package has need of a background goroutine, it must expose an object that is responsible for managing a goroutine's lifetime.
The object must provide a method (`Close`, `Stop`, `Shutdown`, etc) that signals the background goroutine to stop, and waits for it to exit.
Bad - Spawns a background goroutine unconditionally when the user exports this package. The user has no control over the goroutine or a means of stopping it.
Good - Spawns the worker only if the user requests it. Provides a means of shutting down the worker so that the user can free up resources used by the worker. Note that you should use `WaitGroup`s if the worker manages multiple goroutines. Performance-specific guidelines apply only to the hot path.

When converting primitives to/from strings, `strconv` is faster than `fmt`.

Do not create byte slices from a fixed string repeatedly. Instead, perform the conversion once and capture the result.

Specify container capacity where possible in order to allocate memory for the container up front. This minimizes subsequent allocations (by copying and resizing of the container) as elements are added.

Where possible, provide capacity hints when initializing maps with `make()`.

Providing a capacity hint to `make()` tries to right-size the map at initialization time, which reduces the need for growing the map and allocations as elements are added to the map.
Note that, unlike slices, map capacity hints do not guarantee complete, preemptive allocation, but are used to approximate the number of hashmap buckets required. Consequently, allocations may still occur when adding elements to the map, even up to the specified capacity.
Bad. `m` is created without a size hint; there may be more allocations at assignment time.
Good. `m` is created with a size hint; there may be fewer allocations at assignment time.

Where possible, provide capacity hints when initializing slices with `make()`, particularly when appending.

Unlike maps, slice capacity is not a hint: the compiler will allocate enough memory for the capacity of the slice as provided to `make()`, which means that subsequent `append()` operations will incur zero allocations (until the length of the slice matches the capacity, after which any appends will require a resize to hold additional elements).

Avoid lines of code that require readers to scroll horizontally or turn their heads too much.
We recommend a soft line length limit of **99 characters**. Authors should aim to wrap lines before hitting this limit, but it is not a hard limit. Code is allowed to exceed this limit.

Some of the guidelines outlined in this document can be evaluated objectively; others are situational, contextual, or subjective. Above all else, **be consistent**.
Consistent code is easier to maintain, is easier to rationalize, requires less cognitive overhead, and is easier to migrate or update as new conventions emerge or classes of bugs are fixed.
Conversely, having multiple disparate or conflicting styles within a single codebase causes maintenance overhead, uncertainty, and cognitive dissonance, all of which can directly contribute to lower velocity, painful code reviews, and bugs.
When applying these guidelines to a codebase, it is recommended that changes are made at a package (or larger) level: application at a sub-package level violates the above concern by introducing multiple styles into the same code.

Go supports grouping similar declarations.

This also applies to constants, variables, and type declarations.

Only group related declarations. Do not group declarations that are unrelated.

Groups are not limited in where they can be used. For example, you can use them inside of functions.

Exception: Variable declarations, particularly inside functions, should be grouped together if declared adjacent to other variables. Do this for variables declared together even if they are unrelated.

There should be two import groups:
- Standard library
- Everything else
This is the grouping applied by goimports by default.

When naming packages, choose a name that is:
- All lower-case. No capitals or underscores.
- Does not need to be renamed using named imports at most call sites.
- Short and succinct. Remember that the name is identified in full at every call site.
- Not plural. For example, `net/url`, not `net/urls`.
- Not "common", "util", "shared", or "lib". These are bad, uninformative names.

We follow the Go community's convention of using MixedCaps for function names. An exception is made for test functions, which may contain underscores for the purpose of grouping related test cases, e.g., `TestMyFunction_WhatIsBeingTested`.

Import aliasing must be used if the package name does not match the last element of the import path.

In all other scenarios, import aliases should be avoided unless there is a direct conflict between imports.

- Functions should be sorted in rough call order.
- Functions in a file should be grouped by receiver.
Therefore, exported functions should appear first in a file, after `struct`, `const`, `var` definitions.
A `newXYZ()`/`NewXYZ()` may appear after the type is defined, but before the rest of the methods on the receiver.
Since functions are grouped by receiver, plain utility functions should appear towards the end of the file.

Code should reduce nesting where possible by handling error cases/special conditions first and returning early or continuing the loop. Reduce the amount of code that is nested multiple levels.

If a variable is set in both branches of an if, it can be replaced with a single if.

At the top level, use the standard `var` keyword. Do not specify the type, unless it is not the same type as the expression.
Specify the type if the type of the expression does not match the desired type exactly.

Prefix unexported top-level `var`s and `const`s with `_` to make it clear when they are used that they are global symbols.
Rationale: Top-level variables and constants have a package scope. Using a generic name makes it easy to accidentally use the wrong value in a different file.

**Exception**: Unexported error values may use the prefix `err` without the underscore.

Embedded types should be at the top of the field list of a struct, and there must be an empty line separating embedded fields from regular fields.

Embedding should provide tangible benefit, like adding or augmenting functionality in a semantically-appropriate way. It should do this with zero adverse user-facing effects. Exception: Mutexes should not be embedded, even on unexported types.
Embedding **should not**:
- Be purely cosmetic or convenience-oriented.
- Make outer types more difficult to construct or use.
- Affect outer types' zero values. If the outer type has a useful zero value, it should still have a useful zero value after embedding the inner type.
- Expose unrelated functions or fields from the outer type as a side-effect of embedding the inner type.
- Expose unexported types.
- Affect outer types' copy semantics.
- Change the outer type's API or type semantics.
- Embed a non-canonical form of the inner type.
- Expose implementation details of the outer type.
- Allow users to observe or control type internals.
- Change the general behavior of inner functions through wrapping in a way that would reasonably surprise users.
Simply put, embed consciously and intentionally. A good litmus test is, "would all of these exported inner methods/fields be added directly to the outer type"; if the answer is "some" or "no", don't embed the inner type - use a field instead.

Short variable declarations (`:=`) should be used if a variable is being set to some value explicitly.

However, there are cases where the default value is clearer when the `var` keyword is used. Declaring Empty Slices, for example.

`nil` is a valid slice of length 0. This means that,
- You should not return a slice of length zero explicitly. Return `nil` instead.
- To check if a slice is empty, always use `len(s) == 0`. Do not check for `nil`.
- The zero value (a slice declared with `var`) is usable immediately without `make()`.

Remember that, while it is a valid slice, a nil slice is not equivalent to an allocated slice of length 0 - one is nil and the other is not - and the two may be treated differently in different situations (such as serialization).

Where possible, reduce scope of variables and constants. Do not reduce the scope if it conflicts with Reduce Nesting.

If you need a result of a function call outside of the if, then you should not try to reduce the scope.

Constants do not need to be global unless they are used in multiple functions or files or are part of an external contract of the package.

Naked parameters in function calls can hurt readability. Add C-style comments (`/* ... */`) for parameter names when their meaning is not obvious.

Better yet, replace naked `bool` types with custom types for more readable and type-safe code. This allows more than just two states (true/false) for that parameter in the future.

Go supports raw string literals, which can span multiple lines and include quotes. Use these to avoid hand-escaped strings which are much harder to read.

You should almost always specify field names when initializing structs. This is now enforced by `go vet`.
Exception: Field names *may* be omitted in test tables when there are 3 or fewer fields.

When initializing structs with field names, omit fields that have zero values unless they provide meaningful context. Otherwise, let Go set these to zero values automatically.

This helps reduce noise for readers by omitting values that are default in that context. Only meaningful values are specified.
Include zero values where field names provide meaningful context. For example, test cases in Test Tables can benefit from names of fields even when they are zero-valued.

When all the fields of a struct are omitted in a declaration, use the `var` form to declare the struct.

This differentiates zero valued structs from those with non-zero fields similar to the distinction created for map initialization, and matches how we prefer to declare empty slices.

Use `&T{}` instead of `new(T)` when initializing struct references so that it is consistent with the struct initialization.

Prefer `make(..)` for empty maps, and maps populated programmatically. This makes map initialization visually distinct from declaration, and it makes it easy to add size hints later if available.
Bad
Declaration and initialization are visually similar.
Good
Declaration and initialization are visually distinct.
Where possible, provide capacity hints when initializing maps with `make()`.
On the other hand, if the map holds a fixed list of elements, use map literals to initialize the map.

The basic rule of thumb is to use map literals when adding a fixed set of elements at initialization time, otherwise use `make` (and specify a size hint if available).

If you declare format strings for `Printf`-style functions outside a string literal, make them `const` values.
This helps `go vet` perform static analysis of the format string.

When you declare a `Printf`-style function, make sure that `go vet` can detect it and check the format string.
This means that you should use predefined `Printf`-style function names if possible. `go vet` will check these by default.
If using the predefined names is not an option, end the name you choose with f: `Wrapf`, not `Wrap`. `go vet` can be asked to check specific `Printf`-style names but they must end with f.

Table-driven tests with subtests can be a helpful pattern for writing tests to avoid duplicating code when the core test logic is repetitive.
If a system under test needs to be tested against *multiple conditions* where certain parts of the the inputs and outputs change, a table-driven test should be used to reduce redundancy and improve readability.

Test tables make it easier to add context to error messages, reduce duplicate logic, and add new test cases.
We follow the convention that the slice of structs is referred to as `tests` and each test case `tt`. Further, we encourage explicating the input and output values for each test case with `give` and `want` prefixes.

Table tests can be difficult to read and maintain if the subtests contain conditional assertions or other branching logic. Table tests should **NOT** be used whenever there needs to be complex or conditional logic inside subtests (i.e. complex logic inside the `for` loop).
Large, complex table tests harm readability and maintainability because test readers may have difficulty debugging test failures that occur.
Table tests like this should be split into either multiple test tables or multiple individual `Test...` functions.
Some ideals to aim for are:
* Focus on the narrowest unit of behavior
* Minimize "test depth", and avoid conditional assertions (see below)
* Ensure that all table fields are used in all tests
* Ensure that all test logic runs for all table cases
In this context, "test depth" means "within a given test, the number of successive assertions that require previous assertions to hold" (similar to cyclomatic complexity). Having "shallower" tests means that there are fewer relationships between assertions and, more importantly, that those assertions are less likely to be conditional by default.
Concretely, table tests can become confusing and difficult to read if they use multiple branching pathways (e.g. `shouldError`, `expectCall`, etc.), use many `if` statements for specific mock expectations (e.g. `shouldCallFoo`), or place functions inside the table (e.g. `setupMocks func(*FooMock)`).
However, when testing behavior that only changes based on changed input, it may be preferable to group similar cases together in a table test to better illustrate how behavior changes across all inputs, rather than splitting otherwise comparable units into separate tests and making them harder to compare and contrast.
If the test body is short and straightforward, it's acceptable to have a single branching pathway for success versus failure cases with a table field like `shouldErr` to specify error expectations.

This complexity makes it more difficult to change, understand, and prove the correctness of the test.
While there are no strict guidelines, readability and maintainability should always be top-of-mind when deciding between Table Tests versus separate tests for multiple inputs/outputs to a system.

Parallel tests, like some specialized loops (for example, those that spawn goroutines or capture references as part of the loop body), must take care to explicitly assign loop variables within the loop's scope to ensure that they hold the expected values.

In the example above, we must declare a `tt` variable scoped to the loop iteration because of the use of `t.Parallel()` below. If we do not do that, most or all tests will receive an unexpected value for `tt`, or a value that changes as they're running.

Functional options is a pattern in which you declare an opaque `Option` type that records information in some internal struct. You accept a variadic number of these options and act upon the full information recorded by the options on the internal struct.
Use this pattern for optional arguments in constructors and other public APIs that you foresee needing to expand, especially if you already have three or more arguments on those functions.

Bad
The cache and logger parameters must always be provided, even if the user wants to use the default.

Good
Options are provided only if needed.

Our suggested way of implementing this pattern is with an `Option` interface that holds an unexported method, recording options on an unexported `options` struct.

Note that there's a method of implementing this pattern with closures but we believe that the pattern above provides more flexibility for authors and is easier to debug and test for users. In particular, it allows options to be compared against each other in tests and mocks, versus closures where this is impossible. Further, it lets options implement other interfaces, including `fmt.Stringer` which allows for user-readable string representations of the options.
