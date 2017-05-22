package model_test

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/piotrkowalczuk/pqt/example/app/internal/model"
)

var (
	testPostgresAddress string
	testPostgresDebug   bool
)

func TestMain(m *testing.M) {
	flag.BoolVar(&testPostgresDebug, "postgres.debug", getBoolEnvOr("PQT_POSTGRES_DEBUG", false), "if true, all queries will be logged")
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

func join(arr []string, id int) string {
	arr2 := make([]string, 0, len(arr))

	for _, a := range arr {
		arr2 = append(arr2, fmt.Sprintf("t%d.%s", id, a))
	}
	return strings.Join(arr2, ", ")
}

type suite struct {
	db       *sql.DB
	news     *model.NewsRepositoryBase
	category *model.CategoryRepositoryBase
	comment  *model.CommentRepositoryBase
	pkg      *model.PackageRepositoryBase
	complete *model.CompleteRepositoryBase
}

func setup(t testing.TB) *suite {
	db, err := sql.Open("postgres", testPostgresAddress)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	if _, err = db.Exec(model.SQL); err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	return &suite{
		db: db,
		news: &model.NewsRepositoryBase{
			Table: model.TableNews,
			DB:    db,
		},
		comment: &model.CommentRepositoryBase{
			Table: model.TableComment,
			DB:    db,
		},
		category: &model.CategoryRepositoryBase{
			Table: model.TableCategory,
			DB:    db,
		},
		complete: &model.CompleteRepositoryBase{
			Table: model.TableComplete,
			DB:    db,
		},
	}
}

func (s *suite) teardown(t testing.TB) {
	if _, err := s.db.Exec("DROP SCHEMA example CASCADE"); err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func populateNews(t testing.TB, r *model.NewsRepositoryBase, nb int) {
	for i := 1; i <= nb; i++ {
		_, err := r.Insert(context.Background(), &model.NewsEntity{
			Title:    fmt.Sprintf("title-%d", i),
			Content:  fmt.Sprintf("content-%d", i),
			Lead:     sql.NullString{String: fmt.Sprintf("lead-%d", i), Valid: true},
			Continue: true,
			Score:    10.11,
		})
		if err != nil {
			t.Fatalf("unexpected error #%d: %s", i, err.Error())
		}
	}
}

func populateCategory(t testing.TB, r *model.CategoryRepositoryBase, nb int) {
	for i := 1; i <= nb; i++ {
		_, err := r.Insert(context.Background(), &model.CategoryEntity{
			Name:      fmt.Sprintf("name-%d", i),
			Content:   fmt.Sprintf("content-%d", i),
			CreatedAt: time.Now(),
		})
		if err != nil {
			t.Fatalf("unexpected error #%d: %s", i, err.Error())
		}

		for j := 1; j <= nb; j++ {
			_, err := r.Insert(context.Background(), &model.CategoryEntity{
				ParentID:  sql.NullInt64{Int64: int64(i), Valid: true},
				Name:      fmt.Sprintf("name-%d-%d", i, j),
				Content:   fmt.Sprintf("content-%d-%d", i, j),
				CreatedAt: time.Now(),
				UpdatedAt: pq.NullTime{Time: time.Now(), Valid: true},
			})
			if err != nil {
				t.Fatalf("unexpected error #%d: %s", i, err.Error())
			}
		}
	}
}
