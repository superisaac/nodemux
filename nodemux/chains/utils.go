package chains

func resolveMap(root interface{}, path ...string) (interface{}, bool) {
	v := root
	for {
		m, ok := v.(map[string]interface{})
		if !ok {
			return nil, false
		}
		hop := path[0]
		path = path[1:]

		k, ok := m[hop]
		if len(path) <= 0 {
			return k, ok
		} else {
			if !ok {
				return nil, false
			}
			v = k
		}
	}
}
