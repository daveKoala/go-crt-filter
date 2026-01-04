package status

import "testing"

func TestCheck(t *testing.T) {
	result := Check()
	if result != "Database: OK" {
		t.Errorf("Expected 'Database: OK', got '%s'", result)
	}
}
