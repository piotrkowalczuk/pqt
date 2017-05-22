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
			Content:  sql.NullString{String: "content-5-1", Valid: true},
			Name:     sql.NullString{String: "name-5-1", Valid: true},
			ParentID: sql.NullInt64{Int64: 5, Valid: true},
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
					Content:  sql.NullString{String: "content-5-1", Valid: true},
					Name:     sql.NullString{String: "name-5-1", Valid: true},
					ParentID: sql.NullInt64{Int64: 5, Valid: true},
				},
				Limit:  10,
				Offset: 0,
				OrderBy: map[string]bool{
					model.TableCategoryColumnID: true,
				},
			},
		},
		"full": {
			max: 10,
			exp: 0,
			expr: model.CategoryFindExpr{
				Where: &model.CategoryCriteria{
					Content:   sql.NullString{String: "content-5-1", Valid: true},
					Name:      sql.NullString{String: "name-5-1", Valid: true},
					ParentID:  sql.NullInt64{Int64: 5, Valid: true},
					CreatedAt: pq.NullTime{Time: time.Now().Add(5 * time.Minute), Valid: true},
					UpdatedAt: pq.NullTime{Time: time.Now().Add(-5 * time.Minute), Valid: true},
				},
				Limit:  10,
				Offset: 10,
				OrderBy: map[string]bool{
					model.TableCategoryColumnID: true,
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
		got, err := s.category.UpdateOneByID(context.Background(), int64(i), &model.CategoryPatch{
			Content: sql.NullString{
				Valid:  true,
				String: fmt.Sprintf("content-updated-by-id-%d", i),
			},
			Name: sql.NullString{
				Valid:  true,
				String: fmt.Sprintf("name-updated-by-id-%d", i),
			},
			CreatedAt: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
			UpdatedAt: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
		})
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
