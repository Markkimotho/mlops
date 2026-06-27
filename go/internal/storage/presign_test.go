package storage

import (
	"strings"
	"testing"
	"time"
)

func TestPresignCreatesBoundedSigV4URL(t *testing.T) {
	got, err := Presign(Config{Endpoint: "http://minio:9000", Region: "us-east-1", AccessKey: "key", SecretKey: "secret"}, Request{Bucket: "models", Key: "churn/1/model.pkl", Operation: "GET", TTLSeconds: 300}, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"models/churn/1/model.pkl", "X-Amz-Algorithm=AWS4-HMAC-SHA256", "X-Amz-Signature="} {
		if !strings.Contains(got, expected) {
			t.Fatalf("URL missing %q: %s", expected, got)
		}
	}
}

func TestPresignRejectsExcessiveTTL(t *testing.T) {
	_, err := Presign(Config{Endpoint: "http://minio", AccessKey: "key", SecretKey: "secret"}, Request{Bucket: "x", Key: "x", Operation: "PUT", TTLSeconds: 901}, time.Now())
	if err == nil {
		t.Fatal("expected TTL error")
	}
}
