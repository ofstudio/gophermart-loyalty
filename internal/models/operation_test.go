package models

import "testing"

func TestStatusCanTransit(t *testing.T) {
	tests := []struct {
		name string
		from OperationStatus
		to   OperationStatus
		want bool
	}{
		{
			name: "new to processed",
			from: StatusNew,
			to:   StatusProcessed,
			want: true,
		}, {
			name: "canceled to canceled",
			from: StatusCanceled,
			to:   StatusCanceled,
			want: true,
		},
		{
			name: "from non-existing to canceled",
			from: OperationStatus("NON-EXISTING-STATUS"),
			to:   StatusCanceled,
			want: false,
		},
		{
			name: "from non-existing to non-existing",
			from: "NON-EXISTING-STATUS-1",
			to:   "NON-EXISTING-STATUS-2",
			want: false,
		},
		{
			name: "from new to non-existing",
			from: StatusNew,
			to:   "NON-EXISTING-STATUS",
			want: false,
		},
		{
			name: "from processed to processing",
			from: StatusProcessed,
			to:   StatusProcessing,
			want: false,
		},
		{
			name: "from processing to canceled",
			from: StatusProcessing,
			to:   StatusCanceled,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.from.CanTransit(tt.to); got != tt.want {
				t.Errorf("CanTransit() = %v, want %v", got, tt.want)
			}
		})
	}
}
