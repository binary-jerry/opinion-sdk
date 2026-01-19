package common

import (
	"testing"
	"time"
)

func TestTimestampMs(t *testing.T) {
	ts := TimestampMs()
	now := time.Now().UnixMilli()
	if ts > now+1000 || ts < now-1000 {
		t.Errorf("TimestampMs() = %d, should be close to %d", ts, now)
	}
}

func TestTimestampSec(t *testing.T) {
	ts := TimestampSec()
	now := time.Now().Unix()
	if ts > now+1 || ts < now-1 {
		t.Errorf("TimestampSec() = %d, should be close to %d", ts, now)
	}
}

func TestTimestampSecStr(t *testing.T) {
	ts := TimestampSecStr()
	if ts == "" {
		t.Error("TimestampSecStr() should not return empty string")
	}
}

func TestTimestampMsStr(t *testing.T) {
	ts := TimestampMsStr()
	if ts == "" {
		t.Error("TimestampMsStr() should not return empty string")
	}
}

func TestGenerateRandomSalt(t *testing.T) {
	salt1, err := GenerateRandomSalt()
	if err != nil {
		t.Fatalf("GenerateRandomSalt() error: %v", err)
	}
	if salt1 == 0 {
		t.Fatal("GenerateRandomSalt() returned zero")
	}

	salt2, _ := GenerateRandomSalt()
	if salt1 == salt2 {
		t.Error("GenerateRandomSalt() should return different values")
	}
}

func TestGenerateRandomNonce(t *testing.T) {
	nonce1, err := GenerateRandomNonce()
	if err != nil {
		t.Fatalf("GenerateRandomNonce() error: %v", err)
	}
	if nonce1 < 0 {
		t.Error("GenerateRandomNonce() should return positive value")
	}

	nonce2, _ := GenerateRandomNonce()
	if nonce1 == nonce2 {
		t.Error("GenerateRandomNonce() should return different values")
	}
}

func TestGenerateRandomHex(t *testing.T) {
	hex1, err := GenerateRandomHex(16)
	if err != nil {
		t.Fatalf("GenerateRandomHex() error: %v", err)
	}
	if len(hex1) != 32 { // 16 bytes = 32 hex chars
		t.Errorf("GenerateRandomHex(16) length = %d, want 32", len(hex1))
	}

	hex2, _ := GenerateRandomHex(16)
	if hex1 == hex2 {
		t.Error("GenerateRandomHex() should return different values")
	}
}

func TestPointerHelpers(t *testing.T) {
	s := StringPtr("test")
	if *s != "test" {
		t.Errorf("StringPtr() = %s, want test", *s)
	}

	i := IntPtr(42)
	if *i != 42 {
		t.Errorf("IntPtr() = %d, want 42", *i)
	}

	i64 := Int64Ptr(123456789)
	if *i64 != 123456789 {
		t.Errorf("Int64Ptr() = %d, want 123456789", *i64)
	}

	b := BoolPtr(true)
	if *b != true {
		t.Errorf("BoolPtr() = %v, want true", *b)
	}
}

func TestParseBigInt(t *testing.T) {
	tests := []struct {
		input   string
		wantOk  bool
		wantStr string
	}{
		{"12345", true, "12345"},
		{"0", true, "0"},
		{"invalid", false, ""},
	}

	for _, tt := range tests {
		n, ok := ParseBigInt(tt.input)
		if ok != tt.wantOk {
			t.Errorf("ParseBigInt(%s) ok = %v, want %v", tt.input, ok, tt.wantOk)
		}
		if ok && n.String() != tt.wantStr {
			t.Errorf("ParseBigInt(%s) = %s, want %s", tt.input, n.String(), tt.wantStr)
		}
	}
}

func TestFormatBigInt(t *testing.T) {
	n, _ := ParseBigInt("12345")
	if FormatBigInt(n) != "12345" {
		t.Errorf("FormatBigInt(12345) = %s, want 12345", FormatBigInt(n))
	}

	if FormatBigInt(nil) != "0" {
		t.Errorf("FormatBigInt(nil) = %s, want 0", FormatBigInt(nil))
	}
}

func TestMinMax(t *testing.T) {
	if Min(1, 2) != 1 {
		t.Errorf("Min(1, 2) = %d, want 1", Min(1, 2))
	}
	if Min(2, 1) != 1 {
		t.Errorf("Min(2, 1) = %d, want 1", Min(2, 1))
	}

	if Max(1, 2) != 2 {
		t.Errorf("Max(1, 2) = %d, want 2", Max(1, 2))
	}
	if Max(2, 1) != 2 {
		t.Errorf("Max(2, 1) = %d, want 2", Max(2, 1))
	}
}
