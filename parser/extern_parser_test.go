package parser

import (
	"testing"
)

func TestParserExtern(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid extern fun",
			input:   "extern 'native' fun f(a, b)",
			wantErr: false,
		},
		{
			name:    "valid extern let",
			input:   "extern 'native' let v",
			wantErr: false,
		},
		{
			name:    "multiple valid extern statements",
			input:   "extern 'native' fun f1()\nextern 'native' let v1\nextern 'native' fun f2()",
			wantErr: false,
		},
		{
			name:    "invalid extern placement",
			input:   "let x = 1\nextern 'native' let v",
			wantErr: true,
		},
		{
			name:    "invalid extern syntax - missing literal",
			input:   "extern fun f()",
			wantErr: true,
		},
		{
			name:    "invalid extern syntax - missing fun/let",
			input:   "extern 'native' f()",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]rune(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
