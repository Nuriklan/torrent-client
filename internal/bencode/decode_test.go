package bencode

import (
	"reflect"
	"testing"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    any
		wantErr bool
	}{
		{"int", "i42e", 42, false},
		{"negative int", "i-15e", -15, false},
		{"string", "4:spam", "spam", false},
		{"list", "l4:spami42ee", []any{"spam", 42}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Decoder{data: []byte(tt.input)}
			got, err := d.Decode()
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
