package luhn

import "testing"

func TestCheck(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"valid", "5105105105105100", true},
		{"valid", "5100705011796135", true},
		{"valid", "2200150223544344", true},
		{"valid", "02200150223544344", true},
		{"valid", "18", true},
		{"valid", "018", true},
		{"valid", "0018", true},
		{"invalid", "0123456", false},
		{"invalid", "0", false},
		{"valid", "00", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Check(tt.number); got != tt.want {
				t.Errorf("Check() = %v, want %v", got, tt.want)
			}
		})
	}
}
