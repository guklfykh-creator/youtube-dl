package main

import (
	"os"
	"strings"
	"testing"
)

func TestMySQLURLToDSN(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "railway url with port",
			raw:  "mysql://root:oxoVtPfjMAWpybzeyUAgIHcKhbOMumab@monorail.proxy.rlwy.net:47715/railway",
			want: "root:oxoVtPfjMAWpybzeyUAgIHcKhbOMumab@tcp(monorail.proxy.rlwy.net:47715)/railway?charset=utf8mb4&parseTime=true",
		},
		{
			name: "url without port defaults 3306",
			raw:  "mysql://user:pass@localhost/mydb",
			want: "user:pass@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=true",
		},
		{
			name: "url with existing query params preserved",
			raw:  "mysql://user:pass@localhost:3306/db?tls=true",
			want: "user:pass@tcp(localhost:3306)/db?charset=utf8mb4&parseTime=true&tls=true",
		},
		{
			name: "url without password",
			raw:  "mysql://user@localhost:3306/mydb",
			want: "user@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mysqlURLToDSN(tt.raw)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMySQLURLToDSNInvalid(t *testing.T) {
	_, err := mysqlURLToDSN("not-a-url")
	if err == nil {
		t.Error("expected error for invalid input")
	}
}

func TestResolveMySQLDSNPriority(t *testing.T) {
	os.Setenv("MYSQL_URL", "mysql://root:pass@host:3306/db1")
	os.Setenv("MYSQL_DSN", "root:pass@tcp(host2:3306)/db2")
	defer os.Unsetenv("MYSQL_URL")
	defer os.Unsetenv("MYSQL_DSN")

	got := resolveMySQLDSN()
	if !contains(got, "host:3306") || !contains(got, "db1") {
		t.Errorf("MYSQL_URL should take priority, got %q", got)
	}
}

func TestResolveMySQLDSNFallback(t *testing.T) {
	os.Unsetenv("MYSQL_URL")
	os.Setenv("MYSQL_DSN", "root:pass@tcp(host2:3306)/db2")
	defer os.Unsetenv("MYSQL_DSN")

	got := resolveMySQLDSN()
	if got != "root:pass@tcp(host2:3306)/db2" {
		t.Errorf("should fall back to MYSQL_DSN, got %q", got)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && strings.Contains(s, sub)
}