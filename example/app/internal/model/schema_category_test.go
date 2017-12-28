package model_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/piotrkowalczuk/pqt/example/app/internal/model"
)

func assertCategoryEntity(t *testing.T, given, got *model.CategoryEntity) {
	if given.Name != got.Name {
		t.Errorf("wrong name, expected %s but got %s", given.Name, got.Name)
	}
	if given.Content != got.Content {
		t.Errorf("wrong content, expected %s but got %s", given.Content, got.Content)
	}
	if !given.UpdatedAt.Valid && got.UpdatedAt.Valid {
		t.Error("updated at expected to be invalid")
	}
	if got.CreatedAt.IsZero() {
		t.Error("created at should not be zero value")
	}
}

var testCategoryInsertData = map[string]struct {
	entity *model.CategoryEntity
	query  string
	assert func(*testing.T, *model.CategoryEntity, *model.CategoryEntity)
}{
	"minimum": {
		entity: &model.CategoryEntity{
			Name:    "name - minimum",
			Content: "content - minimum",
		},
		query:  "INSERT INTO example.category (content, name) VALUES ($1, $2) RETURNING content, created_at, id, name, parent_id, updated_at",
		assert: assertCategoryEntity,
	},
	"full": {
		entity: &model.CategoryEntity{
			Name:      "name - full",
			Content:   "content - full",
			CreatedAt: time.Now(),
			UpdatedAt: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
		},
		query:  "INSERT INTO example.category (content, created_at, name, updated_at) VALUES ($1, $2, $3, $4) RETURNING content, created_at, id, name, parent_id, updated_at",
		assert: assertCategoryEntity,
	},
}

func BenchmarkCategoryRepositoryBase_InsertQuery(b *testing.B) {
	s := setup(b)
	defer s.teardown(b)
	b.ResetTimer()

	for hint, given := range testCategoryInsertData {
		b.Run(hint, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				query, args, err := s.category.InsertQuery(given.entity, true)
				if err != nil {
					b.Fatalf("unexpected error: %s", err.Error())
				}
				benchQuery = query
				benchArgs = args
			}
		})
	}
}

var testCategoryFindData = map[string]struct {
	expr  model.CategoryFindExpr
	query string
}{
	"minimum": {
		expr: model.CategoryFindExpr{
			Where: &model.CategoryCriteria{
				Content: sql.NullString{String: "content - minimum", Valid: true},
			},
		},
		query: "SELECT t0.content, t0.created_at, t0.id, t0.name, t0.parent_id, t0.updated_at FROM example.category AS t0 WHERE t0.content=$1",
	},
	"logical-operator": {
		expr: model.CategoryFindExpr{
			Where: model.CategoryOr(
				&model.CategoryCriteria{
					Content: sql.NullString{String: "content - minimum", Valid: true},
				},

				model.CategoryAnd(
					&model.CategoryCriteria{
						Content: sql.NullString{String: "content - maximum", Valid: true},
					},
					model.CategoryOr(
						&model.CategoryCriteria{
							ParentID: sql.NullInt64{Int64: 20, Valid: true},
						},
						&model.CategoryCriteria{
							ParentID: sql.NullInt64{Int64: 10, Valid: true},
							Content:  sql.NullString{String: "content - minimum", Valid: true},
						},
					),
				),
			),
		},
		query: "SELECT t0.content, t0.created_at, t0.id, t0.name, t0.parent_id, t0.updated_at FROM example.category AS t0 WHERE (t0.content=$1) OR ((t0.content=$2) AND ((t0.parent_id=$3) OR (t0.content=$4 AND t0.parent_id=$5)))",
	},
	"full": {
		expr: model.CategoryFindExpr{
			Where: &model.CategoryCriteria{
				Content: sql.NullString{String: "content - full", Valid: true},
				CreatedAt: pq.NullTime{
					Valid: true,
					Time:  time.Now(),
				},
				Name: sql.NullString{
					String: "games",
					Valid:  true,
				},
				UpdatedAt: pq.NullTime{
					Valid: true,
					Time:  time.Now(),
				},
			},
		},
		query: "SELECT t0.content, t0.created_at, t0.id, t0.name, t0.parent_id, t0.updated_at FROM example.category AS t0 WHERE t0.content=$1 AND t0.created_at=$2 AND t0.name=$3 AND t0.updated_at=$4",
	},
}

