// Command initdata prints a genuinely signed Telegram initData string, so the app can be
// opened in a plain browser without weakening the check the server performs.
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {
	id := flag.Int64("id", 1, "telegram user id")
	name := flag.String("name", "Dev", "first name")
	lang := flag.String("lang", "en", "language code")
	flag.Parse()

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "BOT_TOKEN is not set")
		os.Exit(1)
	}

	fields := map[string]string{
		"auth_date": strconv.FormatInt(time.Now().Unix(), 10),
		"user": fmt.Sprintf(`{"id":%d,"first_name":%q,"username":"dev","language_code":%q}`,
			*id, *name, *lang),
	}

	pairs := make([]string, 0, len(fields))
	for k, v := range fields {
		pairs = append(pairs, k+"="+v)
	}
	sort.Strings(pairs)

	secret := mac([]byte("WebAppData"), token)
	sum := mac(secret, strings.Join(pairs, "\n"))

	q := url.Values{}
	for k, v := range fields {
		q.Set(k, v)
	}
	q.Set("hash", hex.EncodeToString(sum))
	fmt.Println(q.Encode())
}

func mac(key []byte, msg string) []byte {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(msg))
	return m.Sum(nil)
}
