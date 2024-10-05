package shorten

import (
	"math/rand"
	"strings"
	"time"
)

func RandomString(length int) string {
	// Набор символов для генерации строки
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	sb := strings.Builder{}
	sb.Grow(length)

	for i := 0; i < length; i++ {

		randomIndex := seededRand.Intn(len(charset))
		sb.WriteByte(charset[randomIndex])
	}

	return sb.String()

}
