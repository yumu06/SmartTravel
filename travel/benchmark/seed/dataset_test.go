package seed

import "testing"

func TestBuildProducesExactCountsAndUniqueRelations(t *testing.T) {
	scale := Scale{
		Users: 7, Posts: 11, Comments: 23,
		Likes: 31, Favorites: 29, Foots: 13,
	}

	dataset, err := Build(scale)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	checks := []struct {
		name string
		got  int
		want int
	}{
		{"users", len(dataset.Users), scale.Users},
		{"posts", len(dataset.Posts), scale.Posts},
		{"comments", len(dataset.Comments), scale.Comments},
		{"likes", len(dataset.Likes), scale.Likes},
		{"favorites", len(dataset.Favorites), scale.Favorites},
		{"foots", len(dataset.Foots), scale.Foots},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s count = %d, want %d", check.name, check.got, check.want)
		}
	}

	likePairs := make(map[string]struct{}, len(dataset.Likes))
	for _, like := range dataset.Likes {
		key := relationKey(like.UserID, like.PostID.String())
		if _, exists := likePairs[key]; exists {
			t.Fatalf("duplicate like pair %s", key)
		}
		likePairs[key] = struct{}{}
	}
	favoritePairs := make(map[string]struct{}, len(dataset.Favorites))
	for _, favorite := range dataset.Favorites {
		key := relationKey(favorite.UserID, favorite.PostID.String())
		if _, exists := favoritePairs[key]; exists {
			t.Fatalf("duplicate favorite pair %s", key)
		}
		favoritePairs[key] = struct{}{}
	}
}

func TestBuildRejectsImpossibleRelationCounts(t *testing.T) {
	_, err := Build(Scale{Users: 2, Posts: 3, Likes: 7})
	if err == nil {
		t.Fatal("Build() error = nil, want impossible relation count error")
	}
}

func TestValidateDatabaseNameOnlyAllowsBenchmarkDatabase(t *testing.T) {
	if err := ValidateDatabaseName("travel_benchmark"); err != nil {
		t.Fatalf("travel_benchmark rejected: %v", err)
	}
	for _, name := range []string{"", "travel_database", "TRAVEL_BENCHMARK", "travel_benchmark; DROP DATABASE travel_database"} {
		if err := ValidateDatabaseName(name); err == nil {
			t.Errorf("ValidateDatabaseName(%q) = nil, want error", name)
		}
	}
}
