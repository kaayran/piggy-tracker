// Command devlogin prints a genuinely signed Telegram Login Widget payload, so the site can be
// signed into on localhost, where the widget does not render, without weakening the check the
// server performs.
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
	flag.Parse()

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "BOT_TOKEN is not set")
		os.Exit(1)
	}

	fields := map[string]string{
		"id":         strconv.FormatInt(*id, 10),
		"first_name": *name,
		"username":   "dev",
		"auth_date":  strconv.FormatInt(time.Now().Unix(), 10),
	}

	pairs := make([]string, 0, len(fields))
	q := url.Values{}
	for k, v := range fields {
		pairs = append(pairs, k+"="+v)
		q.Set(k, v)
	}
	sort.Strings(pairs)

	secret := sha256.Sum256([]byte(token))
	m := hmac.New(sha256.New, secret[:])
	m.Write([]byte(strings.Join(pairs, "\n")))
	q.Set("hash", hex.EncodeToString(m.Sum(nil)))
	fmt.Println(q.Encode())
}
