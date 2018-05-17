package model_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/piotrkowalczuk/pqt/example/app/internal/model"
)

var testPackageInsertData = map[string]struct {
	entity model.PackageEntity
	query  string
}{
	"minimum": {
		entity: model.PackageEntity{
			Break: sql.NullString{String: "break - minimum", Valid: true},
		},
		query: "INSERT INTO example.package (break) VALUES ($1) RETURNING " + strings.Join(model.TablePackageColumns, ", "),
	},
	"full": {
		entity: model.PackageEntity{
			Break:      sql.NullString{String: "break - minimum", Valid: true},
			CreatedAt:  time.Now(),
			CategoryID: sql.NullInt64{Int64: 100, Valid: true},
			UpdatedAt: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
		},
		query: "INSERT INTO example.package (break, category_id, created_at, updated_at) VALUES ($1, $2, $3, $4) RETURNING " + strings.Join(model.TablePackageColumns, ", "),
	},
}

func BenchmarkPackageRepositoryBase_InsertQuery(b *testing.B) {
	s := setup(b)
	defer s.teardown(b)
	b.ResetTimer()

	for hint, given := range testPackageInsertData {
		b.Run(hint, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				query, args, err := s.pkg.InsertQuery(&given.entity, true)
				if err != nil {
					b.Fatalf("unexpected error: %s", err.Error())
				}
				benchQuery = query
				benchArgs = args
			}
		})
	}
}

func TestPackageRepositoryBase_InsertQuery(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	for hint, given := range testPackageInsertData {
		t.Run(hint, func(t *testing.T) {
			query, _, err := s.pkg.InsertQuery(&given.entity, true)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if given.query != query {
				t.Errorf("wrong output, expected:\n	%s\nbut got:\n	%s", given.query, query)
			}
		})
	}
}

func TestPackageRepositoryBase_Insert(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	for hint, given := range testPackageInsertData {
		t.Run(hint, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			category, err := s.category.Insert(ctx, &model.CategoryEntity{Content: "content"})
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			given.entity.CategoryID = sql.NullInt64{Int64: category.ID, Valid: true}

			got, err := s.pkg.Insert(ctx, &given.entity)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if given.entity.Break != got.Break {
				t.Errorf("wrong break, expected %v but got %v", given.entity.Break, got.Break)
			}
			if given.entity.CategoryID != got.CategoryID {
				t.Errorf("wrong category id, expected %v but got %v", given.entity.CategoryID, got.CategoryID)
			}
			if !given.entity.UpdatedAt.Valid && got.UpdatedAt.Valid {
				t.Error("updated at expected to be invalid")
			}
			if got.CreatedAt.IsZero() {
				t.Error("created at should not be zero value")
			}
		})
	}
}

var testPackageFindData = map[string]struct {
	expr  model.PackageFindExpr
	query string
}{
	"minimum": {
		expr: model.PackageFindExpr{
			Where: &model.PackageCriteria{
				Break: sql.NullString{String: "break - minimum", Valid: true},
			},
		},
		query: "SELECT " + join(model.TablePackageColumns, 0) + " FROM example.package AS t0 WHERE t0.break=$1",
	},
	"logical-condition": {
		expr: model.PackageFindExpr{
			Where: model.PackageOr(
				model.PackageAnd(
					&model.PackageCriteria{
						Break: sql.NullString{String: "break - minimum", Valid: true},
					},
					&model.PackageCriteria{
						ID: sql.NullInt64{Int64: 10, Valid: true},
					},
				),
				model.PackageAnd(
					&model.PackageCriteria{
						Break: sql.NullString{String: "break - 1", Valid: true},
					},
					&model.PackageCriteria{
						Break: sql.NullString{String: "break - 2", Valid: true},
					},
					&model.PackageCriteria{
						ID:    sql.NullInt64{Int64: 1000, Valid: true},
						Break: sql.NullString{String: "break - 3", Valid: true},
					},
				),
			),
		},
		query: "SELECT " + join(model.TablePackageColumns, 0) + " FROM example.package AS t0 WHERE ((t0.break=$1) AND ()) OR ((t0.break=$2) AND (t0.break=$3) AND (t0.break=$4))",
	},
	"full": {
		expr: model.PackageFindExpr{
			Where: &model.PackageCriteria{
				ID: sql.NullInt64{
					Int64: 1,
					Valid: true,
				},
				Break:      sql.NullString{String: "break - full", Valid: true},
				CategoryID: sql.NullInt64{Int64: 100, Valid: true},
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
					Name:       model.TablePackageColumnBreak,
					Descending: true,
				},
				{
					Name: model.TablePackageColumnID,
				},
			},
		},
		query: "SELECT " + join(model.TablePackageColumns, 0) + " FROM example.package AS t0 WHERE t0.break=$1 AND t0.category_id=$2 AND t0.created_at=$3 AND t0.updated_at=$4 ORDER BY break DESC, id OFFSET $5  LIMIT $6 ",
	},
}

