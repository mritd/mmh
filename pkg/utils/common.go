package utils

func CheckAndExit(err error) {
	if err != nil {
		panic(err)
	}
}

func ShortenString(str string, n int) string {
	if len(str) <= n {
		return str
	} else {
		return str[:n]
	}
}
