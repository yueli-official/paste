package site

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSettingsLifecycle(t *testing.T) {
	now := time.Date(2026, 8, 11, 8, 0, 0, 0, time.UTC)
	service, err := New(NewMemoryStore(), Options{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	current, err := service.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if current.Name != DefaultName || current.Revision != 1 {
		t.Fatalf("unexpected defaults: %#v", current)
	}
	updated, err := service.Update(context.Background(), UpdateInput{
		Name: " 月离代码 ", Description: " 分享调试现场 ", ExpectedRevision: 1, UpdatedBy: "usr_ADMIN",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "月离代码" || updated.Description != "分享调试现场" || updated.Revision != 2 || updated.UpdatedBy != "usr_ADMIN" {
		t.Fatalf("unexpected update: %#v", updated)
	}
	if _, err := service.Update(context.Background(), UpdateInput{Name: "再次修改", ExpectedRevision: 1}); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestSettingsValidation(t *testing.T) {
	service, err := New(NewMemoryStore(), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Update(context.Background(), UpdateInput{Name: " ", ExpectedRevision: 1}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected required name validation, got %v", err)
	}
	if _, err := service.Update(context.Background(), UpdateInput{Name: strings.Repeat("站", MaxNameRunes+1), ExpectedRevision: 1}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected name length validation, got %v", err)
	}
}
