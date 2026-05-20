package vpsadmin

import (
	"errors"
	"strings"
	"testing"

	"github.com/vpsfreecz/vpsadmin-go-client/client"
)

func TestWaitForOperation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		watcher   *fakeOperationWatcher
		wantErr   string
		wantWaits int
	}{
		{
			name:    "non-blocking",
			watcher: &fakeOperationWatcher{blocking: false},
		},
		{
			name: "finished success",
			watcher: &fakeOperationWatcher{
				blocking: true,
				responses: []*client.ActionActionStatePollResponse{
					pollResponse(true, true, true, ""),
				},
			},
			wantWaits: 1,
		},
		{
			name: "wait error",
			watcher: &fakeOperationWatcher{
				blocking: true,
				errs:     []error{errors.New("poll failed")},
			},
			wantErr:   "poll failed",
			wantWaits: 1,
		},
		{
			name: "failed envelope",
			watcher: &fakeOperationWatcher{
				blocking: true,
				responses: []*client.ActionActionStatePollResponse{
					pollResponse(false, false, false, "api failed"),
				},
			},
			wantErr:   "api failed",
			wantWaits: 1,
		},
		{
			name: "finished failure",
			watcher: &fakeOperationWatcher{
				blocking: true,
				responses: []*client.ActionActionStatePollResponse{
					pollResponse(true, true, false, ""),
				},
			},
			wantErr:   "Operation failed",
			wantWaits: 1,
		},
		{
			name: "timeout",
			watcher: &fakeOperationWatcher{
				blocking: true,
				responses: []*client.ActionActionStatePollResponse{
					pollResponse(true, false, false, ""),
				},
			},
			wantErr:   "Operation timed out",
			wantWaits: 60,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := waitForOperation(tt.watcher)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("waitForOperation() error = %v, want nil", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("waitForOperation() error = %v, want %q", err, tt.wantErr)
			}

			if tt.watcher.waits != tt.wantWaits {
				t.Fatalf("waits = %d, want %d", tt.watcher.waits, tt.wantWaits)
			}
		})
	}
}

type fakeOperationWatcher struct {
	blocking  bool
	responses []*client.ActionActionStatePollResponse
	errs      []error
	waits     int
}

func (w *fakeOperationWatcher) IsBlocking() bool {
	return w.blocking
}

func (w *fakeOperationWatcher) OperationStatus() (*client.ActionActionStateShowResponse, error) {
	return nil, nil
}

func (w *fakeOperationWatcher) WaitForOperation(timeout float64) (*client.ActionActionStatePollResponse, error) {
	w.waits++

	if len(w.errs) >= w.waits && w.errs[w.waits-1] != nil {
		return nil, w.errs[w.waits-1]
	}

	if len(w.responses) == 0 {
		return pollResponse(true, false, false, ""), nil
	}

	if len(w.responses) >= w.waits {
		return w.responses[w.waits-1], nil
	}

	return w.responses[len(w.responses)-1], nil
}

func (w *fakeOperationWatcher) WatchOperation(
	timeout float64,
	updateIn float64,
	callback client.OperationProgressCallback,
) (*client.ActionActionStatePollResponse, error) {
	return nil, nil
}

func pollResponse(status bool, finished bool, outputStatus bool, message string) *client.ActionActionStatePollResponse {
	return &client.ActionActionStatePollResponse{
		Envelope: &client.Envelope{
			Status:  status,
			Message: message,
		},
		Output: &client.ActionActionStatePollOutput{
			Finished: finished,
			Status:   outputStatus,
		},
	}
}
