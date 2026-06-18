package service

import (
	"testing"

	"taskflow-pro/backend/internal/model"
)

func TestPasswordHashCanBeVerified(t *testing.T) {
	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if hash == "secret123" {
		t.Fatal("password hash must not equal plaintext")
	}
	if !CheckPassword("secret123", hash) {
		t.Fatal("expected correct password to verify")
	}
	if CheckPassword("wrong", hash) {
		t.Fatal("expected wrong password to fail")
	}
}

func TestTaskStatusValidation(t *testing.T) {
	valid := []model.TaskStatus{model.TaskTodo, model.TaskDoing, model.TaskDone}
	for _, status := range valid {
		if !model.IsValidTaskStatus(status) {
			t.Fatalf("expected %s to be valid", status)
		}
	}
	if model.IsValidTaskStatus("blocked") {
		t.Fatal("expected blocked to be invalid")
	}
}
