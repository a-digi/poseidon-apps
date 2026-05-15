package repko

import "crypto/rand"

const codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
const codeLength = 6

// NewCode returns a 6-char room code from a 32-character ambiguity-free alphabet.
// I, O, 0, 1 are excluded so codes are readable on phones and over the air.
func NewCode() string {
	buf := make([]byte, codeLength)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	for i, b := range buf {
		buf[i] = codeAlphabet[int(b)%len(codeAlphabet)]
	}
	return string(buf)
}
