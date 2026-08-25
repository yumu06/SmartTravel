package seed

import (
	"context"
	"fmt"
	"strings"
	"travel/TravelModel"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const BenchmarkDatabase = "travel_benchmark"

func ValidateDatabaseName(name string) error {
	if name != BenchmarkDatabase {
		return fmt.Errorf("refusing database %q: only %q is allowed", name, BenchmarkDatabase)
	}
	return nil
}

func ResetAndSeed(ctx context.Context, adminDSN, databaseName string, dataset Dataset) error {
	if err := ValidateDatabaseName(databaseName); err != nil {
		return err
	}
	admin, err := gorm.Open(mysql.Open(adminDSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return fmt.Errorf("connect mysql admin database: %w", err)
	}
	admin = admin.WithContext(ctx)
	if err := admin.Exec("CREATE DATABASE IF NOT EXISTS `travel_benchmark` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci").Error; err != nil {
		return fmt.Errorf("create benchmark database: %w", err)
	}

	databaseDSN, err := withDatabase(adminDSN, databaseName)
	if err != nil {
		return err
	}
	db, err := gorm.Open(mysql.Open(databaseDSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return fmt.Errorf("connect benchmark database: %w", err)
	}
	benchmarkDB := db.WithContext(ctx)
	if err := benchmarkDB.Exec("SET FOREIGN_KEY_CHECKS=0").Error; err != nil {
		return fmt.Errorf("disable benchmark foreign key checks: %w", err)
	}
	models := []interface{}{
		&TravelModel.PostComment{}, &TravelModel.TraUserPostLike{}, &TravelModel.TraUserPostStart{},
		&TravelModel.TraUserFootStart{}, &TravelModel.TraUserFoot{}, &TravelModel.ChatMessage{},
		&TravelModel.TraFoot{}, &TravelModel.Post{}, &TravelModel.Notice{}, &TravelModel.TraUser{},
	}
	if err := benchmarkDB.Migrator().DropTable(models...); err != nil {
		_ = benchmarkDB.Exec("SET FOREIGN_KEY_CHECKS=1").Error
		return fmt.Errorf("drop benchmark tables: %w", err)
	}
	if err := benchmarkDB.Exec("SET FOREIGN_KEY_CHECKS=1").Error; err != nil {
		return fmt.Errorf("restore benchmark foreign key checks: %w", err)
	}
	if err := benchmarkDB.AutoMigrate(
		&TravelModel.TraUser{}, &TravelModel.TraUserFoot{}, &TravelModel.TraUserFootStart{},
		&TravelModel.TraUserPostStart{}, &TravelModel.TraUserPostLike{}, &TravelModel.TraFoot{},
		&TravelModel.Post{}, &TravelModel.PostComment{}, &TravelModel.Notice{}, &TravelModel.ChatMessage{},
	); err != nil {
		return fmt.Errorf("migrate benchmark database: %w", err)
	}

	batches := []struct {
		name string
		run  func() error
	}{
		{"users", func() error { return createBatch(db, &dataset.Users, false) }},
		{"posts", func() error { return createBatch(db, &dataset.Posts, true) }},
		{"comments", func() error { return createBatch(db, &dataset.Comments, true) }},
		{"likes", func() error { return createBatch(db, &dataset.Likes, false) }},
		{"favorites", func() error { return createBatch(db, &dataset.Favorites, false) }},
		{"foots", func() error { return createBatch(db, &dataset.Foots, true) }},
	}
	for _, batch := range batches {
		if err := batch.run(); err != nil {
			return fmt.Errorf("insert %s: %w", batch.name, err)
		}
	}
	counts := []struct {
		name  string
		model interface{}
		want  int
	}{
		{"users", &TravelModel.TraUser{}, len(dataset.Users)},
		{"posts", &TravelModel.Post{}, len(dataset.Posts)},
		{"comments", &TravelModel.PostComment{}, len(dataset.Comments)},
		{"likes", &TravelModel.TraUserPostLike{}, len(dataset.Likes)},
		{"favorites", &TravelModel.TraUserPostStart{}, len(dataset.Favorites)},
		{"foots", &TravelModel.TraFoot{}, len(dataset.Foots)},
	}
	for _, count := range counts {
		var got int64
		if err := benchmarkDB.Model(count.model).Count(&got).Error; err != nil {
			return fmt.Errorf("count %s: %w", count.name, err)
		}
		if got != int64(count.want) {
			return fmt.Errorf("verify %s count: got %d, want %d", count.name, got, count.want)
		}
	}
	return nil
}

func createBatch(db *gorm.DB, values interface{}, skipHooks bool) error {
	session := db.Session(&gorm.Session{SkipHooks: skipHooks})
	return session.CreateInBatches(values, 500).Error
}

func withDatabase(adminDSN, databaseName string) (string, error) {
	marker := "/?"
	if !strings.Contains(adminDSN, marker) {
		return "", fmt.Errorf("admin DSN must contain %q before query parameters", marker)
	}
	return strings.Replace(adminDSN, marker, "/"+databaseName+"?", 1), nil
}
