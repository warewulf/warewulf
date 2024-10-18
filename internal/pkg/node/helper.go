package node

/*
helper functions
*/

// string which is printed if no value is set
const NoValue = "--"

// print NoValue if string is empty
func PrintVal(in string) (out string) {
	if in != "" {
		return in
	}
	return NoValue
}
