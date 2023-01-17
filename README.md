# new-lang

## Features Planned

 - simple and enough, friendly syntax
 - bool, number and string
 - function, if-else-end, while and variables

## Code Example

```new-lang
fun is_space(c)
  return c == ' ' or
         c == '\t' or
         c == '\r' or
         c == '\n'
end

fun is_eof(text, pos)
  return len(text) <= pos
end

fun skip_whitespaces(text, pos)
  while not(is_eof(text, pos)) and is_space(text[pos]) do
    pos = pos + 1
  end
  return pos
end

fun trim_start(text)
  pos = skip_whitespaces(text)
  return slice(text, pos, len(text))
end

fun split(text, sep)
  lines = []
  pos = 0
  start = 0
  while not(is_eof(text, pos)) do
    if slice(text, pos, pos+len(sep)) == sep do
      lines = append(lines, slice(text, start, pos))
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
  i = 0
  cells = split(line, ",")
  while i < len(cells) do
    cols = append(cols, trim_start(cells[i]))
    i = i + 1
  end
  return cols
end

fun parse_csv(text)
  rows = []
  i = 0
  lines = split(text, "\n")
  while i < len(lines) do
    rows = append(rows, parse_line(lines[i]))
    i = i + 1
  end
  return rows
end

fun add_spaces(text, min)
  diff = min - len(text)
  if diff <= 0 do
    return text
  end
  return append(text, " " * diff)
end

fun print_csv(values)
  i = 0
  lines = []
  while i < len(values) do
    lines = append(join(values, ", "))
  end
  print(join(lines, "\n"))
end

text = read_file("test.csv")
values = parse_csv(text)
print_csv(values)
```
