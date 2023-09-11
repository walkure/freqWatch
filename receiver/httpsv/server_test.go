package httpsv

import (
	"fmt"
	"testing"
)

func Test_getListener(t *testing.T) {
	type args struct {
		defaultListener string
		defaultPort     uint16
		envListener     string
		envPort         string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"default host(INADDR_ANY)/port none envrions",
			args{"", 8080, "", ""},
			":8080",
		},
		{
			"default host(specified)/port none envrions",
			args{"foo", 8080, "", "0"},
			"foo:8080",
		},
		{
			"default host(INADDR_ANY)/port non host env.",
			args{"", 8080, "", "8081"},
			":8081",
		},
		{
			"default host(INADDR_ANY)/port envrions exists",
			args{"", 8080, "hoge", "8081"},
			"hoge:8081",
		},
		{
			"default host(INADDR_ANY)/port envrions overflowerd",
			args{"", 8080, "hoge", "114514"},
			"hoge:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.envListener != "" {
				t.Setenv("LISTEN", tt.args.envListener)
			}
			if tt.args.envPort != "" {
				t.Setenv("PORT", tt.args.envPort)
			}
			if got := getListener(tt.args.defaultListener, tt.args.defaultPort); got != tt.want {
				t.Errorf("getListener() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_initDumper(t *testing.T) {
	type args struct {
		defaultSize int
		envSize     int
		keyName     string
		envName     string
	}
	tests := []struct {
		name     string
		args     args
		wantSize int
	}{
		{
			"default",
			args{
				defaultSize: 13,
				envSize:     0,
				keyName:     "DUMPER_SIZE",
				envName:     "",
			},
			13,
		},
		{
			"env changed",
			args{
				defaultSize: 13,
				envSize:     213,
				keyName:     "DUMPER_SIZE",
				envName:     "DUMPER_SIZE",
			},
			213,
		},
		{
			"env not match",
			args{
				defaultSize: 13,
				envSize:     213,
				keyName:     "HOGE_SIZE",
				envName:     "DUMPER_SIZE",
			},
			13,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.envName != "" {
				t.Setenv(tt.args.envName, fmt.Sprint(tt.args.envSize))
			}
			dumper := initDumper(tt.args.defaultSize, tt.args.keyName)
			rb := dumper.GetRingBuffer("test")
			if rb.Length() != tt.wantSize {
				t.Errorf("initDumper(%d,%s) = %d, want %d", tt.args.defaultSize, tt.args.keyName, rb.Length(), tt.wantSize)
			}
		})
	}
}
