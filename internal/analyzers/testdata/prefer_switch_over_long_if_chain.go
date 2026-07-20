package p

func value(status string) int {
	if status == "new" {
		return 1
	} else if status == "active" {
		return 2
	} else if status == "closed" {
		return 3
	}
	return 0
}
