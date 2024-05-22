package shorten

//
//import (
//	"net/url"
//	"strings"
//)
//
//const alphabet = "ynAJfoSgdXHB5VasEMtcbPCr1uNZ4LG723ehWkvwYR6KpxjTm8iQUFqz9D"
//
//func Reverse[S ~[]E, E any](s S) {
//	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
//		s[i], s[j] = s[j], s[i]
//	}
//}
//
//var alphabetLen = uint32(len(alphabet))
//
//func Shorten(id uint32) string {
//	var (
//		digits  []uint32
//		num     = id
//		builder strings.Builder
//	)
//
//	for num > 0 {
//		digits = append(digits, num%alphabetLen)
//		num /= alphabetLen
//	}
//	Reverse(digits)
//
//	for _, digit := range digits {
//		builder.WriteString(string(alphabet[digit]))
//	}
//
//	return builder.String()
//}
//
//func PrependBaseURL(baseURL, identifier string) (string, error) {
//	parsed, err := url.Parse(baseURL)
//	if err != nil {
//		return "", err
//	}
//
//	parsed.Path = identifier
//
//	return parsed.String(), nil
//}
