package types

type DefaultHeaders []string

// EraseEntry nulls key in case it is presented, if not - nothing happens
func (d DefaultHeaders) EraseEntry(key string) {
	if len(d) == 0 {
		return
	}

	for i := 0; i < len(d); i += 2 {
		if d[i] == key {
			d[i] = ""
		}
	}
}

func (d DefaultHeaders) Copy(into []string) {
	copy(into, d)
}
