package pqt_test

import (
	"testing"

	"github.com/piotrkowalczuk/pqt"
)

func TestTable_AddColumn(t *testing.T) {
	c1 := &pqt.Column{Name: "c1"}
	c2 := &pqt.Column{Name: "c2"}
	c3 := &pqt.Column{Name: "c3"}

	tbl := pqt.NewTable("test").
		AddColumn(c1).
		AddColumn(c2).
		AddColumn(c3)

	if len(tbl.Columns) != 3 {
		t.Errorf("wrong number of colums, expected %d but got %d", 3, len(tbl.Columns))
	}

	for i, c := range tbl.Columns {
		if c.Name == "" {
			t.Errorf("column #%d table name is empty", i)
		}
		if c.Table == nil {
			t.Errorf("column #%d table nil pointer", i)
		}
	}
}

func TestTable_AddRelationship_oneToOneBidirectional(t *testing.T) {
	user := pqt.NewTable("user").AddColumn(pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey()))
	userDetail := pqt.NewTable("user_detail").AddColumn(pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey()))

	user.AddRelationship(pqt.OneToOneBidirectional(
		userDetail,
		pqt.WithInversedBy("user"),
		pqt.WithMappedBy("details"),
	))

	if len(user.Relationships) != 1 {
		t.Fatalf("user should have 1 relationship, but has %d", len(user.Relationships))
	}

	if user.Relationships[0].MappedBy != "details" {
		t.Errorf("user relationship to user_detail should be mapped by details")
	}

	if user.Relationships[0].MappedTable != userDetail {
		t.Errorf("user relationship to user_detail should be mapped by user_detail table")
	}

	if user.Relationships[0].Type != pqt.RelationshipTypeOneToOneBidirectional {
		t.Errorf("user relationship to user_detail should be one to one bidirectional")
	}

	if len(userDetail.Relationships) != 1 {
		t.Fatalf("user_detail should have 1 relationship, but has %d", len(userDetail.Relationships))
	}

	if userDetail.Relationships[0].InversedBy != "user" {
		t.Errorf("user_detail relationship to user should be mapped by user")
	}

	if userDetail.Relationships[0].InversedTable != user {
		t.Errorf("user_detail relationship to user should be mapped by user table")
	}

	if userDetail.Relationships[0].Type != pqt.RelationshipTypeOneToOneBidirectional {
		t.Errorf("user_detail relationship to user should be %d, but is %d", pqt.RelationshipTypeOneToOneBidirectional, userDetail.Relationships[0].Type)
	}
}

func TestTable_AddRelationship_oneToOneUnidirectional(t *testing.T) {
	user := pqt.NewTable("user").AddColumn(pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey()))
	userDetail := pqt.NewTable("user_detail").AddColumn(pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey()))

	user.AddRelationship(pqt.OneToOneUnidirectional(
		userDetail,
		pqt.WithInversedBy("user"),
		pqt.WithMappedBy("details"),
	))

	if len(user.Relationships) != 0 {
		t.Fatalf("user should have 0 relationship, but has %d", len(user.Relationships))
	}

	if len(userDetail.Relationships) != 1 {
		t.Fatalf("user_detail should have 1 relationship, but has %d", len(userDetail.Relationships))
	}

	if userDetail.Relationships[0].InversedBy != "user" {
		t.Errorf("user_detail relationship to user should be mapped by user")
	}

	if userDetail.Relationships[0].InversedTable != user {
		t.Errorf("user_detail relationship to user should be mapped by user table")
	}

	if userDetail.Relationships[0].Type != pqt.RelationshipTypeOneToOneUnidirectional {
		t.Errorf("user_detail relationship to user should be %d, but is %d", pqt.RelationshipTypeOneToOneUnidirectional, userDetail.Relationships[0].Type)
	}
}

func TestTable_AddRelationship_oneToOneSelfReferencing(t *testing.T) {
	user := pqt.NewTable("user").AddColumn(pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey()))

	user.AddRelationship(pqt.OneToOneSelfReferencing(
		pqt.WithInversedBy("child"),
		pqt.WithMappedBy("parent"),
	))

	if len(user.Relationships) != 2 {
		t.Fatalf("user should have 2 relationship, but has %d", len(user.Relationships))
	}

	if user.Relationships[0].MappedBy != "parent" {
		t.Errorf("user relationship to user should be mapped by parent")
	}

	if user.Relationships[0].MappedTable != user {
		t.Errorf("user relationship to user should be mapped by user table")
	}

	if user.Relationships[0].Type != pqt.RelationshipTypeOneToOneSelfReferencing {
		t.Errorf("user relationship to user should be %d, but is %d", pqt.RelationshipTypeOneToOneSelfReferencing, user.Relationships[0].Type)
	}

	if user.Relationships[1].InversedBy != "child" {
		t.Errorf("user relationship to user should be mapped by user")
	}

	if user.Relationships[1].InversedTable != user {
		t.Errorf("user relationship to user should be mapped by user table")
	}

	if user.Relationships[1].Type != pqt.RelationshipTypeOneToOneSelfReferencing {
		t.Errorf("user relationship to user should be %d, but is %d", pqt.RelationshipTypeOneToOneSelfReferencing, user.Relationships[1].Type)
	}
}

