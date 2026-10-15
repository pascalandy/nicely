package runner

import "testing"

func TestOlderThanComparesNumberByNumber(t *testing.T) {
	cases := []struct {
		found, want string
		older       bool
	}{
		{"3.12", "3.12.0", false},
		{"3.12.0", "3.12", false},
		{"1.9.9", "2.0", true},
		{"10.0", "9.9", false},
		{"6.01", "6.1", false},
		{"0.9", "1", true},
		{"123456789012345678901.0", "123456789012345678900.9", false},
	}
	for _, c := range cases {
		if got := olderThan(c.found, c.want); got != c.older {
			t.Errorf("olderThan(%s, %s) = %t, want %t", c.found, c.want, got, c.older)
		}
	}
}
