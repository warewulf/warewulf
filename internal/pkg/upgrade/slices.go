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

func replaceOverlay(originals []string, toReplace string, replacements []string) []string {
	if indexOf(originals, toReplace) == -1 {
		return originals
	}

	lookup := make(map[string]bool)
	for _, v := range originals {
		lookup[v] = true
	}

	var newReplacements []string
	for _, v := range replacements {
		if !lookup[v] || v == toReplace {
			newReplacements = append(newReplacements, v)
		}
	}

	return replaceSliceElement(
		originals,
		indexOf(originals, toReplace),
		newReplacements)
}