func BenchmarkPackageRepositoryBase_FindQuery(b *testing.B) {
	s := setup(b)
	defer s.teardown(b)
	b.ResetTimer()

	for hint, given := range testPackageFindData {
		b.Run(hint, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				query, args, err := s.pkg.FindQuery(&given.expr)
				if err != nil {
					b.Fatalf("unexpected error: %s", err.Error())
				}
				benchQuery = query
				benchArgs = args
			}
		})
	}
}

func TestPackageRepositoryBase_FindQuery(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	for hint, given := range testPackageFindData {
		t.Run(hint, func(t *testing.T) {
			query, _, err := s.pkg.FindQuery(&given.expr)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if given.query != query {
				t.Errorf("wrong output, expected:\n	%s\nbut got:\n	%s", given.query, query)
			}
		})
	}
}

func TestPackageRepositoryBase_DeleteOneByID(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	nb := 10
	populatePackage(t, s.pkg, nb)
	for i := 1; i <= nb; i++ {
		got, err := s.pkg.DeleteOneByID(context.Background(), int64(i))
		if err != nil {
			t.Fatalf("unexpected error: %s", err.Error())
		}
		if got != 1 {
			t.Errorf("wrong output, expected %d but got %d", 1, got)
		}
	}
}

func TestPackageRepositoryBase_Find(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	expected := 10
	populatePackage(t, s.pkg, expected)
	got, err := s.pkg.Find(context.Background(), &model.PackageFindExpr{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if len(got) != expected {
		t.Errorf("wrong output, expected %d but got %v", expected, got)
	}
}

func TestPackageRepositoryBase_FindIter(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	expected := 10
	populatePackage(t, s.pkg, expected)
	iter, err := s.pkg.FindIter(context.Background(), &model.PackageFindExpr{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	defer iter.Close()

	var got []*model.PackageEntity
	for iter.Next() {
		ent, err := iter.Package()
		if err != nil {
			t.Fatalf("unexpected error: %s", err.Error())
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

func TestPackageRepositoryBase_FindOneByID(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	expected := 10
	populatePackage(t, s.pkg, expected)

	for i := 1; i <= expected; i++ {
		got, err := s.pkg.FindOneByID(context.Background(), int64(i))
		if err != nil {
			t.Fatalf("unexpected error: %s", err.Error())
		}
		if got.ID != int64(i) {
			t.Errorf("wrong id, expected %d but got %d", i, got.ID)
		}
	}
}

func TestPackageRepositoryBase_Count(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	expected := 1
	populatePackage(t, s.pkg, 3)
	got, err := s.pkg.Count(context.Background(), &model.PackageCountExpr{
		Where: &model.PackageCriteria{
			Break: sql.NullString{String: "break-1", Valid: true},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if got != int64(expected) {
		t.Errorf("wrong output, expected %d but got %d", expected, got)
	}
}

var testPackageUpdateData = map[string]struct {
	patch model.PackagePatch
	query string
}{
	"minimum": {
		patch: model.PackagePatch{
			Break: sql.NullString{String: "break - minimum", Valid: true},
		},
		query: "UPDATE example.package SET break=$1, updated_at=NOW() WHERE id=$2 RETURNING " + strings.Join(model.TablePackageColumns, ", "),
	},
	"full": {
		patch: model.PackagePatch{
			Break:      sql.NullString{String: "break - full", Valid: true},
			CategoryID: sql.NullInt64{Int64: 100, Valid: true},
			CreatedAt: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
			UpdatedAt: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
		},
		query: "UPDATE example.package SET break=$1, category_id=$2, created_at=$3, updated_at=$4 WHERE id=$5 RETURNING " + strings.Join(model.TablePackageColumns, ", "),
	},
}

func BenchmarkPackageRepositoryBase_UpdateOneByIDQuery(b *testing.B) {
	s := setup(b)
	defer s.teardown(b)
	b.ResetTimer()

	for hint, given := range testPackageUpdateData {
		b.Run(hint, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				query, args, err := s.pkg.UpdateOneByIDQuery(1, &given.patch)
				if err != nil {
					b.Fatalf("unexpected error: %s", err.Error())
				}
				benchQuery = query
				benchArgs = args
			}
		})
	}
}

func TestPackageRepositoryBase_UpdateOneByIDQuery(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	for hint, given := range testPackageUpdateData {
		t.Run(hint, func(t *testing.T) {
			query, _, err := s.pkg.UpdateOneByIDQuery(1, &given.patch)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if given.query != query {
				t.Errorf("wrong output, expected:\n	%s\nbut got:\n	%s", given.query, query)
			}
		})
	}
}

func TestPackageRepositoryBase_UpdateOneByID(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	for hint, given := range testPackageInsertData {
		t.Run(hint, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			category, err := s.category.Insert(ctx, &model.CategoryEntity{Content: "content"})
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			given.entity.CategoryID = sql.NullInt64{Int64: category.ID, Valid: true}

			inserted, err := s.pkg.Insert(ctx, &given.entity)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			category, err = s.category.Insert(ctx, &model.CategoryEntity{Content: "content"})
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			got, err := s.pkg.UpdateOneByID(ctx, inserted.ID, &model.PackagePatch{
				Break:      sql.NullString{String: inserted.Break.String + " (edited)", Valid: inserted.Break.Valid},
				CategoryID: sql.NullInt64{Int64: category.ID, Valid: true},
			})
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			if given.entity.Break.Valid {
				if !strings.Contains(got.Break.String, "(edited)") {
					t.Errorf("wrong break, should contains 'edited' but got %v", got.Break)
				}
			}
			if !got.UpdatedAt.Valid {
				t.Error("updated at expected to be valid")
			}
			if got.CreatedAt.IsZero() {
				t.Error("created at should not be zero value")
			}
		})
	}
}

var testPackageUpsertData = map[string]struct {
	entity model.PackageEntity
	patch  model.PackagePatch
	query  string
}{
	"full": {
		patch: model.PackagePatch{
			Break: sql.NullString{String: "title - full", Valid: true},
			CreatedAt: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
			UpdatedAt: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
		},
		entity: model.PackageEntity{
			Break:     sql.NullString{String: "title - full", Valid: true},
			CreatedAt: time.Now(),
			UpdatedAt: pq.NullTime{
				Valid: true,
				Time:  time.Now(),
			},
		},
		query: "INSERT INTO example.package (break, created_at, updated_at) VALUES ($1, $2, $3) ON CONFLICT (example.package_category_id_fkey) DO UPDATE SET break=$4, created_at=$5, updated_at=$6 RETURNING " + strings.Join(model.TablePackageColumns, ", "),
	},
}

func TestPackageRepositoryBase_UpsertQuery(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)

	for hint, given := range testPackageUpsertData {
		t.Run(hint, func(t *testing.T) {

			query, _, err := s.pkg.UpsertQuery(&given.entity, &given.patch, model.TablePackageConstraintCategoryIDForeignKey)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if given.query != query {
				t.Errorf("wrong output, expected:\n	%s\nbut got:\n	%s", given.query, query)
			}
		})
	}
}
