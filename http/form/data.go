package form

import "iter"

type Data struct {
	Name     string
	Filename string
	Type     string
	Charset  string
	Value    string
}

type Form []Data

func (f Form) Name(name string) iter.Seq[Data] {
	return func(yield func(Data) bool) {
		for _, entry := range f {
			if entry.Name == name {
				yield(entry)
			}
		}
	}
}

func (f Form) File(name string) iter.Seq[Data] {
	return func(yield func(Data) bool) {
		for _, entry := range f {
			if entry.Filename == name {
				yield(entry)
			}
		}
	}
}
