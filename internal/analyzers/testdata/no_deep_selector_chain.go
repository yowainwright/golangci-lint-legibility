package p

func value(config Config) bool {
	return config.User.Profile.Settings.Email.Enabled
}
