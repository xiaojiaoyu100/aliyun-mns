package alimns

import "testing"

func TestCheckQueueName(t *testing.T) {

	tests := [...]struct {
		Name string
		Want bool
	}{
		0: {
			Name: "",
			Want: false,
		},
		1: {
			Name: "w8hEny8a77VmUeOSj2IV1HiCsgbwjHdMviMsWrePDEUkt56QNJ7appr6itwJx4v3j8AUpCvbbyyNPcjgEYfhnSS0ZajuNo54YgIvYCTBHJ0fZnV0LNkDVR4ovi0639i6cNW4rwJER8r2de2buUj2Zmn8U623mX9GRTqwqhqS4KxZvU90lDwtIOOE8XjZbLFYEVfm8tQDvBOumDJffufSZaLmeQOTZ5fHxCtY6aZPzUvHmQawSsLvnVOchE8mlEHCO",
			Want: false,
		},
		2: {
			Name: "-123",
			Want: false,
		},
		3: {
			Name: "123",
			Want: true,
		},
		4: {
			Name: "a123",
			Want: true,
		},
		5: {
			Name: "a1*3",
			Want: false,
		},
		6: {
			Name: "stu_adu",
			Want: false,
		},
	}

	for _, test := range tests {
		if got := checkQueueName(test.Name); got != test.Want {
			t.Errorf("%s: want %t, got %t", test.Name, test.Want, got)
		}
	}
}
