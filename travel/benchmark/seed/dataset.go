package seed

import (
	"fmt"
	"time"
	"travel/TravelModel"

	uuid "github.com/satori/go.uuid"
)

type Scale struct {
	Users     int
	Posts     int
	Comments  int
	Likes     int
	Favorites int
	Foots     int
}

type Dataset struct {
	Users     []TravelModel.TraUser
	Posts     []TravelModel.Post
	Comments  []TravelModel.PostComment
	Likes     []TravelModel.TraUserPostLike
	Favorites []TravelModel.TraUserPostStart
	Foots     []TravelModel.TraFoot
}

func DefaultScale() Scale {
	return Scale{
		Users: 1000, Posts: 10000, Comments: 50000,
		Likes: 100000, Favorites: 30000, Foots: 10000,
	}
}

func Build(scale Scale) (Dataset, error) {
	if err := validateScale(scale); err != nil {
		return Dataset{}, err
	}

	dataset := Dataset{
		Users:     make([]TravelModel.TraUser, scale.Users),
		Posts:     make([]TravelModel.Post, scale.Posts),
		Comments:  make([]TravelModel.PostComment, scale.Comments),
		Likes:     make([]TravelModel.TraUserPostLike, scale.Likes),
		Favorites: make([]TravelModel.TraUserPostStart, scale.Favorites),
		Foots:     make([]TravelModel.TraFoot, scale.Foots),
	}
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)

	for i := range dataset.Users {
		id := uint64(i + 1)
		dataset.Users[i] = TravelModel.TraUser{
			ID: id, OpenID: fmt.Sprintf("benchmark-openid-%06d", id),
			NickName: fmt.Sprintf("benchmark-user-%04d", id), Motto: "benchmark user",
			City: "Shanghai", Province: "Shanghai", Country: "China",
			CreatedAt: TravelModel.CustomTime(baseTime.Add(time.Duration(i) * time.Second)),
			UpdatedAt: TravelModel.CustomTime(baseTime.Add(time.Duration(i) * time.Second)),
		}
	}
	for i := range dataset.Posts {
		created := baseTime.Add(time.Duration(i) * time.Minute)
		dataset.Posts[i] = TravelModel.Post{
			ID: deterministicUUID("post", i), UserID: uint64(i%scale.Users + 1),
			Title:     fmt.Sprintf("benchmark-post-%05d", i+1),
			HeadImg:   "https://example.invalid/benchmark.jpg",
			Content:   "Deterministic benchmark article content for isolated load testing.",
			ViewCount: uint64((i * 37) % 100000), LikeCount: uint64((i * 17) % 10000),
			CreatedAt: TravelModel.CustomTime(created), UpdatedAt: TravelModel.CustomTime(created),
		}
	}
	for i := range dataset.Comments {
		created := baseTime.Add(time.Duration(i) * time.Second)
		dataset.Comments[i] = TravelModel.PostComment{
			ID: deterministicUUID("comment", i), PostID: dataset.Posts[i%scale.Posts].ID,
			UserID: uint64((i*7)%scale.Users + 1), Content: fmt.Sprintf("benchmark-comment-%d", i+1),
			CreatedAt: TravelModel.CustomTime(created), UpdatedAt: TravelModel.CustomTime(created),
		}
	}
	for i := range dataset.Likes {
		userIndex, postIndex := relationIndexes(i, scale.Users, scale.Posts)
		dataset.Likes[i] = TravelModel.TraUserPostLike{
			ID: uint64(i + 1), UserID: uint64(userIndex + 1), PostID: dataset.Posts[postIndex].ID,
		}
	}
	for i := range dataset.Favorites {
		userIndex, postIndex := relationIndexes(i, scale.Users, scale.Posts)
		dataset.Favorites[i] = TravelModel.TraUserPostStart{
			ID: uint64(i + 1), UserID: uint64(userIndex + 1), PostID: dataset.Posts[postIndex].ID,
		}
	}
	for i := range dataset.Foots {
		created := baseTime.Add(time.Duration(i) * time.Minute)
		dataset.Foots[i] = TravelModel.TraFoot{
			ID: deterministicUUID("foot", i), UserID: uint64(i%scale.Users + 1),
			Title: fmt.Sprintf("benchmark-foot-%05d", i+1), Origin: "121.4737,31.2304",
			OriginName: "Shanghai", Destinations: "116.4074,39.9042",
			DestinationNames: "Beijing", Mode: "driving", RouteResult: `{"benchmark":true}`,
			CreatedAt: TravelModel.CustomTime(created), UpdatedAt: TravelModel.CustomTime(created),
		}
	}
	return dataset, nil
}

func validateScale(scale Scale) error {
	counts := []struct {
		name  string
		value int
	}{{"users", scale.Users}, {"posts", scale.Posts}, {"comments", scale.Comments}, {"likes", scale.Likes}, {"favorites", scale.Favorites}, {"foots", scale.Foots}}
	for _, count := range counts {
		if count.value < 0 {
			return fmt.Errorf("%s must not be negative", count.name)
		}
	}
	if (scale.Posts > 0 || scale.Comments > 0 || scale.Likes > 0 || scale.Favorites > 0 || scale.Foots > 0) && scale.Users == 0 {
		return fmt.Errorf("users must be positive when dependent records exist")
	}
	if (scale.Comments > 0 || scale.Likes > 0 || scale.Favorites > 0) && scale.Posts == 0 {
		return fmt.Errorf("posts must be positive when post relations exist")
	}
	capacity := int64(scale.Users) * int64(scale.Posts)
	if int64(scale.Likes) > capacity || int64(scale.Favorites) > capacity {
		return fmt.Errorf("relation count exceeds unique user/post capacity %d", capacity)
	}
	return nil
}

func relationIndexes(index, users, posts int) (int, int) {
	return index % users, (index / users) % posts
}

func relationKey(userID uint64, postID string) string {
	return fmt.Sprintf("%d/%s", userID, postID)
}

func deterministicUUID(kind string, index int) uuid.UUID {
	return uuid.NewV5(uuid.NamespaceOID, fmt.Sprintf("ongoing-trip-benchmark/%s/%d", kind, index))
}
