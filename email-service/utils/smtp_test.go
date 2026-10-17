package utils

import (
	"bufio"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func TestProductionRejectsCaptureMode(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("SMTP_MODE", "capture")
	if SendOTPEmail("person@example.test", "Demo", "123456") == nil {
		t.Fatal("production accepted test email capture")
	}
}

func TestCaptureDeliversOTPToLocalSMTP(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	host, port, _ := net.SplitHostPort(listener.Addr().String())
	t.Setenv("APP_ENV", "test")
	t.Setenv("SMTP_MODE", "capture")
	t.Setenv("SMTP_HOST", host)
	t.Setenv("SMTP_PORT", port)
	t.Setenv("SMTP_FROM", "taaza@example.test")
	message := make(chan string, 1)
	go func() {
		connection, err := listener.Accept()
		if err != nil {
			return
		}
		defer connection.Close()
		connection.SetDeadline(time.Now().Add(5 * time.Second))
		reader := bufio.NewReader(connection)
		writer := bufio.NewWriter(connection)
		reply := func(s string) { writer.WriteString(s + "\r\n"); writer.Flush() }
		reply("220 local test SMTP")
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			switch {
			case strings.HasPrefix(line, "EHLO"), strings.HasPrefix(line, "HELO"):
				reply("250 local")
			case strings.HasPrefix(line, "MAIL FROM"), strings.HasPrefix(line, "RCPT TO"):
				reply("250 OK")
			case strings.HasPrefix(line, "DATA"):
				reply("354 continue")
				var body strings.Builder
				for {
					part, err := reader.ReadString('\n')
					if err != nil {
						return
					}
					if part == ".\r\n" {
						break
					}
					body.WriteString(part)
				}
				message <- body.String()
				reply("250 accepted")
			case strings.HasPrefix(line, "QUIT"):
				reply("221 bye")
				return
			default:
				reply("500 unsupported command")
			}
		}
	}()
	// Tests execute in utils/, so point the template loader at the service directory.
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(".."); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Error(err)
		}
	})
	if err := SendOTPEmail("person@example.test", "Demo Customer", "123456"); err != nil {
		t.Fatal(err)
	}
	select {
	case body := <-message:
		if !strings.Contains(body, "123456") || !strings.Contains(body, "person@example.test") {
			t.Fatal("captured email missing OTP or recipient")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no email received")
	}
}
