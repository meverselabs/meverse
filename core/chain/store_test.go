package chain

import (
	"log"
	"testing"
)

func TestStore_Version(t *testing.T) {
	SetVersion(0, 1)
	type args struct {
		h uint32
	}
	tests := []struct {
		name string
		args args
		want uint16
	}{
		{"test", args{14}, 1},
		{"test", args{15}, 1},
		{"test", args{16}, 2},
		{"test", args{17}, 4},
		{"test", args{18}, 4},
	}
	SetVersion(15, 2)
	SetVersion(16, 4)
	for _, tt := range tests {
		st := Store{}
		t.Run(tt.name, func(t *testing.T) {
			if got := st.Version(tt.args.h); got != tt.want {
				t.Errorf("Store.Version() = %v, want %v", got, tt.want)
			} else {
				log.Println(tt.args.h, got)
			}
		})
	}
}
