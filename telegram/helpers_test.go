package telegram

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_escapeURLMarkdownSymbols(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			args: struct{ msg string }{msg: "https://www.instagram.com/p/Bj2mSWCnoPl/?utm_source=ig_web_button_share_sheet"},
			want: "https://www.instagram.com/p/Bj2mSWCnoPl/?utm\\_source=ig\\_web\\_button\\_share\\_sheet",
		},
		{
			name: "success",
			args: struct{ msg string }{msg: "https://www.instagram.com/p/Bj2mSWCnoPl/?utm_source=ig_web_button_share_sheet\n test *line\nhttp://www.instagram.com/p/Bj2mSWCnoPl/?utm_source=ig_web_button_share_sheet"},
			want: "https://www.instagram.com/p/Bj2mSWCnoPl/?utm\\_source=ig\\_web\\_button\\_share\\_sheet\n test *line\nhttp://www.instagram.com/p/Bj2mSWCnoPl/?utm\\_source=ig\\_web\\_button\\_share\\_sheet",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, escapeURLMarkdownSymbols(tt.args.msg), tt.want)
		})
	}
}
