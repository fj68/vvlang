package docstring

import (
	"strings"

	"github.com/fj68/vvlang/lexer"
)

type Parser struct {
	tokens []*lexer.Token
	pos    int
}

func (p *Parser) Parse(tokens []*lexer.Token) (map[string]string, error) {
	p.tokens = tokens
	p.pos = 0

	docs := make(map[string]string)
	lang := "en"

	for p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		p.pos++

		if tok.Type != lexer.TDocstring {
			continue
		}

		line := tok.Text
		if strings.HasPrefix(line, "@lang ") {
			lang = strings.TrimSpace(line[6:])
			continue
		}
		docs[lang] += line + "\n"
	}

	for k, v := range docs {
		docs[k] = strings.TrimRight(v, "\n")
	}

	if len(docs) == 0 {
		return nil, nil
	}

	return docs, nil
}
