package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

const testToken = "123456:test-bot-token"

func signedWith(token string, fields map[string]string) url.Values {
	pairs := make([]string, 0, len(fields))
	q := url.Values{}
	for k, v := range fields {
		pairs = append(pairs, k+"="+v)
		q.Set(k, v)
	}
	sort.Strings(pairs)
	secret := sha256.Sum256([]byte(token))
	q.Set("hash", hex.EncodeToString(hmacSHA256(secret[:], strings.Join(pairs, "\n"))))
	return q
}

func validFields(now time.Time) map[string]string {
	return map[string]string{
		"id":         "42",
		"first_name": "Ann",
		"username":   "ann",
		"photo_url":  "https://t.me/i/userpic/320/ann.jpg",
		"auth_date":  strconv.FormatInt(now.Unix(), 10),
	}
}

func TestVerifyLogin(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)

	tampered := signedWith(testToken, validFields(now))
	tampered.Set("first_name", "Bob")

	wrongHash := signedWith(testToken, validFields(now))
	wrongHash.Set("hash", "00"+wrongHash.Get("hash"))

	noHash := signedWith(testToken, validFields(now))
	noHash.Del("hash")

	noName := validFields(now)
	delete(noName, "first_name")

	badID := validFields(now)
	badID["id"] = "-1"

	duplicate := signedWith(testToken, validFields(now))
	duplicate.Add("id", "1")

	tests := []struct {
		name string
		in   url.Values
		ok   bool
	}{
		{"valid", signedWith(testToken, validFields(now)), true},
		{"empty", url.Values{}, false},
		{"tampered field", tampered, false},
		{"wrong hash", wrongHash, false},
		{"no hash", noHash, false},
		{"stale auth_date", signedWith(testToken, validFields(now.Add(-25*time.Hour))), false},
		{"auth_date in the future", signedWith(testToken, validFields(now.Add(time.Hour))), false},
		{"missing first_name", signedWith(testToken, noName), false},
		{"non-positive id", signedWith(testToken, badID), false},
		{"duplicate key", duplicate, false},
		{"signed by another token", signedWith("999:other", validFields(now)), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u, err := verifyLogin(tc.in, testToken, now)
			if tc.ok {
				if err != nil {
					t.Fatalf("want ok, got %v", err)
				}
				if u.ID != 42 || u.FirstName != "Ann" || u.Username != "ann" {
					t.Fatalf("unexpected user %+v", u)
				}
				return
			}
			if err == nil {
				t.Fatalf("want rejection, got user %+v", u)
			}
		})
	}
}

func TestPrimaryLanguage(t *testing.T) {
	for in, want := range map[string]string{
		"ru-RU,ru;q=0.9,en;q=0.8": "RUB",
		"de":                      "EUR",
		"":                        "USD",
		"xx-YY":                   "USD",
	} {
		if got := guessCurrency(primaryLanguage(in)); got != want {
			t.Errorf("%q: got %s, want %s", in, got, want)
		}
	}
}
