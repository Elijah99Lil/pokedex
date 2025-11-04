package main

import "testing"

func TestCaseInput(t *testing.T) {
	cases := []struct {
	input    string
	expected []string
}{
	{
		input:    "  hello  world  ",
		expected: []string{"hello", "world"},
	},
	{
		input:	 "	Sonic the Hedgehog!	",
		expected: []string{"sonic", "the", "hedgehog!"},
	},
	{
		input:   "  SUPER SONIC STYLE!	",
		expected: []string{"super", "sonic", "style!"},
	},
}

	for _, c := range cases {
		actual := cleanInput(c.input)
		if len(actual) != len(c.expected) {
			t.Errorf("lengths don't match: '%v' vs '%v'", actual, c.expected)
			continue
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
				t.Errorf("cleanInput(%v) == %v, expected %v", c.input, actual, c.expected)
			}
		}
	}
}