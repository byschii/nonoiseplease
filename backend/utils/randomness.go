package utils

import "math/rand"

func GenerateRandomHexColor() string {
	color := ""
	for i := 0; i < 6; i++ {
		color += string("0123456789ABCDEF"[rand.Intn(16)])
	}

	return color
}

// return a 15 character long random string
func RandomID() string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	const length = 15
	var bytes = make([]byte, length)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}
