package email

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	mail "gopkg.in/mail.v2"
)

func TestBuildMessage_NoAttachments(t *testing.T) {
	config := &Config{
		mailParameters: mailParameters{
			FromAddress: "foo@example.com",
			FromName:    "uTask Test",
			To:          []string{"\"Test To 1\" <to1@example.com>", "\"Test To 2\" <to2@example.com>"},
			Subject:     "Test Subject",
			Body:        "Hello, world",
		},
	}

	message := buildMessage(config)
	lines := messageLines(message)

	assert.Contains(t, lines, "From: \"uTask Test\" <foo@example.com>")
	assert.Contains(t, lines, "To: \"Test To 1\" <to1@example.com>, \"Test To 2\" <to2@example.com>")
	assert.Contains(t, lines, "Subject: Test Subject")
	assert.Contains(t, lines, "Hello, world")
}

func TestBuildMessage_WithAttachment(t *testing.T) {
	config := &Config{
		mailParameters: mailParameters{
			FromAddress: "foo@example.com",
			FromName:    "uTask Test",
			To:          []string{"\"Test To 1\" <to1@example.com>"},
			Subject:     "Test Subject",
			Body:        "Hello, world",
			Attachments: []string{
				"../../../../testdata/email-attachment.txt",
			},
		},
	}

	message := buildMessage(config)
	lines := messageLines(message)

	assert.Contains(t, lines, "From: \"uTask Test\" <foo@example.com>")
	assert.Contains(t, lines, "To: \"Test To 1\" <to1@example.com>")
	assert.Contains(t, lines, "Subject: Test Subject")
	assert.Contains(t, lines, "Hello, world")
	// Emails with attachments will be multipart/mixed
	assert.Contains(t, lines, "Content-Type: multipart/mixed;")
	// Did we get the attachment header and body?
	assert.Contains(t, lines, "Content-Type: text/plain; charset=utf-8; name=\"email-attachment.txt\"")
	assert.Contains(t, lines, "U2FtcGxlIGVtYWlsIGF0dGFjaG1lbnQg8J+OiQo=")
}

func messageLines(message *mail.Message) []string {
	var buffer bytes.Buffer
	message.WriteTo(&buffer)
	scanner := bufio.NewScanner(&buffer)
	lines := make([]string, 0)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}