func TestTable_AddRelationship_oneToMany(t *testing.T) {
	user := pqt.NewTable("user").AddColumn(pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey()))
	comment := pqt.NewTable("comment").AddColumn(pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey()))

	user.AddRelationship(pqt.OneToMany(
		comment,
		pqt.WithInversedBy("author"),
		pqt.WithMappedBy("comments"),
	))

	if len(user.Relationships) != 1 {
		t.Fatalf("user should have 1 relationship, but has %d", len(user.Relationships))
	}

	if user.Relationships[0].MappedBy != "comments" {
		t.Errorf("user relationship to comment should be mapped by comments")
	}

	if user.Relationships[0].MappedTable != comment {
		t.Errorf("user relationship to comment should be mapped by comment table")
	}

	if user.Relationships[0].Type != pqt.RelationshipTypeOneToMany {
		t.Errorf("user relationship to comment should be one to many")
	}

	if len(comment.Relationships) != 1 {
		t.Fatalf("comment should have 1 relationship, but has %d", len(comment.Relationships))
	}

	if comment.Relationships[0].InversedBy != "author" {
		t.Errorf("comment relationship to user should be mapped by author")
	}

	if comment.Relationships[0].InversedTable != user {
		t.Errorf("comment relationship to user should be mapped by user table")
	}

	if comment.Relationships[0].Type != pqt.RelationshipTypeOneToMany {
		t.Errorf("comment relationship to user should be %d, but is %d", pqt.RelationshipTypeOneToMany, comment.Relationships[0].Type)
	}
}

func TestTable_AddRelationship_manyToMany(t *testing.T) {
	user := pqt.NewTable("user").AddColumn(pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey()))
	group := pqt.NewTable("group").AddColumn(pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey()))
	userGroups := pqt.NewTable("user_groups")
	user.AddRelationship(pqt.ManyToMany(
		group,
		userGroups,
		pqt.WithInversedBy("users"),
		pqt.WithMappedBy("groups"),
	))

	if len(user.Relationships) != 1 {
		t.Fatalf("user should have 1 relationship, but has %d", len(user.Relationships))
	}

	if user.Relationships[0].MappedBy != "groups" {
		t.Errorf("user relationship to group should be mapped by groups")
	}

	if user.Relationships[0].MappedTable != group {
		t.Errorf("user relationship to group should be mapped by group table")
	}

	if user.Relationships[0].Type != pqt.RelationshipTypeManyToMany {
		t.Errorf("user relationship to group should be many to many")
	}

	if len(group.Relationships) != 1 {
		t.Fatalf("group should have 1 relationship, but has %d", len(group.Relationships))
	}

	if group.Relationships[0].InversedBy != "users" {
		t.Errorf("group relationship to user should be mapped by users")
	}

	if group.Relationships[0].InversedTable != user {
		t.Errorf("group relationship to user should be mapped by user table")
	}

	if group.Relationships[0].Type != pqt.RelationshipTypeManyToMany {
		t.Errorf("group relationship to user should be %d, but is %d", pqt.RelationshipTypeManyToMany, group.Relationships[0].Type)
	}
}

func TestTable_AddRelationship_manyToManySelfReferencing(t *testing.T) {
	friendship := pqt.NewTable("friendship")
	user := pqt.NewTable("user").
		AddColumn(pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey())).
		AddRelationship(pqt.ManyToManySelfReferencing(
		friendship,
		pqt.WithInversedBy("friends_with_me"),
		pqt.WithMappedBy("my_friends"),
	))

	if len(user.Relationships) != 2 {
		t.Fatalf("user should have 2 relationships, but has %d", len(user.Relationships))
	}

	if user.Relationships[0].MappedBy != "my_friends" {
		t.Errorf("user relationship to user should be mapped by my_friends")
	}

	if user.Relationships[0].MappedTable != user {
		t.Errorf("user relationship to group should be mapped by group table")
	}

	if user.Relationships[0].Type != pqt.RelationshipTypeManyToManySelfReferencing {
		t.Errorf("user relationship to group should be many to many")
	}

	if user.Relationships[1].InversedBy != "friends_with_me" {
		t.Errorf("user relationship to user should be mapped by friends_with_me")
	}

	if user.Relationships[1].InversedTable != user {
		t.Errorf("user relationship to user should be mapped by user table")
	}

	if user.Relationships[1].Type != pqt.RelationshipTypeManyToManySelfReferencing {
		t.Errorf("user relationship to user should be %d, but is %d", pqt.RelationshipTypeManyToManySelfReferencing, user.Relationships[1].Type)
	}
}
