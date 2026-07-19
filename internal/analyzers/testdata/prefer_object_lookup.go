package p

func value(role string) bool {
	return role == "admin" || role == "owner" || role == "staff"
}
