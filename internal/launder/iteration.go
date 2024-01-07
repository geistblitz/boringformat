package launder

func (s *Selection) Each(f func(int, *Selection)) *Selection {
	for i, n := range s.Nodes {
		f(i, newSingleSelection(n, s.document))
	}
	return s
}

func (s *Selection) EachWithBreak(f func(int, *Selection) bool) *Selection {
	for i, n := range s.Nodes {
		if !f(i, newSingleSelection(n, s.document)) {
			return s
		}
	}
	return s
}

func (s *Selection) Map(f func(int, *Selection) string) (result []string) {
	for i, n := range s.Nodes {
		result = append(result, f(i, newSingleSelection(n, s.document)))
	}

	return result
}
