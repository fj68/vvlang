[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/fj68/vvlang)

# vv programming language

_This project is still in __early stage of development__. Nothing useful for end-users._

## Features

 - simple and enough, friendly syntax
 - bool, number, string (with interpolation), list and record
 - sorry, we have `null` too
 - function, if-else-end, while, defer and variables
 - modular system with `import` and `pub`

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

For more information about the language and tooling, see [Documentation](https://fj68.github.io/vvlang/).

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
        return interp.NoneValue, nil
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

