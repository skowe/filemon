package filemon

import "testing"

func TestSignalReciever_Update(t *testing.T) {
	type args struct {
		e *Event
	}
	tests := []struct {
		name string
		s    *SignalReciever
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.Update(tt.args.e)
		})
	}
}
