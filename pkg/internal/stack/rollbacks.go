package stack

type Rollbacks []func()

func (r Rollbacks) Add(f func()) Rollbacks {
	return append(r, f)
}

func (r Rollbacks) Run() {
	for i := len(r) - 1; i >= 0; i-- {
		r[i]()
	}
}
