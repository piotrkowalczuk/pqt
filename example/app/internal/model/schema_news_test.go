package model_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/piotrkowalczuk/pqt/example/app/internal/model"
)

var testNewsInsertData = map[string]struct {
	entity model.NewsEntity
	query  string
}{
	"minimum": {
		entity: model.NewsEntity{
			Title:   "title - minimum",
			Content: "content - minimum",
		},
		query: "INSERT INTO example.news (content, continue, score, title, version) VALUES ($1, $2, $3, $4, $5) RETURNING " + strings.Join(model.TableNewsColumns, ", "),
	},
	"full": {
		entity: model.NewsEntity{
			Title: "title - full",
			Lead: sql.NullString{
				Valid:  true,
				String: "lead - full",
			},
			ViewsDistribution: model.NullFloat64Array{
				Valid:        true,
				Float64Array: []float64{1.2, 2.3, 3.4, 4.5},
			},
			MetaData:  []byte(`{"something": 1}`),
			Score:     10.11,
			Content:   "content - full",
			Continue:  true,
			CreatedAt: time.Now(),
			UpdatedAt: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
		},
		query: "INSERT INTO example.news (content, continue, created_at, lead, meta_data, score, title, updated_at, version, views_distribution) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING " + strings.Join(model.TableNewsColumns, ", "),
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
				query, args, err := s.news.InsertQuery(&given.entity, true)
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
			query, _, err := s.news.InsertQuery(&given.entity, true)
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
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			got, err := s.news.Insert(ctx, &given.entity)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if given.entity.Title != got.Title {
				t.Errorf("wrong title, expected %s but got %s", given.entity.Title, got.Title)
			}
			if given.entity.Lead != got.Lead {
				t.Errorf("wrong lead, expected %v but got %v", given.entity.Lead, got.Lead)
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
				}
			}
		})
	}
}

