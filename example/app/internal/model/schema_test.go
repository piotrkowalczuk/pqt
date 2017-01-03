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

	"github.com/lib/pq"
	_ "github.com/lib/pq"
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

type suite struct {
	db       *sql.DB
	news     *model.NewsRepositoryBase
	category *model.CategoryRepositoryBase
	comment  *model.CommentRepositoryBase
	pkg      *model.PackageRepositoryBase
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
			Table:   model.TableNews,
			Columns: model.TableNewsColumns,
			DB:      db,
			Debug:   testPostgresDebug,
		},
		comment: &model.CommentRepositoryBase{
			Table:   model.TableComment,
			Columns: model.TableCommentColumns,
			DB:      db,
			Debug:   testPostgresDebug,
		},
		category: &model.CategoryRepositoryBase{
			Table:   model.TableCategory,
			Columns: model.TableCategoryColumns,
			DB:      db,
			Debug:   testPostgresDebug,
		},
	}
}

func (s *suite) teardown(t testing.TB) {
	if _, err := s.db.Exec("DROP SCHEMA example CASCADE"); err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func populateNews(t testing.TB, r *model.NewsRepositoryBase, nb int) {
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

var testNewsInsertData = map[string]struct {
	entity model.NewsEntity
	query  string
}{
	"minimum": {
		entity: model.NewsEntity{
			Title:   "title - minimum",
			Content: "content - minimum",
		},
		query: "INSERT INTO example.news (content, continue, score, title) VALUES ($1, $2, $3, $4) RETURNING content, continue, created_at, id, lead, score, title, updated_at, views_distribution",
	},
	"full": {
		entity: model.NewsEntity{
			Title: "title - full",
			Lead: sql.NullString{
				Valid:  true,
				String: "lead - full",
			},
			ViewsDistribution: model.NullFloat64Array{
				Valid: true,
				Float64Array: []float64{1.2, 2.3, 3.4, 4.5},
			},
			Score:             10.11,
			Content:           "content - full",
			Continue:          true,
			CreatedAt:         time.Now(),
			UpdatedAt: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
		},
		query: "INSERT INTO example.news (content, continue, created_at, lead, score, title, updated_at, views_distribution) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING content, continue, created_at, id, lead, score, title, updated_at, views_distribution",
	},
}

var (
	benchQuery string
	benchArgs  []interface{}
)

func BenchmarkNewsRepositoryBase_InsertQuery(b *testing.B) {
	s := setup(b)
	defer s.teardown(b)
	b.ResetTimer()

	for hint, given := range testNewsInsertData {
		b.Run(hint, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				query, args, err := s.news.InsertQuery(&given.entity)
				if err != nil {
					b.Fatalf("unexpected error: %s", err.Error())
				}
				benchQuery = query
				benchArgs = args
			}
		})
	}
}

func TestNewsRepositoryBase_InsertQuery(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	for hint, given := range testNewsInsertData {
		t.Run(hint, func(t *testing.T) {
			query, _, err := s.news.InsertQuery(&given.entity)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if given.query != query {
				t.Errorf("wrong output, expected:\n	%s\nbut got:\n	%s", given.query, query)
			}
		})
	}
}

func TestNewsRepositoryBase_Insert(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	for hint, given := range testNewsInsertData {
		t.Run(hint, func(t *testing.T) {
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			got, err := s.news.Insert(ctx, &given.entity)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if given.entity.Title != got.Title {
				t.Errorf("wrong title, expected %s but got %s", given.entity.Title, got.Title)
			}
			if given.entity.Lead != got.Lead {
				t.Errorf("wrong lead, expected %s but got %s", given.entity.Lead, got.Lead)
			}
			if given.entity.Content != got.Content {
				t.Errorf("wrong content, expected %s but got %s", given.entity.Content, got.Content)
			}
			if !given.entity.UpdatedAt.Valid && got.UpdatedAt.Valid {
				t.Error("updated at expected to be invalid")
			}
			if got.CreatedAt.IsZero() {
				t.Error("created at should not be zero value")
			}
			if given.entity.ViewsDistribution.Valid {
				if !got.ViewsDistribution.Valid {
					t.Error("views distribution should be valid")
				} else {

				}
			}
		})
	}
}

var testNewsFindData = map[string]struct {
	criteria model.NewsCriteria
	query    string
}{
	"minimum": {
		criteria: model.NewsCriteria{
			Title:   sql.NullString{String: "title - minimum", Valid: true},
			Content: sql.NullString{String: "content - minimum", Valid: true},
		},
		query: "SELECT content, continue, created_at, id, lead, score, title, updated_at, views_distribution FROM example.news WHERE content=$1 AND title=$2",
	},
	"full": {
		criteria: model.NewsCriteria{
			ID: sql.NullInt64{
				Int64: 1,
				Valid: true,
			},
			Score: sql.NullFloat64{
				Valid:   true,
				Float64: 10.11,
			},
			Title: sql.NullString{String: "title - full", Valid: true},
			Lead: sql.NullString{
				Valid:  true,
				String: "lead - full",
			},
			Content:  sql.NullString{String: "content - full", Valid: true},
			ViewsDistribution: model.NullFloat64Array{
				Valid: true,
				Float64Array: []float64{1.1, 1.2, 1.3},
			},
			Continue: sql.NullBool{Bool: true, Valid: true},
			CreatedAt: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
			UpdatedAt: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
		},
		query: "SELECT content, continue, created_at, id, lead, score, title, updated_at, views_distribution FROM example.news WHERE content=$1 AND continue=$2 AND created_at=$3 AND id=$4 AND lead=$5 AND score=$6 AND title=$7 AND updated_at=$8 AND views_distribution=$9",
	},
}

