package model_test

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/piotrkowalczuk/pqt/example/app/internal/model"
	"github.com/piotrkowalczuk/sklog"
)

var (
	testPostgresAddress string
	testPostgresDebug   bool
)

func TestMain(m *testing.M) {
	flag.BoolVar(&testPostgresDebug, "postgres.debug", getBoolEnvOr("PQT_POSTGRES_DEBUG", true), "if true, all queries will be logged")
	flag.StringVar(&testPostgresAddress, "postgres.address", getStringEnvOr("PQT_POSTGRES_ADDRESS", "postgres://postgres:@localhost/test?sslmode=disable"), "postgres database connection address")
	flag.Parse()

	os.Exit(m.Run())
}

func getStringEnvOr(env, or string) string {
	if v := os.Getenv(env); v != "" {
		return v
	}
	return or
}

func getBoolEnvOr(env string, or bool) bool {
	if v := os.Getenv(env); v != "" {
		f, err := strconv.ParseBool(v)
		if err != nil {
			return or
		}
		return f
	}
	return or
}

type suite struct {
	db       *sql.DB
	news     *model.NewsRepositoryBase
	category *model.CategoryRepositoryBase
	comment  *model.CommentRepositoryBase
	pkg      *model.PackageRepositoryBase
}

func setup(t *testing.T) *suite {
	db, err := sql.Open("postgres", testPostgresAddress)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	if _, err = db.Exec(model.SQL); err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	log := sklog.NewTestLogger(t)
	return &suite{
		db: db,
		news: &model.NewsRepositoryBase{
			Table:   model.TableNews,
			Columns: model.TableNewsColumns,
			DB:      db,
			Debug:   testPostgresDebug,
			Log:     log,
		},
		comment: &model.CommentRepositoryBase{
			Table:   model.TableComment,
			Columns: model.TableCommentColumns,
			DB:      db,
			Debug:   testPostgresDebug,
			Log:     log,
		},
		category: &model.CategoryRepositoryBase{
			Table:   model.TableCategory,
			Columns: model.TableCategoryColumns,
			DB:      db,
			Debug:   testPostgresDebug,
			Log:     log,
		},
	}
}

func (s *suite) teardown(t *testing.T) {
	if _, err := s.db.Exec("DROP SCHEMA example CASCADE"); err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func populateNews(t *testing.T, r *model.NewsRepositoryBase, nb int) {
	for i := 0; i < nb; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := r.Insert(ctx, &model.NewsEntity{
			Title:    fmt.Sprintf("test news %d - %d", time.Now().Unix(), rand.Int63()),
			Content:  fmt.Sprintf("content %d - %d", time.Now().Unix(), rand.Int63()),
			Continue: true,
		})
		if err != nil {
			t.Fatalf("unexpected error #%d: %s", i, err.Error())
		}
		cancel()
	}
}

func TestNewsRepositoryBase_Count(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	expected := 10
	populateNews(t, s.news, expected)
	got, err := s.news.Count(context.Background(), &model.NewsCriteria{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if got != int64(expected) {
		t.Errorf("wrong output, expected %d but got %d", expected, got)
	}
}