var testNewsFindData = map[string]struct {
	expr  model.NewsFindExpr
	query string
}{
	"minimum": {
		expr: model.NewsFindExpr{
			Where: &model.NewsCriteria{
				Title:   sql.NullString{String: "title - minimum", Valid: true},
				Content: sql.NullString{String: "content - minimum", Valid: true},
			},
		},
		query: "SELECT " + join(model.TableNewsColumns, 0) + " FROM example.news AS t0 WHERE t0.content=$1 AND t0.title=$2",
	},
	"logical-condition": {
		expr: model.NewsFindExpr{
			Where: model.NewsOr(
				model.NewsAnd(
					&model.NewsCriteria{
						Title: sql.NullString{String: "title - minimum", Valid: true},
					},
					&model.NewsCriteria{
						Content: sql.NullString{String: "content - minimum", Valid: true},
					},
				),
				model.NewsAnd(
					&model.NewsCriteria{
						Content: sql.NullString{String: "content - maximum", Valid: true},
					},
					&model.NewsCriteria{
						Title: sql.NullString{String: "title - maximum", Valid: true},
					},
					&model.NewsCriteria{
						Version: sql.NullInt64{Int64: 1000, Valid: true},
						Lead:    sql.NullString{String: "lead - maximum", Valid: true},
					},
				),
			),
		},
		query: "SELECT " + join(model.TableNewsColumns, 0) + " FROM example.news AS t0 WHERE ((t0.title=$1) AND (t0.content=$2)) OR ((t0.content=$3) AND (t0.title=$4) AND (t0.lead=$5 AND t0.version=$6))",
	},
	"full": {
		expr: model.NewsFindExpr{
			Where: &model.NewsCriteria{
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
				MetaData: []byte(`{"something": 1}`),
				Content:  sql.NullString{String: "content - full", Valid: true},
				ViewsDistribution: model.NullFloat64Array{
					Valid:        true,
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
			Limit:  10,
			Offset: 100,
			OrderBy: []model.RowOrder{
				{
					Name:       model.TableNewsColumnTitle,
					Descending: true,
				},
				{
					Name: model.TableNewsColumnLead,
				},
			},
		},
		query: "SELECT " + join(model.TableNewsColumns, 0) + " FROM example.news AS t0 WHERE t0.content=$1 AND t0.continue=$2 AND t0.created_at=$3 AND t0.lead=$4 AND t0.meta_data=$5 AND t0.score=$6 AND t0.title=$7 AND t0.updated_at=$8 AND t0.views_distribution=$9 ORDER BY title DESC, lead OFFSET $10  LIMIT $11 ",
	},
}

func BenchmarkNewsRepositoryBase_FindQuery(b *testing.B) {
	s := setup(b)
	defer s.teardown(b)
	b.ResetTimer()

	for hint, given := range testNewsFindData {
		b.Run(hint, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				query, args, err := s.news.FindQuery(&given.expr)
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
			query, _, err := s.news.FindQuery(&given.expr)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if given.query != query {
				t.Errorf("wrong output, expected:\n	%s\nbut got:\n	%s", given.query, query)
			}
		})
	}
}

func TestNewsRepositoryBase_DeleteOneByID(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	nb := 10
	populateNews(t, s.news, nb)
	for i := 1; i <= nb; i++ {
		got, err := s.news.DeleteOneByID(context.Background(), int64(i))
		if err != nil {
			t.Fatalf("unexpected error: %s", err.Error())
		}
		if got != 1 {
			t.Errorf("wrong output, expected %d but got %d", 1, got)
		}
	}
}

func TestNewsRepositoryBase_Find(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	expected := 10
	populateNews(t, s.news, expected)
	got, err := s.news.Find(context.Background(), &model.NewsFindExpr{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if len(got) != expected {
		t.Errorf("wrong output, expected %d but got %v", expected, got)
	}
}

func TestNewsRepositoryBase_FindIter(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	expected := 10
	populateNews(t, s.news, expected)
	iter, err := s.news.FindIter(context.Background(), &model.NewsFindExpr{
		Columns: []string{model.TableNewsColumnID, model.TableNewsColumnContent},
	})
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
		if ent.Title != "" {
			t.Error("title should be empty")
		}
		got = append(got, ent)
	}
	if err = iter.Err(); err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	if len(got) != expected {
		t.Errorf("wrong output, expected %d but got %v", expected, got)
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

func TestNewsRepositoryBase_FindOneByTitle(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	expected := 10
	populateNews(t, s.news, expected)

	for i := 1; i <= expected; i++ {
		got, err := s.news.FindOneByTitle(context.Background(), fmt.Sprintf("title-%d", i))
		if err != nil {
			t.Fatalf("unexpected error: %s", err.Error())
		}
		if got.ID != int64(i) {
			t.Errorf("wrong id, expected %d but got %d", i, got.ID)
		}
	}
}

func TestNewsRepositoryBase_FindOneByTitleAndLead(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	expected := 10
	populateNews(t, s.news, expected)

	for i := 1; i <= expected; i++ {
		got, err := s.news.FindOneByTitleAndLead(context.Background(), fmt.Sprintf("title-%d", i), fmt.Sprintf("lead-%d", i))
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
	got, err := s.news.Count(context.Background(), &model.NewsCountExpr{
		Where: &model.NewsCriteria{
			Score: sql.NullFloat64{
				Float64: 10.11,
				Valid:   true,
			},
			Continue: sql.NullBool{Bool: true, Valid: true},
		},
	})
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
		query: "UPDATE example.news SET content=$1, title=$2, updated_at=NOW(), version=version+1 WHERE id=$3 RETURNING " + strings.Join(model.TableNewsColumns, ", "),
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
			ViewsDistribution: model.NullFloat64Array{
				Valid:        true,
				Float64Array: []float64{1.2, 2.3, 3.4, 4.5},
			},
			MetaData: []byte(`{"something": 1}`),
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
			Version: sql.NullInt64{
				Int64: 999,
				Valid: true,
			},
		},
		query: "UPDATE example.news SET content=$1, continue=$2, created_at=$3, lead=$4, meta_data=$5, score=$6, title=$7, updated_at=$8, version=$9, views_distribution=$10 WHERE id=$11 RETURNING " + strings.Join(model.TableNewsColumns, ", "),
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

func TestNewsRepositoryBase_UpdateOneByID(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	for hint, given := range testNewsInsertData {
		t.Run(hint, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			inserted, err := s.news.Insert(ctx, &given.entity)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			got, err := s.news.UpdateOneByID(ctx, inserted.ID, &model.NewsPatch{
				Title:   sql.NullString{String: inserted.Title + " (edited)", Valid: true},
				Lead:    sql.NullString{String: inserted.Lead.String + " (edited)", Valid: inserted.Lead.Valid},
				Content: sql.NullString{String: inserted.Content + " (edited)", Valid: true},
			})
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if !strings.Contains(got.Title, "(edited)") {
				t.Errorf("wrong title, should contains 'edited' but got %s", got.Title)
			}
			if given.entity.Lead.Valid {
				if !strings.Contains(got.Lead.String, "(edited)") {
					t.Errorf("wrong lead, should contains 'edited' but got %v", got.Lead)
				}
			}
			if !got.UpdatedAt.Valid {
				t.Error("updated at expected to be valid")
			}
			if got.CreatedAt.IsZero() {
				t.Error("created at should not be zero value")
			}
			if given.entity.Version >= got.Version {
				t.Errorf("new version should be higher than %d but got %d", given.entity.Version, got.Version)
			}
		})
	}
}

func TestNewsRepositoryBase_FindOneByIDAndUpdate(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	for hint, given := range testNewsInsertData {
		t.Run(hint, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			inserted, err := s.news.Insert(ctx, &given.entity)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			was, got, err := s.news.FindOneByIDAndUpdate(ctx, inserted.ID, &model.NewsPatch{
				Title:   sql.NullString{String: inserted.Title + " (edited)", Valid: true},
				Lead:    sql.NullString{String: inserted.Lead.String + " (edited)", Valid: inserted.Lead.Valid},
				Content: sql.NullString{String: inserted.Content + " (edited)", Valid: true},
			})
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if was.Title != given.entity.Title || was.Lead != given.entity.Lead || was.Version != given.entity.Version {
				t.Errorf("wrong previous value, expecte %v but got %v", given.entity, was)
			}
			if !strings.Contains(got.Title, "(edited)") {
				t.Errorf("wrong title, should contains 'edited' but got %s", got.Title)
			}
			if given.entity.Lead.Valid {
				if !strings.Contains(got.Lead.String, "(edited)") {
					t.Errorf("wrong lead, should contains 'edited' but got %v", got.Lead)
				}
			}
			if !got.UpdatedAt.Valid {
				t.Error("updated at expected to be valid")
			}
			if got.CreatedAt.IsZero() {
				t.Error("created at should not be zero value")
			}
			if given.entity.Version >= got.Version {
				t.Errorf("new version should be higher than %d but got %d", given.entity.Version, got.Version)
			}
		})
	}
}

var testNewsUpdateOneByTitleData = map[string]struct {
	patch model.NewsPatch
	query string
}{
	"minimum": {
		patch: model.NewsPatch{
			Content: sql.NullString{String: "content - minimum", Valid: true},
		},
		query: "UPDATE example.news SET content=$1, updated_at=NOW(), version=version+1 WHERE title=$2 RETURNING " + strings.Join(model.TableNewsColumns, ", "),
	},
	"full": {
		patch: model.NewsPatch{
			Lead: sql.NullString{
				Valid:  true,
				String: "lead - full",
			},
			Score: sql.NullFloat64{
				Valid:   true,
				Float64: 12.14,
			},
			ViewsDistribution: model.NullFloat64Array{
				Valid:        true,
				Float64Array: []float64{1.2, 2.3, 3.4, 4.5},
			},
			MetaData: []byte(`{"something": 1}`),
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
		query: "UPDATE example.news SET content=$1, continue=$2, created_at=$3, lead=$4, meta_data=$5, score=$6, updated_at=$7, version=version+1, views_distribution=$8 WHERE title=$9 RETURNING " + strings.Join(model.TableNewsColumns, ", "),
	},
}

func TestNewsRepositoryBase_UpdateOneByTitleQuery(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	for hint, given := range testNewsUpdateOneByTitleData {
		t.Run(hint, func(t *testing.T) {
			query, _, err := s.news.UpdateOneByTitleQuery("title", &given.patch)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if given.query != query {
				t.Errorf("wrong output, expected:\n	%s\nbut got:\n	%s", given.query, query)
			}
		})
	}
}

func TestNewsRepositoryBase_UpdateOneByTitle(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	expected := 10
	populateNews(t, s.news, expected)

	for i := 1; i <= expected; i++ {
		got, err := s.news.UpdateOneByTitle(context.Background(), fmt.Sprintf("title-%d", i), &model.NewsPatch{
			Content: sql.NullString{
				Valid:  true,
				String: fmt.Sprintf("content-updated-by-title-%d", i),
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %s", err.Error())
		}
		if got.ID != int64(i) {
			t.Errorf("wrong id, expected %d but got %d", i, got.ID)
		}
		if !strings.HasPrefix(got.Content, "content-updated-by-title") {
			t.Error("content was not updated properly")
		}
	}
}

func TestNewsRepositoryBase_UpdateOneByTitleAndLead(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	expected := 10
	populateNews(t, s.news, expected)

	for i := 1; i <= expected; i++ {
		got, err := s.news.UpdateOneByTitleAndLead(context.Background(), fmt.Sprintf("title-%d", i), fmt.Sprintf("lead-%d", i), &model.NewsPatch{
			Content: sql.NullString{
				Valid:  true,
				String: fmt.Sprintf("content-updated-by-title-and-lead-%d", i),
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %s", err.Error())
		}
		if got.ID != int64(i) {
			t.Errorf("wrong id, expected %d but got %d", i, got.ID)
		}
		if !strings.HasPrefix(got.Content, "content-updated-by-title-and-lead") {
			t.Error("content was not updated properly")
		}
	}
}

var testNewsUpsertData = map[string]struct {
	entity model.NewsEntity
	patch  model.NewsPatch
	query  string
}{
	"full": {
		patch: model.NewsPatch{
			Lead: sql.NullString{
				Valid:  true,
				String: "lead - full",
			},
			Score: sql.NullFloat64{
				Valid:   true,
				Float64: 12.14,
			},
			ViewsDistribution: model.NullFloat64Array{
				Valid:        true,
				Float64Array: []float64{1.2, 2.3, 3.4, 4.5},
			},
			MetaData: []byte(`{"something": 1}`),
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
		entity: model.NewsEntity{
			Title: "title - full",
			Lead: sql.NullString{
				Valid:  true,
				String: "lead - full",
			},
			ViewsDistribution: model.NullFloat64Array{
				Valid:        true,
				Float64Array: []float64{1.2, 2.3, 3.4, 4.5},
			},
			MetaData:  []byte(`{"something": 1}`),
			Score:     10.11,
			Content:   "content - full",
			Continue:  true,
			CreatedAt: time.Now(),
			UpdatedAt: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
		},
		query: "INSERT INTO example.news (content, continue, created_at, lead, meta_data, score, title, updated_at, version, views_distribution) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) ON CONFLICT (example.news_title_key) DO UPDATE SET content=$11, continue=$12, created_at=$13, lead=$14, meta_data=$15, score=$16, updated_at=$17, version=version+1, views_distribution=$18 RETURNING " + strings.Join(model.TableNewsColumns, ", "),
	},
}

func TestNewsRepositoryBase_UpsertQuery(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	for hint, given := range testNewsUpsertData {
		t.Run(hint, func(t *testing.T) {
			query, _, err := s.news.UpsertQuery(&given.entity, &given.patch, model.TableNewsConstraintTitleUnique)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if given.query != query {
				t.Errorf("wrong output, expected:\n	%s\nbut got:\n	%s", given.query, query)
			}
		})
	}
}
