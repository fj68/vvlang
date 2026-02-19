# vv programming language

_This project is still in __early stage of development__. Nothing useful for end-users._

## Features Implemented

 - simple and enough, friendly syntax
 - bool, number, string, list and record
 - sorry, we have `null` too
 - function, if-else-end, while and variables

## Code Example

### How it looks

```vv
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
  print('x is less than 10.')
end

// we have while, break and continue too
// but, we don't have for or for-in
let i = 0
while i < 10
  // call some functions
  incr_x()
  i += 1
end

print(x)  // 11
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

## Language Reference

### Value types

 - null - `null`
 - bool - `true` and `false`
 - number - `5`, `0.4`, `-8.2`
 - string - `'this is string'`
 - function - `fun name(arg) return 'fun' end`
 - list (array) - `[3, true, 'item']`
 - record (struct) - `{ name = 'value', key = 8 }`

### Variables

```vv
## definition
let variable_x = true

## update
variable_x = false
```

Variables are mutable and dynamically typed.

Variable scopes are enclosed within 'block's.

```vv
let x = 0

// use block to create new scope
begin
  // this variable is available only within this block
  let y = 1

  // outer variable is accessible
  print(x)  // it prints '0'

  // shadowing (only availabe in this block)
  let x = 3
  print(x)  // now it prints '3'
end

print(x)  // it prints '0'
print(y)  // variable 'y' is not defined
```

### if/else

```vv
if c == 'a'
  print('char is letter \'a\'')
end
```

```vv
if c == ' ' or c == '\t'
  print('char is a space.')
else
  print('char is not a space.')
end
```

`if` creates block (separated variable scope).

```vv
if x < 10
  let y = 11
end

print(y)  // 'y' is not defined
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

Operators for not equal (`!=`), greater than (`>`) and greater than or equal to (`>=`) is intentionally omitted.

To negate the result of condition, use builtin function `not()`.

```vv
if not(c == ' ')
  print('char is not a space.')
end
```

### while

```vv
let i = 0
while i < 10
  i = i + 1
end

print(i)
```

`break` / `continue` is also available.

While creates new variable scope.

### Functions

```vv
fun incr(x)
  return x + 1
end

print(incr(5))  // 6
```

Function is a value. Lambda functions are also supported.

```vv
fun incr(x)
  return x + 1
end

let apply = fun(v, f)
  return f(v)
end

print(apply(5, incr))  // 6
```

### return

vv allows top level `return`.

### defer

Like Swift (and unlike Go), `defer` in vv is block scoped.

So the following `io.close(f)` will be called per loop, not at the end of function.

```vv
import io from 'std/io.vv'
import list from 'std/list.vv'

fun do_something_on(files)
  let i = 0
  while i < list.length(files)
    let { value as f, error } = io.open('test.txt', 'rt')
    if not(error == null)
      print(error)
      break
    end
    // this will be called per loop
    defer io.close(f)

    // ...do something...

    i += 1
  end
end
```

### Builtin Functions

 - `not(value)` - negate boolean `value`
 - `print(value)` - print out the `value` (will be replaced with `io.print(string)`)
 - `type(value)` - get the type of `value` (will be replaced with `typeof value`)
<!--
 - `len(value)` - get the size of `value` which should be array or string
 - `bool(value)` - convert the `value` to bool
 - `number(value)` - convert the `value` to number
 - `floor(number)` - floor the `number` to int
 - `ceil(number)` - ceil the `number` to int
 - `string(value)` - convert the `value` to string
-->

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

