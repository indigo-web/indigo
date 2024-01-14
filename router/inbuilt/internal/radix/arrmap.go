package radix

type segmentEntry struct {
	Key  string
	Node *Node
}

// loadfactor defines threshold, exceeding which one results in fallback to
// hashmap
const loadfactor = 30

// arrMap is a wrapper of standard hashmap, however on small amount of entries (usually <40)
// it uses an array instead. This is related to the constant lookup time by a hashmap, that
// sometimes may be pretty expensive
type arrMap struct {
	arrOverflow bool
	arr         []segmentEntry
	m           map[string]*Node
}

func (a *arrMap) Lookup(key string) *Node {
	if !a.arrOverflow {
		for _, entry := range a.arr {
			if entry.Key == key {
				return entry.Node
			}
		}

		return nil
	}

	return a.m[key]
}

func (a *arrMap) Add(key string, node *Node) {
	if a.arrOverflow {
		a.m[key] = node
		return
	}

	if len(a.arr)+1 > loadfactor {
		a.arrOverflow = true
		a.escapeToMap()
		a.m[key] = node
		return
	}

	a.arr = append(a.arr, segmentEntry{key, node})
}

func (a *arrMap) escapeToMap() {
	a.m = make(map[string]*Node, len(a.arr)+1)

	for _, entry := range a.arr {
		a.m[entry.Key] = entry.Node
	}
}
