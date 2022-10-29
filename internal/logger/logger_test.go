package logger

import (
	"testing"
)

func Test_callerMarshalFunc(t *testing.T) {
	type args struct {
		pc       uintptr
		filepath string
		line     int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "full path",
			args: args{
				filepath: "/path/to/package/function.go",
				line:     1,
			},
			want: "package/function.go:1",
		},
		{
			name: "only file",
			args: args{
				filepath: "function.go",
				line:     1,
			},
			want: "function.go:1",
		},
		{
			name: "only file and package",
			args: args{
				filepath: "package/function.go",
				line:     1,
			},
			want: "package/function.go:1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := callerMarshalFunc(tt.args.pc, tt.args.filepath, tt.args.line); got != tt.want {
				t.Errorf("callerMarshalFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}
