package crypt

import "testing"

func TestCRC32(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{"1th", args{str: "123456"}, 158520161},
		{"2th", args{str: "hello,word"}, 3812273915},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CRC32(tt.args.str); got != tt.want {
				t.Errorf("CRC32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMD5(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"1th", args{str: "123456"}, "e10adc3949ba59abbe56e057f20f883e"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MD5(tt.args.str); got != tt.want {
				t.Errorf("MD5() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSHA1(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"1th", args{"123456"}, "7c4a8d09ca3762af61e59520943dc26494f8941b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SHA1(tt.args.str); got != tt.want {
				t.Errorf("SHA1() = %v, want %v", got, tt.want)
			}
		})
	}
}
