package prediction

import (
	"context"
	"testing"
	"time"

	"github.com/sidpromo/spotify-setlistfm/internal/setlist"
)

func TestService_Predict_Success(t *testing.T) {
	svc := NewService(DefaultConfig())

	input := PredictionInput{
		Artist:   "Metallica",
		MBID:     "abc",
		TourName: "M72",
		Setlists: []setlist.Setlist{
			{Songs: []setlist.Song{{Name: "Enter Sandman"}, {Name: "Fuel"}, {Name: "One"}, {Name: "NEM"}}, EventDate: time.Now()},
			{Songs: []setlist.Song{{Name: "Enter Sandman"}, {Name: "Fuel"}, {Name: "One"}, {Name: "NEM"}}, EventDate: time.Now().AddDate(0, 0, -7)},
			{Songs: []setlist.Song{{Name: "Enter Sandman"}, {Name: "Fuel"}, {Name: "One"}, {Name: "NEM"}}, EventDate: time.Now().AddDate(0, 0, -14)},
		},
	}

	result, err := svc.Predict(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.BasedOnCount != 3 {
		t.Errorf("expected basedOnCount 3, got %d", result.BasedOnCount)
	}
	if len(result.Songs) != 4 {
		t.Fatalf("expected 4 songs (median length), got %d", len(result.Songs))
	}
	if result.Songs[0].Name != "Enter Sandman" {
		t.Errorf("expected Enter Sandman as opener, got %q", result.Songs[0].Name)
	}
	if result.TourName != "M72" {
		t.Errorf("expected tour M72, got %q", result.TourName)
	}
}

func TestService_Predict_IdenticalSetlists(t *testing.T) {
	svc := NewService(DefaultConfig())

	songs := []setlist.Song{{Name: "A"}, {Name: "B"}, {Name: "C"}, {Name: "D"}, {Name: "E"}}
	input := PredictionInput{
		Artist: "Band",
		MBID:   "xyz",
		Setlists: []setlist.Setlist{
			{Songs: songs, EventDate: time.Now()},
			{Songs: songs, EventDate: time.Now().AddDate(0, 0, -3)},
			{Songs: songs, EventDate: time.Now().AddDate(0, 0, -6)},
		},
	}

	result, err := svc.Predict(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// All identical → prediction should match exactly
	if len(result.Songs) != 5 {
		t.Fatalf("expected 5, got %d", len(result.Songs))
	}
	for i, s := range result.Songs {
		expected := string(rune('A' + i))
		if s.Name != expected {
			t.Errorf("position %d: expected %q, got %q", i+1, expected, s.Name)
		}
	}
}

func TestService_Predict_SingleSetlist(t *testing.T) {
	svc := NewService(DefaultConfig())

	input := PredictionInput{
		Artist:   "Solo",
		MBID:     "solo",
		Setlists: []setlist.Setlist{
			{Songs: []setlist.Song{{Name: "X"}, {Name: "Y"}, {Name: "Z"}}, EventDate: time.Now()},
		},
	}

	result, err := svc.Predict(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Songs) != 3 {
		t.Fatalf("expected 3, got %d", len(result.Songs))
	}
}

func TestService_Predict_EmptyInput(t *testing.T) {
	svc := NewService(DefaultConfig())
	_, err := svc.Predict(context.Background(), PredictionInput{})
	if err != ErrNotEnoughData {
		t.Fatalf("expected ErrNotEnoughData, got %v", err)
	}
}

func TestService_Predict_VariableSetlists(t *testing.T) {
	svc := NewService(DefaultConfig())

	input := PredictionInput{
		Artist: "Varied",
		MBID:   "v",
		Setlists: []setlist.Setlist{
			{Songs: []setlist.Song{{Name: "Hit"}, {Name: "B"}, {Name: "C"}, {Name: "Closer"}}, EventDate: time.Now()},
			{Songs: []setlist.Song{{Name: "Hit"}, {Name: "D"}, {Name: "E"}, {Name: "Closer"}}, EventDate: time.Now().AddDate(0, 0, -3)},
			{Songs: []setlist.Song{{Name: "Hit"}, {Name: "F"}, {Name: "G"}, {Name: "Closer"}}, EventDate: time.Now().AddDate(0, 0, -6)},
		},
	}

	result, err := svc.Predict(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Hit appears in all (opener lock), Closer appears in all (closer lock)
	if result.Songs[0].Name != "Hit" {
		t.Errorf("expected Hit as opener, got %q", result.Songs[0].Name)
	}
	if result.Songs[len(result.Songs)-1].Name != "Closer" {
		t.Errorf("expected Closer as last, got %q", result.Songs[len(result.Songs)-1].Name)
	}
}
