package encoding

import "unicode/utf8"

func ToUTF8(content string) []byte {
	b := make([]byte, len(content))
	i := 0
	for _, r := range content {
		i += utf8.EncodeRune(b[i:], r)
	}

	return b[:i]
}
