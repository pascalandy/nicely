package runner

// Capture keeps what a program writes, up to Limit bytes, and drops the
// rest, so a noisy program never fills the memory of ncly. A write never
// fails, so the program never waits on a full pipe. Read it once Run
// returns, because Run stops every copy before it returns.
type Capture struct {
	Limit int
	kept  []byte
	over  bool
}

// Write keeps what fits under Limit and drops the rest.
func (c *Capture) Write(p []byte) (int, error) {
	n := len(p)
	if room := max(c.Limit-len(c.kept), 0); len(p) > room {
		p, c.over = p[:room], true
	}
	c.kept = append(c.kept, p...)
	return n, nil
}

// Bytes returns what Capture kept.
func (c *Capture) Bytes() []byte { return c.kept }

// Over reports whether the program wrote more than Limit bytes, so that
// Bytes holds only the first ones.
func (c *Capture) Over() bool { return c.over }
