package utils

func CheckAndExit(err error) {
	if err != nil {
		panic(err)
	}
}
