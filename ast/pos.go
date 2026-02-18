package ast

import "fmt"

type Pos struct {
	Start int
	End   int
	Line  int
	Col   int	// index of End in the current line
}

func (pos Pos) String() string {
	return fmt.Sprintf("Pos{%d, %d, %d, %d}", pos.Start, pos.End, pos.Line, pos.Col)
}

func (pos Pos) Copy() Pos {
	return Pos{
		pos.Start,
		pos.End,
		pos.Line,
		pos.Col,
	}
}

func (pos Pos) Eq(other Pos) bool {
	return pos.Start == other.Start &&
		pos.End == other.End &&
		pos.Line == other.Line &&
		pos.Col == other.Col
}
