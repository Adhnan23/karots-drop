package store

import (
	"fmt"
	"sync"
	"testing"
)

func TestSetAndGet(t *testing.T) {
	s := New()
	defer s.Stop()

	if err := s.Set("123456", []byte("hello"), "", "text/plain", false); err != nil {
		t.Fatal(err)
	}
	item, ok := s.Get("123456")
	if !ok {
		t.Fatal("expected item")
	}
	if string(item.Data) != "hello" {
		t.Fatalf("expected 'hello', got %q", item.Data)
	}
	if item.ContentType != "text/plain" {
		t.Fatalf("expected 'text/plain', got %q", item.ContentType)
	}
	if item.Encrypted {
		t.Fatal("expected not encrypted")
	}
	if item.Filename != "" {
		t.Fatalf("expected empty filename, got %q", item.Filename)
	}
}

func TestGetNonExistent(t *testing.T) {
	s := New()
	defer s.Stop()

	_, ok := s.Get("000000")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestExists(t *testing.T) {
	s := New()
	defer s.Stop()

	if err := s.Set("123456", []byte("a"), "", "", false); err != nil {
		t.Fatal(err)
	}
	if !s.Exists("123456") {
		t.Fatal("expected exists")
	}
	if s.Exists("000000") {
		t.Fatal("expected not exist")
	}
}

func TestDelete(t *testing.T) {
	s := New()
	defer s.Stop()

	if err := s.Set("123456", []byte("a"), "", "", false); err != nil {
		t.Fatal(err)
	}
	s.Delete("123456")
	_, ok := s.Get("123456")
	if ok {
		t.Fatal("expected deleted")
	}
}

func TestLen(t *testing.T) {
	s := New()
	defer s.Stop()

	if s.Len() != 0 {
		t.Fatalf("expected 0, got %d", s.Len())
	}
	if err := s.Set("111111", []byte("a"), "", "", false); err != nil {
		t.Fatal(err)
	}
	if err := s.Set("222222", []byte("b"), "", "", false); err != nil {
		t.Fatal(err)
	}
	if s.Len() != 2 {
		t.Fatalf("expected 2, got %d", s.Len())
	}
	s.Delete("111111")
	if s.Len() != 1 {
		t.Fatalf("expected 1, got %d", s.Len())
	}
}

func TestGenerateCode(t *testing.T) {
	code, err := GenerateCode()
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != 6 {
		t.Fatalf("expected 6 digits, got %d", len(code))
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			t.Fatalf("non-digit char %c in code", c)
		}
	}
}

func TestGenerateUniqueCode(t *testing.T) {
	s := New()
	defer s.Stop()

	code, err := s.GenerateUniqueCode()
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != 6 {
		t.Fatalf("expected 6 digits, got %d", len(code))
	}
}

func TestGenerateUniqueCodeNoCollision(t *testing.T) {
	s := New()
	defer s.Stop()

	// Pre-fill first 100 codes, then generate should still work
	// (code gen retries up to 100 times, different codes each time)
	for i := 0; i < 1000; i++ {
		code, err := s.GenerateUniqueCode()
		if err != nil {
			t.Fatal(err)
		}
		if err := s.Set(code, []byte("x"), "", "", false); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMaxItems(t *testing.T) {
	s := NewWithMax(2)
	defer s.Stop()

	if err := s.Set("111111", []byte("a"), "", "", false); err != nil {
		t.Fatal(err)
	}
	if err := s.Set("222222", []byte("b"), "", "", false); err != nil {
		t.Fatal(err)
	}
	if err := s.Set("333333", []byte("c"), "", "", false); err != ErrStoreFull {
		t.Fatalf("expected ErrStoreFull, got %v", err)
	}
}

func TestMaxItemsZeroIsUnlimited(t *testing.T) {
	s := New()
	defer s.Stop()

	for i := 0; i < 1000; i++ {
		code := fmt.Sprintf("%06d", i)
		if err := s.Set(code, []byte("a"), "", "", false); err != nil {
			t.Fatalf("unexpected error at %d: %v", i, err)
		}
	}
	if s.Len() != 1000 {
		t.Fatalf("expected 1000 items, got %d", s.Len())
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := New()
	defer s.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			code, err := s.GenerateUniqueCode()
			if err != nil {
				t.Error(err)
				return
			}
			if err := s.Set(code, []byte{byte(i)}, "", "", false); err != nil {
				t.Error(err)
				return
			}
			s.Get(code)
			s.Exists(code)
		}(i)
	}
	wg.Wait()
}
