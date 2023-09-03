package unreader

type Unreader struct {
	pending []byte
}

func (u *Unreader) PendingOr(or func() ([]byte, error)) (data []byte, err error) {
	if len(u.pending) > 0 {
		data, u.pending = u.pending, nil
		return data, nil
	}

	return or()
}

func (u *Unreader) Unread(b []byte) {
	u.pending = b
}

func (u *Unreader) Reset() {
	u.pending = nil
}
