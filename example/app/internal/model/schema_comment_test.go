package model_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"reflect"

	"github.com/lib/pq"
	"github.com/piotrkowalczuk/pqt/example/app/internal/model"
)

var testCommentFindData = map[string]struct {
	expr  model.CommentFindExpr
	query string
}{
	"minimum": {
		expr: model.CommentFindExpr{
			Where: &model.CommentCriteria{
				Content: sql.NullString{String: "content - minimum", Valid: true},
			},
		},
		query: "SELECT t0.content, t0.created_at, t0.id, multiply(t0.id, t0.id) AS id_multiply, t0.news_id, t0.news_title, now() AS right_now, t0.updated_at FROM example.comment AS t0 WHERE t0.content=$1",
	},
	"logical-operator": {
		expr: model.CommentFindExpr{
			Where: model.CommentOr(
				&model.CommentCriteria{
					Content: sql.NullString{String: "content - minimum", Valid: true},
				},

				model.CommentAnd(
					&model.CommentCriteria{
						Content: sql.NullString{String: "content - maximum", Valid: true},
					},
					model.CommentOr(
						&model.CommentCriteria{
							NewsID: sql.NullInt64{Int64: 20, Valid: true},
						},
						&model.CommentCriteria{
							NewsID:  sql.NullInt64{Int64: 10, Valid: true},
							Content: sql.NullString{String: "content - minimum", Valid: true},
						},
					),
				),
			),
		},
		query: "SELECT t0.content, t0.created_at, t0.id, multiply(t0.id, t0.id) AS id_multiply, t0.news_id, t0.news_title, now() AS right_now, t0.updated_at FROM example.comment AS t0 WHERE (t0.content=$1) OR ((t0.content=$2) AND ((t0.news_id=$3) OR (t0.content=$4 AND t0.news_id=$5)))",
	},
	"minimum-join-news-by-id": {
		expr: model.CommentFindExpr{
			Where: &model.CommentCriteria{
				Content: sql.NullString{String: "content - minimum", Valid: true},
			},
			JoinNewsByTitle: &model.NewsJoin{
				Kind: model.JoinInner,
			},
		},
		query: "SELECT t0.content, t0.created_at, t0.id, multiply(t0.id, t0.id) AS id_multiply, t0.news_id, t0.news_title, now() AS right_now, t0.updated_at FROM example.comment AS t0 INNER JOIN example.news AS t1 ON t0.news_title=t1.title WHERE t0.content=$1",
	},
	"full": {
		expr: model.CommentFindExpr{
			Where: &model.CommentCriteria{
				Content: sql.NullString{String: "content - full", Valid: true},
				CreatedAt: pq.NullTime{
					Valid: true,
					Time:  time.Now(),
				},
				IDMultiply: sql.NullInt64{
					Int64: 10,
					Valid: true,
				},
				UpdatedAt: pq.NullTime{
					Valid: true,
					Time:  time.Now(),
				},
			},
			JoinNewsByID: &model.NewsJoin{
				Kind:  model.JoinLeft,
				Fetch: true,
				On: &model.NewsCriteria{
					Title: sql.NullString{String: "title - minimum", Valid: true},
				},
				Where: &model.NewsCriteria{
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
			},
		},
		query: "SELECT t0.content, t0.created_at, t0.id, multiply(t0.id, t0.id) AS id_multiply, t0.news_id, t0.news_title, now() AS right_now, t0.updated_at, " + join(model.TableNewsColumns, 2) + " FROM example.comment AS t0 LEFT JOIN example.news AS t2 ON t0.news_id=t2.id AND t2.title=$1 WHERE t0.content=$2 AND t0.created_at=$3 AND multiply(t0.id, t0.id)=$4 AND t0.updated_at=$5 AND t2.content=$6 AND t2.continue=$7 AND t2.created_at=$8 AND t2.lead=$9 AND t2.meta_data=$10 AND t2.score=$11 AND t2.title=$12 AND t2.updated_at=$13 AND t2.views_distribution=$14",
	},
}

func BenchmarkCommentRepositoryBase_FindQuery(b *testing.B) {
	s := setup(b)
	defer s.teardown(b)
	b.ResetTimer()

	for hint, given := range testCommentFindData {
		b.Run(hint, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				query, args, err := s.comment.FindQuery(&given.expr)
				if err != nil {
					b.Fatalf("unexpected error: %s", err.Error())
				}
				benchQuery = query
				benchArgs = args
			}
		})
	}
}

func TestCommentRepositoryBase_FindQuery(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	for hint, given := range testCommentFindData {
		t.Run(hint, func(t *testing.T) {
			query, _, err := s.comment.FindQuery(&given.expr)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if given.query != query {
				t.Errorf("wrong output, expected:\n	%s\nbut got:\n	%s", given.query, query)
			}
		})
	}
}

func TestCommentRepositoryBase_Find(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	expected := 10
	populateNews(t, s.news, expected)
	populateComment(t, s.comment, expected)
	got, err := s.comment.Find(context.Background(), &model.CommentFindExpr{
		JoinNewsByID: &model.NewsJoin{
			Kind:  model.JoinLeft,
			Fetch: true,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	if len(got) != expected {
		t.Errorf("wrong output, expected %d but got %d", expected, len(got))
	}
	for _, g := range got {
		if g.NewsByID == nil || reflect.DeepEqual(*g.NewsByID, model.NewsEntity{}) {
			t.Errorf("news expected to be fetched, got: %#v", g.NewsByID)
		}
		if g.RightNow.IsZero() {
			t.Error("dynamic column right_now is invalid")
		} else {
			t.Logf("dynamic column right_now has value: %v", g.RightNow)
		}
		if g.IDMultiply == 0 {
			t.Error("dynamic column id_multiply is zero")
		} else if (g.ID.Int64 * g.ID.Int64) != (g.IDMultiply) {
			t.Logf("dynamic column id_multiply has wrong value, expected %d but got %d", (g.ID.Int64 * g.ID.Int64), g.IDMultiply)
		} else {
			t.Logf("dynamic column id_multiply has value: %v", g.IDMultiply)
		}
	}
}

func TestCommentRepositoryBase_FindIter(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	expected := 10
	populateNews(t, s.news, expected)
	populateComment(t, s.comment, expected)
	iter, err := s.comment.FindIter(context.Background(), &model.CommentFindExpr{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	defer iter.Close()

	var got []*model.CommentEntity
	for iter.Next() {
		ent, err := iter.Comment()
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

func TestCommentRepositoryBase_Count(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	expected := 10
	populateNews(t, s.news, expected)
	populateComment(t, s.comment, expected)
	got, err := s.comment.Count(context.Background(), &model.CommentCountExpr{
		JoinNewsByID: &model.NewsJoin{
			Kind: model.JoinLeft,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	if got != int64(expected) {
		t.Errorf("wrong output, expected %d but got %d", expected, got)
	}
}

func populateComment(t testing.TB, r *model.CommentRepositoryBase, nb int) {
	for i := 1; i <= nb; i++ {
		_, err := r.Insert(context.Background(), &model.CommentEntity{
			NewsID:    int64(i),
			NewsTitle: fmt.Sprintf("title-%d", i),
			Content:   fmt.Sprintf("content-%d", i),
		})
		if err != nil {
			t.Fatalf("unexpected error #%d: %s", i, err.Error())
		}
	}
}
