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

// Name returns the first Data matching the name.
func (f Form) Name(name string) (Data, bool) {
	for data := range f.Names(name) {
		return data, true
	}

	return Data{}, false
}

// Names returns an iterator over all Data matching the name.
func (f Form) Names(name string) iter.Seq[Data] {
	return func(yield func(Data) bool) {
		for _, entry := range f {
			if entry.Name == name {
				if !yield(entry) {
					break
				}
			}
		}
	}
}

// File returns the first Data matching the filename.
func (f Form) File(name string) (Data, bool) {
	for data := range f.Files(name) {
		return data, true
	}

	return Data{}, false
}

// Files returns an iterator over all Data matching the filename.
func (f Form) Files(name string) iter.Seq[Data] {
	return func(yield func(Data) bool) {
		for _, entry := range f {
			if entry.Filename == name {
				if !yield(entry) {
					break
				}
			}
		}
	}
}
