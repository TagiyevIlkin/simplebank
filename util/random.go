package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

const alphabet = "abcdefghijklmnopqrstuvwxyz"

// RandomInt generaters a random integer between min and max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString generaters a random string of lenght n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// RandomOwner generaters a random owner name
func RandomOwner() string {
	return RandomString(8)
}

// RandomMoney generaters a random amount of money
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomCurrency generaters a random currency name
func RandomCurrency() string {
	currencies := []string{USD, EUR, AZN}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}

// RandomEmail generaters a random email
func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(6))
}
