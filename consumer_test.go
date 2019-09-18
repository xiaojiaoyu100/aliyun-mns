package alimns

import "testing"

func TestParallel(t *testing.T) {
	t.Log(Parallel())
}

func TestSetParallel(t *testing.T) {
	tests := []struct{
		In int
		Out int
	}{
		{
			      1,
				1,
		},
		{
			8,
			8,
		},
		{
				16,
				16,
		},
		{
				17,
				16,
		},
	}

	for _, ts := range tests {
		if setParallel(ts.In) != ts.Out {
			t.Errorf("setParallel: in = %d, out = %d", ts.In, ts.Out)
		}
	}
}
