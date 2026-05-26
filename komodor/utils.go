package komodor

func expandStringList(raw []interface{}) []string {
	result := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok && s != "" {
			result = append(result, s)
		}
	}
	return result
}
