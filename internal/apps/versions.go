package apps

func GetVersionIfItsNotNilAndLatest(ver *string, defaultVer string) string {
	if ver == nil {
		return defaultVer
	}
	if *ver == "latest" {
		return defaultVer
	}
	return *ver
}