func BenchmarkNewsRepositoryBase_FindQuery(b *testing.B) {
	s := setup(b)
	defer s.teardown(b)
	b.ResetTimer()

	for hint, given := range testNewsFindData {
		b.Run(hint, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				query, args, err := s.news.FindQuery(s.news.Columns, &given.criteria)
				if err != nil {
					b.Fatalf("unexpected error: %s", err.Error())
				}
				benchQuery = query
				benchArgs = args
			}
		})
	}
}

func TestNewsRepositoryBase_FindQuery(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	for hint, given := range testNewsFindData {
		t.Run(hint, func(t *testing.T) {
			query, _, err := s.news.FindQuery(s.news.Columns, &given.criteria)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if given.query != query {
				t.Errorf("wrong output, expected:\n	%s\nbut got:\n	%s", given.query, query)
			}
		})
	}
}

func TestNewsRepositoryBase_Find(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	expected := 10
	populateNews(t, s.news, expected)
	got, err := s.news.Find(context.Background(), &model.NewsCriteria{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if len(got) != expected {
		t.Errorf("wrong output, expected %d but got %d", expected, got)
	}
}

func TestNewsRepositoryBase_FindIter(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	expected := 10
	populateNews(t, s.news, expected)
	iter, err := s.news.FindIter(context.Background(), &model.NewsCriteria{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	defer iter.Close()

	var got []*model.NewsEntity
	for iter.Next() {
		ent, err := iter.News()
		if err != nil {
			t.Fatalf("unexpected error: %s", err.Error())
		}
		got = append(got, ent)
	}
	if err = iter.Err(); err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	if len(got) != expected {
		t.Errorf("wrong output, expected %d but got %d", expected, got)
	}
}

func TestNewsRepositoryBase_FindOneByID(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	expected := 10
	populateNews(t, s.news, expected)

	for i := 1; i <= expected; i++ {
		got, err := s.news.FindOneByID(context.Background(), int64(i))
		if err != nil {
			t.Fatalf("unexpected error: %s", err.Error())
		}
		if got.ID != int64(i) {
			t.Errorf("wrong id, expected %d but got %d", i, got.ID)
		}
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

var testNewsUpdateData = map[string]struct {
	patch model.NewsPatch
	query string
}{
	"minimum": {
		patch: model.NewsPatch{
			Title:   sql.NullString{String: "title - minimum", Valid: true},
			Content: sql.NullString{String: "content - minimum", Valid: true},
		},
		query: "UPDATE example.news SET content=$1, title=$2, updated_at=NOW() WHERE id=$3 RETURNING content, continue, created_at, id, lead, score, title, updated_at, views_distribution",
	},
	"full": {
		patch: model.NewsPatch{
			Title: sql.NullString{String: "title - full", Valid: true},
			Lead: sql.NullString{
				Valid:  true,
				String: "lead - full",
			},
			Score: sql.NullFloat64{
				Valid:   true,
				Float64: 12.14,
			},
			Content:  sql.NullString{String: "content - full", Valid: true},
			Continue: sql.NullBool{Bool: true, Valid: true},
			CreatedAt: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
			UpdatedAt: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
		},
		query: "UPDATE example.news SET content=$1, continue=$2, created_at=$3, lead=$4, score=$5, title=$6, updated_at=$7 WHERE id=$8 RETURNING content, continue, created_at, id, lead, score, title, updated_at, views_distribution",
	},
}

func BenchmarkNewsRepositoryBase_UpdateOneByIDQuery(b *testing.B) {
	s := setup(b)
	defer s.teardown(b)
	b.ResetTimer()

	for hint, given := range testNewsUpdateData {
		b.Run(hint, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				query, args, err := s.news.UpdateOneByIDQuery(1, &given.patch)
				if err != nil {
					b.Fatalf("unexpected error: %s", err.Error())
				}
				benchQuery = query
				benchArgs = args
			}
		})
	}
}

func TestNewsRepositoryBase_UpdateOneByIDQuery(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	for hint, given := range testNewsUpdateData {
		t.Run(hint, func(t *testing.T) {
			query, _, err := s.news.UpdateOneByIDQuery(1, &given.patch)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if given.query != query {
				t.Errorf("wrong output, expected:\n	%s\nbut got:\n	%s", given.query, query)
			}
		})
	}
}
