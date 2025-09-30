// Based on https://github.com/emersion/go-message/blob/b9039e0d248fca44779a1653069a855e309d6d18/mail/header.go
// SPDX-License-Expression: MIT
// Copyright (c) 2016 emersion

package mail

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func base36(input uint64) string {
	return strings.ToUpper(strconv.FormatUint(input, 36))
}

// GenerateMessageIDWithHostname generates an RFC 2822-compliant Message-Id
// based on the informational draft "Recommendations for generating Message
// IDs", it takes an hostname as argument, so that software using this library
// could use a hostname they know to be unique
func GenerateMessageIDWithHostname(hostname string) (string, error) {
	now := uint64(time.Now().UnixNano())

	nonceByte := make([]byte, 8)
	if _, err := rand.Read(nonceByte); err != nil {
		return "", err
	}
	nonce := binary.BigEndian.Uint64(nonceByte)

	return fmt.Sprintf("%s.%s@%s", base36(now), base36(nonce), hostname), nil
}
