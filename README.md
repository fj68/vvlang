# vv programming language

_This project is still in __early stage of development__. Nothing useful for end-users._

## Features Implemented

 - simple and enough, friendly syntax
 - bool, number, string (with interpolation), list and record
 - sorry, we have `null` too
 - function, if-else-end, while, defer and variables
 - modular system with `import` and `pub`

## Code Example

### How it looks

```vv
import console from 'std/console.vv'

// let's define very simple function
fun add(a, b)
  return a + b
end

// define variable
let x = 0

// function is a value
let incr_x = fun ()
  // variables are mutable
  x += 1
end

// we have if-end and if-else-end
if x < 10
  console.print('x is less than 10.')
end

// we have while, break and continue too
// but, we don't have for or for-in
let i = 0
while i < 10
  // call some functions
  incr_x()
  i += 1
end

let name = "world"
console.print("Hello, {name}!")

console.print(x)  // 11
```

<!--

### CSV Parser

Currently, there are some missing features e.g. list and the code below won't run.

```vv
import list from 'std/list.vv'
import string from 'std/string.vv'

fun is_space(c)
  return c == ' ' or
         c == '\t' or
         c == '\r' or
         c == '\n'
end

fun trim_start(text)
  let pos = 0
  while pos < string.length(text) and is_space(text[pos]) do
    pos += 1
  end
  return text[pos:]
end

fun parse_line(line)
  let cols = []
  let cells = string.split(line, ',')
  while 0 < list.length(cells) do
    let cell = list.shift(cells)
    list.push(cols, trim_start(cell))
  end
  return cols
end

fun parse_csv(text)
  let rows = []
  let lines = string.split(text, '\n')
  while 0 < list.length(lines) do
    let line = list.shift(lines)
    list.push(rows, parse_line(line))
  end
  return rows
end

fun print_csv(rows)
  while 0 < list.length(rows) do
    let cols = shift(rows)
    print(join(cols, "\n"))
  end
end

let text = file.read_all("test.csv")
let values = parse_csv(text)
print_csv(values)
```

### Shooting

Graphic API and related works are future plans.
It's not currently available and the API may change.

```vv
// sprite module
fun new(name, x, y, w, h, face)
  return {
    name = name,
    x = x,
    y = y,
    w = w,
    h = h,
    face = face,
  }
end

fun move_by(sprite, dx, dy)
  sprite.x = sprite.x + dx
  sprite.y = sprite.y + dy
end

fun _between(min, v, max)
  return min <= v and v < max
end

fun collide_with(sprite, other)
  return (
    _between(other.x, sprite.x, other.x+other.w) or _between(other.x, sprite.x+sprite.w, other.x+other.w)
  ) and (
    _between(other.y, sprite.y, other.y+other.h) or _between(other.y, sprite.y+sprite.h, other.y+other.h)
  )
end

fun draw(sprite)
    set_pos(sprite.x, sprite.y)
    draw_image("{sprite.name}_{sprite.face}.png")
end
```

```
let sprite = import("sprite")

let player = sprite.new('player', 0, 0, 50, 100 'left')
let bullets = []

fun fire()
  push(bullets, sprite.new('bullet', player.x, player.y, 20, 5 player.face))
end

fun update_bullet(i)
  let bullet = bullets[i]
  if bullet.face == 'left' do
    sprite.move_by(bullet, -1, 0)
  else
    sprite.move_by(bullet, 1, 0)
  end
  
  if sprite.collide_with(bullet, player) do
    remove(bullets, i)
  end
end

let holding = false

fun update()
  if get_key('left') do
    sprite.move_by(player, -1, 0)
    player.face = 'left'
  end
  if get_key('right') do
    sprite.move_by(player, 1, 0)
    player.face = 'right'
  end
  if not(holding) and get_key('space') or get_key('up') do
    holding = true
    fire()
  end
  if holding and not(get_key('space') or get_key('up')) do
    holding = false
  end
  
  let i = 0
  while i < len(bullets) do
    update_bullet(i)
    i = i + 1
  end
end

fun draw()
  clear(255, 255, 255)
  
  sprite.draw(player)
  
  let i = 0
  while i < len(bullets) do
    sprite.draw(bullets[i])
    i = i + 1
  end
end

while true do
  update()
  draw()
  wait(0.1)
end
```
-->

## Language

### Value types

 - null - `null`
 - bool - `true` and `false`
 - int - `5`, `-8`
 - float - `0.4`, `-8.2`
 - string - `'single quoted'` and `"double {quoted}"`
 - function - `fun name(arg) return 'fun' end`
 - list - `[3, true, 'item']`
 - record - `{ name = 'value', key = 8 }`

### String interpolation

Double quoted strings support interpolation using `{}`.

```vv
import console from 'std/console.vv'

let name = "world"
console.print("Hello, {name}!") // Hello, world!

let a = 1
let b = 2
console.print("{a} + {b} = {a+b}") // 1 + 2 = 3
```

To escape brackets, use `{{` or `}}`.

```vv
import console from 'std/console.vv'
console.print("{{brackets}}") // {brackets}
```

### Variables

```vv
## definition
let variable_x = true

## update
variable_x = false
```

Variables are mutable and dynamically typed.

#### Record destructuring

You can destructure records using `let { ... } = record` syntax.

