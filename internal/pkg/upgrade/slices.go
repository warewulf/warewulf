package upgrade

func indexOf[T comparable](slice []T, item T) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}

func replaceSliceElement[T any](original []T, index int, replacement []T) []T {
	if index < 0 || index >= len(original) {
		return original
	}
	return append(original[:index], append(replacement, original[index+1:]...)...)
}
