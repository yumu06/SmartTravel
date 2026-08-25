package TravelModel

import (
	"reflect"
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestPostRecommendationIndexMatchesQueryOrder(t *testing.T) {
	parsed, err := schema.Parse(&Post{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse Post schema: %v", err)
	}
	index, exists := parsed.ParseIndexes()["idx_post_recommend"]
	if !exists {
		t.Fatal("idx_post_recommend is missing")
	}
	gotFields := make([]string, len(index.Fields))
	gotSort := make([]string, len(index.Fields))
	for i, field := range index.Fields {
		gotFields[i] = field.DBName
		gotSort[i] = field.Sort
	}
	if want := []string{"like_count", "view_count", "created_at"}; !reflect.DeepEqual(gotFields, want) {
		t.Fatalf("index fields = %v, want %v", gotFields, want)
	}
	if want := []string{"desc", "desc", "desc"}; !reflect.DeepEqual(gotSort, want) {
		t.Fatalf("index sort = %v, want %v", gotSort, want)
	}
}