```vv
import console from 'std/console.vv'

let r = { name = "vv", version = 1 }
let { name, version as v } = r

console.print(name) // vv
console.print(v)    // 1
```

#### Block scope

Variable scopes are enclosed within 'block's.

```vv
import console from 'std/console.vv'

let x = 0

// use block to create new scope
begin
  // this variable is available only within this block
  let y = 1

  // outer variable is accessible
  console.print(x)  // it prints '0'

  // shadowing (only availabe in this block)
  let x = 3
  console.print(x)  // now it prints '3'
end

console.print(x)  // it prints '0'
console.print(y)  // variable 'y' is not defined
```

### if/else

```vv
import console from 'std/console.vv'

if c == 'a'
  console.print('char is letter \'a\'')
end
```

```vv
import console from 'std/console.vv'

if c == ' ' or c == '\t'
  console.print('char is a space.')
else
  console.print('char is not a space.')
end
```

`if` creates block (separated variable scope).

```vv
import console from 'std/console.vv'

if x < 10
  let y = 11
end

console.print(y)  // 'y' is not defined
```

To avoid this, declare variable before `if`

```vv
let y = null
if x < 10
  y = 11
end
```

#### Conditional operators

 - `==` - equal to
 - `<` - less than
 - `<=` - less than or equal to

To negate the result of condition, use `not` syntax.

```vv
import console from 'std/console.vv'

if not(c == ' ')
  console.print('char is not a space.')
end
```

### while

```vv
import console from 'std/console.vv'

let i = 0
while i < 10
  i = i + 1
end

console.print(i)
```

`break` / `continue` is also available.

While creates new variable scope.

### Functions

```vv
import console from 'std/console.vv'

fun incr(x)
  return x + 1
end

console.print(incr(5))  // 6
```

Function is a value. Lambda functions are also supported.

```vv
import console from 'std/console.vv'

fun incr(x)
  return x + 1
end

let apply = fun(v, f)
  return f(v)
end

console.print(apply(5, incr))  // 6
```

### defer

Like Swift (and unlike Go), `defer` in vv is block scoped.

So the following `io.close(f)` will be called per loop, not at the end of function.

```vv
import console from 'std/console.vv'
import io from 'std/io.vv'
import list from 'std/list.vv'

fun do_something_on(files)
  let i = 0
  while i < list.length(files)
    let { value as f, error } = io.open('test.txt', 'rt')
    if not(error == null)
      console.print(error)
      break
    end
    // this will be called per loop
    defer io.close(f)

    // ...do something...

    i += 1
  end
end
```

### Builtin Syntax

 - `not(value)` - negate boolean `value`
 - `type(value)` - get the type name of `value` (e.g. `"int"`, `"float"`, `"string"`, `"bool"`, `"list"`, `"record"`, `"fun"`, `"null"`)
 - `str(value)` - explicit conversion of `value` to string

## Module System

vv has a simple module system based on files.

#### Exports

Use `pub` keyword to export variables or functions.

```vv
// math.vv
pub let pi = 3.14
pub fun square(x)
  return x * x
end

// non-pub values are private to the module
let secret = 42
```

#### Imports

Use `import alias from 'path'` syntax.

```vv
import console from 'std/console.vv'
import math from './math.vv'

console.print(math.pi)
console.print(math.square(2))
```

#### Import resolution

1. **Local**: Paths starting with `./` or `../` are resolved relative to the current file.
2. **Vendored**: If not local, vv looks in the `.vv-modules` directory in the project root.
3. **Global**: Finally, it looks in `$VVPATH/.cache`.

#### Remote modules

You can import modules from GitHub or other supported domains.

```vv
import list from 'github.com/user/repo@v1.0.0/std/list.vv'
```

Use `vv get <path>` to download remote modules. You can fix the version of modules in your project's `.vv-modules` directory using `vv vendor`.


#### Standard Library

Commonly used modules:

- `std/list`: `length(l)`, `push(l, v)`, `map(l, f)`, etc.
- `std/string`: `split(s, sep)`, `length(s)`, etc.
- `std/math`: `sqrt(n)`, `abs(n)`, etc.

#### Calling Go code (Native Interop)

You can link to Go functions using `extern "native"`.

```vv
pub extern "native" fun length(l)
```

## Embedding

You can embed `vvlang` into your Go projects.

```go
package main

import (
    "fmt"
    "github.com/fj68/vvlang/interp"
    "github.com/fj68/vvlang/lib"
)

func main() {
    // 1. Create a new state
    s := interp.NewState("main.vv")

    // 2. Register standard library
    s.RegisterBuiltinModules(lib.Std.Natives)
    s.EnsureSystemLibrary(lib.Std.Name, lib.Std.FS)

    // 3. Register your own native function
    s.RegisterNative("hello", interp.VBuiltinFun(func(s *interp.State, args []interp.Value) (interp.Value, error) {
        fmt.Println("Hello from Go!")
        return interp.VNull{}, nil
    }))

    // 4. Evaluate code
    code := `
        import console from 'std/console.vv'
        extern "native" fun hello()
        hello()
    `
    s.Eval([]rune(code))
}
```

## Development

Assuming latest golang is installed:

```sh
# clone the repo
git clone git@github.com:fj68/vvlang.git
cd vvlang
# install dependent modules
go mod tidy
# build executable
go build -o vv
# run interpreter
vv ./test.vv
```

