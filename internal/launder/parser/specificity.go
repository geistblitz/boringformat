package parser

type Specificity [3]int

func (s Specificity) Less(other Specificity) bool {
	for i := range s {
		if s[i] < other[i] {
			return true
		}
		if s[i] > other[i] {
			return false
		}
	}
	return false
}

func (s Specificity) Add(other Specificity) Specificity {
	for i, sp := range other {
		s[i] += sp
	}
	return s
}