func BenchmarkCategoryRepositoryBase_FindQuery(b *testing.B) {
	s := setup(b)
	defer s.teardown(b)
	b.ResetTimer()

	for hint, given := range testCategoryFindData {
		b.Run(hint, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				query, args, err := s.category.FindQuery(&given.expr)
				if err != nil {
					b.Fatalf("unexpected error: %s", err.Error())
				}
				benchQuery = query
				benchArgs = args
			}
		})
	}
}

func TestCategoryRepositoryBase_FindQuery(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	for hint, given := range testCategoryFindData {
		t.Run(hint, func(t *testing.T) {
			query, _, err := s.category.FindQuery(&given.expr)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if given.query != query {
				t.Errorf("wrong output, expected:\n	%s\nbut got:\n	%s", given.query, query)
			}
		})
	}
}

func TestCategoryRepositoryBase_DeleteOneByID(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	nb := 10
	populateCategory(t, s.category, nb)
	categories, err := s.category.Find(context.Background(), &model.CategoryFindExpr{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	for _, c := range categories {
		t.Log(c.Name, c.ID, c.ParentID.Int64)
	}
	for _, c := range categories {
		if !c.ParentID.Valid || c.ParentID.Int64 == 0 {
			// skip if parent
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			got, err := s.category.DeleteOneByID(context.Background(), c.ID)
			if err != nil {
				t.Log(c.ParentID)
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if got != 1 {
				t.Errorf("wrong output, expected %d but got %d", 1, got)
			}
		})
	}
}

func TestCategoryRepositoryBase_FindIter(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	max := 10
	populateCategory(t, s.category, max)
	iter, err := s.category.FindIter(context.Background(), &model.CategoryFindExpr{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	defer iter.Close()

	var got []*model.CategoryEntity
	for iter.Next() {
		ent, err := iter.Category()
		if err != nil {
			t.Fatalf("unexpected error: %s", err.Error())
		}
		got = append(got, ent)
	}
	if err = iter.Err(); err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	exp := (max * max) + max
	if len(got) != exp {
		t.Errorf("wrong output, expected %d but got %d", exp, len(got))
	}
}

func TestCategoryRepositoryBase_Count(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	max := 10
	populateCategory(t, s.category, max)
	got, err := s.category.Count(context.Background(), &model.CategoryCountExpr{
		Where: &model.CategoryCriteria{
			Content:  sql.NullString{String: "content-1-5", Valid: true},
			Name:     sql.NullString{String: "name-1-5", Valid: true},
			ParentID: sql.NullInt64{Int64: 1, Valid: true},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if got != 1 {
		t.Errorf("wrong output, expected %d but got %d", 1, got)
	}
}

func TestCategoryRepositoryBase_Find(t *testing.T) {
	cases := map[string]struct {
		max, exp int
		expr     model.CategoryFindExpr
	}{
		"simple": {
			max: 10,
			exp: 1,
			expr: model.CategoryFindExpr{
				Where: &model.CategoryCriteria{
					Content: sql.NullString{String: "content-5-1", Valid: true},
					Name:    sql.NullString{String: "name-5-1", Valid: true},
				},
				Limit:  10,
				Offset: 0,
				OrderBy: []model.RowOrder{
					{
						Name:       model.TableCategoryColumnID,
						Descending: true,
					},
				},
			},
		},
		"full": {
			max: 10,
			exp: 0,
			expr: model.CategoryFindExpr{
				Where: &model.CategoryCriteria{
					Content:   sql.NullString{String: "content-1-10", Valid: true},
					Name:      sql.NullString{String: "name-1-10", Valid: true},
					ParentID:  sql.NullInt64{Int64: 1, Valid: true},
					CreatedAt: pq.NullTime{Time: time.Now().Add(5 * time.Minute), Valid: true},
					UpdatedAt: pq.NullTime{Time: time.Now().Add(-5 * time.Minute), Valid: true},
				},
				Limit:  10,
				Offset: 10,
				OrderBy: []model.RowOrder{
					{
						Name:       model.TableCategoryColumnID,
						Descending: true,
					},
				},
			},
		},
	}

	for hint, c := range cases {
		t.Run(hint, func(t *testing.T) {
			s := setup(t)
			defer s.teardown(t)

			populateCategory(t, s.category, c.max)
			got, err := s.category.Find(context.Background(), &c.expr)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			if len(got) != c.exp {
				t.Errorf("wrong output, expected %d but got %d", c.exp, len(got))
			}
		})
	}
}

func TestCategoryRepositoryBase_FindOneByID(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	expected := 5
	populateCategory(t, s.category, expected)

	for i := 1; i <= expected; i++ {
		got, err := s.category.FindOneByID(context.Background(), int64(i))
		if err != nil {
			t.Fatalf("unexpected error: %s", err.Error())
		}
		if got.ID != int64(i) {
			t.Errorf("wrong id, expected %d but got %d", i, got.ID)
		}
	}
}

func TestCategoryRepositoryBase_UpdateOneByID(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	expected := 10
	populateCategory(t, s.category, expected)

	for i := 1; i <= expected; i++ {
		patch := &model.CategoryPatch{
			Content: sql.NullString{
				Valid:  true,
				String: fmt.Sprintf("content-updated-by-id-%d", i),
			},

			Name: sql.NullString{
				Valid:  true,
				String: fmt.Sprintf("name-updated-by-id-%d", i),
			},
			ParentID: sql.NullInt64{
				Valid: true,
				Int64: 2,
			},

			CreatedAt: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
		}

		if i%2 == 0 {
			patch.UpdatedAt = pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			}
		}
		got, err := s.category.UpdateOneByID(context.Background(), int64(i), patch)
		if err != nil {
			t.Fatalf("unexpected error: %s", err.Error())
		}
		if got.ID != int64(i) {
			t.Errorf("wrong id, expected %d but got %d", i, got.ID)
		}
		if !strings.HasPrefix(got.Content, "content-updated-by-id") {
			t.Error("content was not updated properly")
		}
	}
}

var testCategoryUpsertData = map[string]struct {
	entity model.CategoryEntity
	patch  model.CategoryPatch
	query  string
}{
	"full": {
		patch: model.CategoryPatch{
			Name: sql.NullString{
				Valid:  true,
				String: "name - full",
			},
			Content: sql.NullString{String: "content - full", Valid: true},
			CreatedAt: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
			ParentID: sql.NullInt64{
				Int64: 2,
				Valid: true,
			},
			UpdatedAt: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
		},
		entity: model.CategoryEntity{
			ID:      1,
			Name:    "name - full",
			Content: "content - full",
			ParentID: sql.NullInt64{
				Int64: 2,
				Valid: true,
			},
			CreatedAt: time.Now(),
			UpdatedAt: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
		},
		query: "INSERT INTO example.category (content, created_at, name, parent_id, updated_at) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (example.category_id_pkey) DO UPDATE SET content=$6, created_at=$7, name=$8, parent_id=$9, updated_at=$10 RETURNING " + strings.Join(model.TableCategoryColumns, ", "),
	},
}

func TestCategoryRepositoryBase_UpsertQuery(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	for hint, given := range testCategoryUpsertData {
		t.Run(hint, func(t *testing.T) {
			query, _, err := s.category.UpsertQuery(&given.entity, &given.patch, model.TableCategoryConstraintPrimaryKey)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if given.query != query {
				t.Errorf("wrong output, expected:\n	%s\nbut got:\n	%s", given.query, query)
			}
		})
	}
}
