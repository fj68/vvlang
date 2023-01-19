# vv programming language

_This project is still in __early stage of development__. Nothing useful for end-users._

## Features Planned

 - simple and enough, friendly syntax
 - bool, number, string, list and struct
 - function, if-else-end, while and variables

## Code Example

### How it looks

```vv
// let's define very simple function
fun add(a, b)
  return a + b
end

// define variable
x = 0

// here, another simple function
fun incr_x()
  // variables are mutable
  x = x + 1
end

if x < 10
  print('x is less than 10.')
end

// call some functions
incr_x()
print(x)  // 1
```

<!--

### CSV Parser

Currently, there are some missing features e.g. list and the code below won't run.

```vv
fun is_space(c)
  return c == ' ' or
         c == '\t' or
         c == '\r' or
         c == '\n'
end

fun trim_start(text)
  pos = 0
  while pos < len(text) and is_space(text[pos]) do
    pos = pos + 1
  end
  return slice(text, pos, len(text))
end

fun split(text, sep)
  lines = []
  pos = 0
  start = 0
  while pos < len(text) do
    if slice(text, pos, pos+len(sep)) == sep do
      push(lines, slice(text, start, pos))
      pos = pos + len(sep)
      start = pos
    else
      pos = pos + 1
    end
  end
  return lines
end

fun parse_line(line)
  cols = []
  cells = split(line, ',')
  while 0 < len(cells) do
    push(cols, trim_start(shift(cells)))
  end
  return cols
end

fun parse_csv(text)
  rows = []
  lines = split(text, '\n')
  while 0 < len(lines) do
    push(rows, parse_line(shift(lines)))
  end
  return rows
end

fun print_csv(rows)
  while 0 < len(rows) do
    cols = shift(rows)
    print(join(cols, "\n"))
  end
end

text = read_file("test.csv")
values = parse_csv(text)
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
sprite = import("sprite")

player = sprite.new('player', 0, 0, 50, 100 'left')
bullets = []

fun fire()
  push(bullets, sprite.new('bullet', player.x, player.y, 20, 5 player.face))
end

fun update_bullet(i)
  bullet = bullets[i]
  if bullet.face == 'left' do
    sprite.move_by(bullet, -1, 0)
  else
    sprite.move_by(bullet, 1, 0)
  end
  
  if sprite.collide_with(bullet, player) do
    remove(bullets, i)
  end
end

holding = false

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
  
  i = 0
  while i < len(bullets) do
    update_bullet(i)
    i = i + 1
  end
end

fun draw()
  clear(255, 255, 255)
  
  sprite.draw(player)
  
  i = 0
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

 - bool - `true` and `false`
 - number - `5`, `0.4`, `-8.2`
 - string - `'this is string'`
 - function - `fun name(arg) return 'fun' end`
 - array (not implemented) - `[3, true, 'item']`
 - struct (not implemented) - `{ name = 'value', key = 8 }`

### Variables

```vv
variable_x = true
```

Variables are mutable and dynamically typed.

### If

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

#### Conditional operators

 - `==` - equal to
 - `<` - less than
 - `<=` - less than or equal to

To negate the result of condition, use builtin function `not()`.

```vv
if not(c == ' ')
  print('char is not a space.')
end
```

### While

```vv
i = 0
while is_eof()
  i = i + 1
end

print(i)
```

`break` / `continue` will be available (not implemented yet).

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

apply = fun(v, f)
  return f(v)
end

print(apply(5, incr))  // 6
```

### Builtin Functions

 - `not(value)` - negate boolean `value`
 - `print(value)` - print out the `value` (will be replaced with `draw_text(string)`)
 - `get_type(value)` - get the type of `value` (will be removed)
 - `len(array)` - get the size of `array` (not implemented)
 - `bool(value)` - convert the `value` to bool
 - `number(value)` - convert the `value` to bool
 - `floor(number)` - floor the `number` to int
 - `ceil(number)` - ceil the `number` to int
 - `string(value)` - convert the `value` to string

## Development

Assuming latest golang is installed:

```sh
go build -o vv
```

