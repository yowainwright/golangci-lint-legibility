package p

func value(users map[string]User, id string) {
	if user, ok := users[id]; ok && user.Active {
		save(user)
	}
}
