package set

func GetSetFromStringArray(a []string) []string {
	m := make(map[string]bool)
	for _, v := range a {
		m[v] = true
	}
	s := make([]string, 0)
	for k := range m {
		s = append(s, k)
	}
	return s
}
