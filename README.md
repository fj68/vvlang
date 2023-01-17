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

fun skip_whitespaces(pos)
  while not(is_eof(text, pos)) do
    if not(is_space(text[pos])) do
      break
    end
    pos = pos + 1
  end
  return pos
end
```
