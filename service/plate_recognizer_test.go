package service

import (
	"errors"
	"testing"
)

func TestNormalizeRecognizeResult_NoPlateDetected(t *testing.T) {
	plate, score, err := normalizeRecognizeResult("", 0, ErrNoPlateDetected)
	if !errors.Is(err, ErrNoPlateDetected) {
		t.Fatalf("expected no plate detection error, got: %v", err)
	}
	if plate != "" {
		t.Fatalf("expected empty plate, got %q", plate)
	}
	if score != 0 {
		t.Fatalf("expected zero score, got %f", score)
	}
}

func TestNormalizeRecognizeResult_OtherErrorStillReturnsError(t *testing.T) {
	unexpected := errors.New("network failure")
	plate, score, err := normalizeRecognizeResult("", 0, unexpected)
	if !errors.Is(err, unexpected) {
		t.Fatalf("expected original error to be returned, got: %v", err)
	}
	if plate != "" {
		t.Fatalf("expected empty plate, got %q", plate)
	}
	if score != 0 {
		t.Fatalf("expected zero score, got %f", score)
	}
}
