package lexer

import "fmt"

type Pos struct {
	Start int
	End   int
}

func (pos Pos) String() string {
	return fmt.Sprintf("Pos{%d, %d}", pos.Start, pos.End)
}

func (pos Pos) Copy() Pos {
	return Pos{
		pos.Start,
		pos.End,
	}
}

func (pos Pos) Eq(other Pos) bool {
	return pos.Start == other.Start &&
		   pos.End == other.End
}
