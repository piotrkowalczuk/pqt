package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/piotrkowalczuk/pqt/example/app/internal/model"
)

var (
	address string
	dbg     bool
)

func init() {
	flag.StringVar(&address, "addr", "postgres://localhost:5432/test?sslmode=disable", "postgres connection string")
	flag.BoolVar(&dbg, "dbg", true, "debug mode")
}

func main() {
	flag.Parse()
	db, err := sql.Open("postgres", address)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(model.SQL)
	if err != nil {
		log.Fatal(err)
	}

	repo := struct {
		news     model.NewsRepositoryBase
		comment  model.CommentRepositoryBase
		category model.CategoryRepositoryBase
	}{
		news: model.NewsRepositoryBase{
			DB:    db,
			Table: model.TableNews,
		},
		comment: model.CommentRepositoryBase{
			DB:    db,
			Table: model.TableComment,
		},
		category: model.CategoryRepositoryBase{
			DB:    db,
			Table: model.TableCategory,
		},
	}

	ctx := context.Background()

	count, err := repo.news.Count(ctx, &model.NewsCountExpr{})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("number of news fetched", "count", count)

	news, err := repo.news.Insert(ctx, &model.NewsEntity{
		Title:   fmt.Sprintf("Lorem Ipsum - %d - %d", time.Now().Unix(), rand.Int63()),
		Content: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nullam a felis vel erat gravida luctus at id nisi. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Vivamus a nibh massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Fusce viverra quam id dolor facilisis ultrices. Donec blandit, justo sit amet consequat gravida, nisi velit efficitur neque, ac ullamcorper leo dui vitae lorem. Pellentesque vitae ligula id massa fringilla facilisis eu sit amet neque. Ut ac fringilla mi. Maecenas id fermentum massa. Duis at tristique felis, nec aliquet nisi. Suspendisse potenti. In sed dolor maximus, dapibus arcu vitae, vehicula ligula. Nunc imperdiet eu ipsum sed pretium. Nullam iaculis nunc id dictum auctor.",
		Lead:    sql.NullString{String: "Neque porro quisquam est qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit...", Valid: true},
	})
	if err != nil {
		switch model.ErrorConstraint(err) {
		case model.TableNewsConstraintTitleUnique:
			log.Fatal(errors.New("news with such title already exists"))
		default:
			log.Fatal(err)
		}
	}

	nb := 20
	for i := 0; i < nb; i++ {
		_, err = repo.comment.Insert(ctx, &model.CommentEntity{
			NewsID:    news.ID,
			NewsTitle: news.Title,
			Content:   "Etiam eget nunc vel tellus placerat accumsan. Quisque dictum commodo orci, a eleifend nulla viverra malesuada. Etiam dui purus, dapibus a risus sed, porta scelerisque lorem. Sed vehicula mauris tellus, at dapibus risus facilisis vitae. Sed at lacus mollis, cursus sapien eu, egestas ligula. Cras blandit, arcu quis aliquam dictum, nibh purus pulvinar turpis, in dapibus est nibh et enim. Donec ex arcu, iaculis eget euismod id, lobortis nec enim. Quisque sed massa vel dui convallis ultrices. Nulla rutrum sed lacus vel ornare. Aliquam vulputate condimentum elit at pellentesque. Curabitur vitae sem tincidunt, volutpat urna ut, consequat turpis. Pellentesque varius justo libero, a volutpat lacus vulputate at. Integer tristique pharetra urna vel pharetra. In porttitor tincidunt eros, vel eleifend quam elementum a.",
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	iter, err := repo.comment.FindIter(ctx, &model.CommentFindExpr{
		OrderBy: []model.RowOrder{
			{
				Name: "id",
			},
			{
				Name:       "non_existing_column",
				Descending: true,
			},
		},
		Where: &model.CommentCriteria{
			NewsID: sql.NullInt64{Int64: news.ID, Valid: true},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer iter.Close()

	got := 0
	for iter.Next() {
		com, err := iter.Comment()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("comment fetched", "comment_id", com.ID)
		got++
	}
	if err = iter.Err(); err != nil {
		log.Fatal(err)
	}
	if nb != got {
		log.Fatal(fmt.Errorf("wrong number of comments, expected %d but got %d", nb, got))
	} else {
		log.Println("proper number of comments")
	}

	category, err := repo.category.Insert(ctx, &model.CategoryEntity{
		Name: "parent",
	})
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < nb; i++ {
		_, err := repo.category.Insert(ctx, &model.CategoryEntity{
			ParentID: sql.NullInt64{Int64: category.ID, Valid: true},
			Name:     "child_category" + strconv.Itoa(i),
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	count, err = repo.category.Count(ctx, &model.CategoryCountExpr{
		Where: &model.CategoryCriteria{
			ParentID: sql.NullInt64{Int64: category.ID, Valid: true},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if count != int64(nb) {
		log.Fatal(fmt.Errorf("wrong number of categories, expected %d but got %d", nb, count))
	} else {
		log.Println("proper number of categories")
	}

	_, err = repo.category.Insert(ctx, &model.CategoryEntity{
		ParentID: sql.NullInt64{Int64: int64(math.MaxInt64 - 1), Valid: true},
		Name:     "does not work",
	})
	if err != nil {
		switch model.ErrorConstraint(err) {
		case model.TableCategoryConstraintParentIDForeignKey:
			log.Println("category parent id constraint properly catched, category with such id does not exists")
		default:
			log.Fatal(fmt.Errorf("category constraint not catched properly, expected %s but got %s", model.TableCategoryConstraintParentIDForeignKey, model.ErrorConstraint(err)))
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()

	count, err = repo.news.Count(ctx, &model.NewsCountExpr{})
	if err != nil {
		if err != context.DeadlineExceeded {
			log.Fatal(err)
		}
		log.Fatal("as expected, news count failed due to deadline")
	}
}
