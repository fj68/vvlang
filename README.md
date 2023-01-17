# new-lang

## Features Planned

 - simple and enough, friendly syntax
 - bool, number and string
 - function, if-else-end, while, for-in and variables

## Code Example

### CSV Parser

```new-lang
fun is_space(c)
  return c == ' ' or
         c == '\t' or
         c == '\r' or
         c == '\n'
end

fun trim_start(text)
  var pos = 0
  while pos < len(text) and is_space(text[pos]) do
    pos = pos + 1
  end
  return slice(text, pos, len(text))
end

fun split(text, sep)
  var lines = []
  var pos = 0
  var start = 0
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
  var cols = []
  var cells = split(line, ',')
  while 0 < len(cells) do
    push(cols, trim_start(shift(cells)))
  end
  return cols
end

fun parse_csv(text)
  var rows = []
  var lines = split(text, '\n')
  while 0 < len(lines) do
    push(rows, parse_line(shift(lines)))
  end
  return rows
end

fun print_csv(rows)
  while 0 < len(rows) do
    var cols = shift(rows)
    print(join(cols, "\n"))
  end
end

var text = read_file("test.csv")
var values = parse_csv(text)
print_csv(values)
```

### Shooting

```new-lang
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
var sprite = import("sprite")

var player = sprite.new('player', 0, 0, 50, 100 'left')
var bullets = []

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

var holding = false

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
