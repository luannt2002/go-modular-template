package config

import "testing"

func TestChainFirstNonEmptyWins(t *testing.T) {
	c := Chain{mapSource{"A": "", "B": "from-first"}, mapSource{"A": "from-second", "B": "x", "C": "c"}}
	for k, want := range map[string]string{"A": "from-second", "B": "from-first", "C": "c"} {
		if got, _ := c.Lookup(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
	if _, ok := c.Lookup("Z"); ok {
		t.Error("Z should be missing")
	}
}
