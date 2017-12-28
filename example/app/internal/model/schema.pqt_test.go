package model_test

import (
	"context"
	"testing"
	"time"

	"github.com/piotrkowalczuk/pqt/example/app/internal/model"
)

func TestCategoryRepositoryBase_InsertQuery(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)
	for hint, given := range testCategoryInsertData {
		t.Run(hint, func(t *testing.T) {
			query, _, err := s.GetCategory().InsertQuery(given.entity, true)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if given.query != query {
				t.Errorf("wrong output, expected:\n	%s\nbut got:\n	%s", given.query, query)
			}
		})
	}
}

func TestCategoryRepositoryBase_Insert(t *testing.T) {
	testCategoryRepositoryBaseInsert(t, 1000)
}

func testCategoryRepositoryBaseInsert(t *testing.T, n int) {
	s := setup(t)
	defer s.teardown(t)
	t.Skip()
	for ent := range model.GenerateCategoryEntity(n) {
		t.Run("", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			got, err := s.GetCategory().Insert(ctx, ent)
			if err != nil {
				t.Fatalf("unexpected error: %s\nfor entity: %#v", err.Error(), ent)
			}
			assertCategoryEntity(t, ent, got)
		})
	}
}

func TestPackageRepositoryBase_InsertQuery(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)
	for hint, given := range testPackageInsertData {
		t.Run(hint, func(t *testing.T) {
			query, _, err := s.GetPackage().InsertQuery(given.entity, true)
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
	testPackageRepositoryBaseInsert(t, 1000)
}

func testPackageRepositoryBaseInsert(t *testing.T, n int) {
	s := setup(t)
	defer s.teardown(t)
	t.Skip()
	for ent := range model.GeneratePackageEntity(n) {
		t.Run("", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			got, err := s.GetPackage().Insert(ctx, ent)
			if err != nil {
				t.Fatalf("unexpected error: %s\nfor entity: %#v", err.Error(), ent)
			}
			assertPackageEntity(t, ent, got)
		})
	}
}

func TestNewsRepositoryBase_InsertQuery(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)
	for hint, given := range testNewsInsertData {
		t.Run(hint, func(t *testing.T) {
			query, _, err := s.GetNews().InsertQuery(given.entity, true)
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
	testNewsRepositoryBaseInsert(t, 1000)
}

func testNewsRepositoryBaseInsert(t *testing.T, n int) {
	s := setup(t)
	defer s.teardown(t)
	uniqueTitle := make(map[string]struct{})
	for ent := range model.GenerateNewsEntity(n) {
		if _, ok := uniqueTitle[ent.Title]; ok {
			continue
		} else {
			uniqueTitle[ent.Title] = struct{}{}
		}
		t.Run("", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			got, err := s.GetNews().Insert(ctx, ent)
			if err != nil {
				t.Fatalf("unexpected error: %s\nfor entity: %#v", err.Error(), ent)
			}
			assertNewsEntity(t, ent, got)
		})
	}
}

func TestCommentRepositoryBase_InsertQuery(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)
	for hint, given := range testCommentInsertData {
		t.Run(hint, func(t *testing.T) {
			query, _, err := s.GetComment().InsertQuery(given.entity, true)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if given.query != query {
				t.Errorf("wrong output, expected:\n	%s\nbut got:\n	%s", given.query, query)
			}
		})
	}
}

func TestCommentRepositoryBase_Insert(t *testing.T) {
	testCommentRepositoryBaseInsert(t, 1000)
}

func testCommentRepositoryBaseInsert(t *testing.T, n int) {
	s := setup(t)
	defer s.teardown(t)
	t.Skip()
	for ent := range model.GenerateCommentEntity(n) {
		t.Run("", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			got, err := s.GetComment().Insert(ctx, ent)
			if err != nil {
				t.Fatalf("unexpected error: %s\nfor entity: %#v", err.Error(), ent)
			}
			assertCommentEntity(t, ent, got)
		})
	}
}

func TestCompleteRepositoryBase_InsertQuery(t *testing.T) {
	s := setup(t)
	defer s.teardown(t)
	for hint, given := range testCompleteInsertData {
		t.Run(hint, func(t *testing.T) {
			query, _, err := s.GetComplete().InsertQuery(given.entity, true)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if given.query != query {
				t.Errorf("wrong output, expected:\n	%s\nbut got:\n	%s", given.query, query)
			}
		})
	}
}

func TestCompleteRepositoryBase_Insert(t *testing.T) {
	testCompleteRepositoryBaseInsert(t, 1000)
}

func testCompleteRepositoryBaseInsert(t *testing.T, n int) {
	s := setup(t)
	defer s.teardown(t)
	for ent := range model.GenerateCompleteEntity(n) {
		if len(ent.ColumnJsonNn) == 0 {
			continue
		}
		if len(ent.ColumnJsonNnD) == 0 {
			continue
		}
		if len(ent.ColumnJsonbNn) == 0 {
			continue
		}
		if len(ent.ColumnJsonbNnD) == 0 {
			continue
		}
		t.Run("", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			got, err := s.GetComplete().Insert(ctx, ent)
			if err != nil {
				t.Fatalf("unexpected error: %s\nfor entity: %#v", err.Error(), ent)
			}
			assertCompleteEntity(t, ent, got)
		})
	}
}
