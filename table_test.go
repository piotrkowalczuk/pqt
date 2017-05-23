package pqt_test

import (
	"reflect"
	"testing"

	"github.com/piotrkowalczuk/pqt"
)

func TestNewTable(t *testing.T) {
	tbl := pqt.NewTable("test", pqt.WithTableIfNotExists(), pqt.WithTableSpace("table_space"), pqt.WithTemporary())

	if !tbl.IfNotExists {
		t.Errorf("table should have field if not exists set to true")
	}

	if !tbl.Temporary {
		t.Errorf("table should have field temporary set to true")
	}

	if tbl.TableSpace != "table_space" {
		t.Errorf("table should have field table space set to table_space")
	}
}

func TestWithTableShortName(t *testing.T) {
	tbl := pqt.NewTable("table", pqt.WithTableShortName("tbl"))
	if tbl.ShortName != "tbl" {
		t.Errorf("wrong table short name: %s", tbl.ShortName)
	}
}

func TestTable_SetIfNotExists(t *testing.T) {
	tbl := pqt.NewTable("table").SetIfNotExists(true)
	if !tbl.IfNotExists {
		t.Error("if not exists expected to be true")
	}
}

func TestTable_SetSchema(t *testing.T) {
	sch1 := pqt.NewSchema("schema1")
	sch2 := pqt.NewSchema("schema2")
	tbl := pqt.NewTable("table").SetSchema(sch1)
	if !reflect.DeepEqual(tbl.Schema, sch1) {
		t.Error("wrong schema")
	}
	tbl.SetSchema(sch2)
	if !reflect.DeepEqual(tbl.Schema, sch2) {
		t.Error("wrong schema")
	}
}

func TestTable_AddColumn(t *testing.T) {
	c0 := pqt.NewColumn("c0", pqt.TypeSerialBig(), pqt.WithPrimaryKey())
	c1 := &pqt.Column{Name: "c1"}
	c2 := &pqt.Column{Name: "c2"}
	c3 := &pqt.Column{Name: "c3"}

	tbl := pqt.NewTable("test").
		AddColumn(c0).
		AddColumn(c1).
		AddColumn(c2).
		AddColumn(c3).
		AddColumn(pqt.NewColumn("c4", pqt.TypeIntegerBig(), pqt.WithReference(c0))).
		AddRelationship(pqt.ManyToOne(pqt.SelfReference()))

	if len(tbl.Columns) != 6 {
		t.Errorf("wrong number of colums, expected %d but got %d", 6, len(tbl.Columns))
	}

	if len(tbl.OwnedRelationships) != 2 {
		// Reference is not a relationship
		t.Errorf("wrong number of owned relationships, expected %d but got %d", 2, len(tbl.OwnedRelationships))
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

func TestTable_AddConstraint(t *testing.T) {
	tbl := pqt.Table{}
	idx := pqt.Constraint{
		Type: pqt.ConstraintTypeIndex,
	}
	got := tbl.AddConstraint(&idx).Constraints.CountOf(pqt.ConstraintTypeIndex)
	if got != 1 {
		t.Errorf("wrong number of indexes: %d", got)
	}
}

func TestTable_AddRelationship_oneToOneBidirectional(t *testing.T) {
	user := pqt.NewTable("user").AddColumn(pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey()))
	userDetail := pqt.NewTable("user_detail").AddColumn(pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey()))

	user.AddRelationship(pqt.OneToOne(
		userDetail,
		pqt.WithInversedName("details"),
		pqt.WithOwnerName("user"),
		pqt.WithBidirectional(),
	))

	if len(user.OwnedRelationships) != 1 {
		t.Fatalf("user should have 1 relationship, but has %d", len(user.OwnedRelationships))
	}

	if user.OwnedRelationships[0].OwnerName != "user" {
		t.Errorf("user relationship to user_detail should be mapped by user, but is %s", user.OwnedRelationships[0].OwnerName)
	}

	if user.OwnedRelationships[0].OwnerTable != user {
		t.Errorf("user relationship to user_detail should be mapped by user table, but is %v", user.OwnedRelationships[0].OwnerTable)
	}

	if user.OwnedRelationships[0].Type != pqt.RelationshipTypeOneToOne {
		t.Error("user relationship to user_detail should be one to one bidirectional")
	}

	if len(userDetail.InversedRelationships) != 1 {
		t.Fatalf("user_detail should have 1 relationship, but has %d", len(userDetail.InversedRelationships))
	}

	if userDetail.InversedRelationships[0].InversedName != "details" {
		t.Error("user_detail relationship to user should be mapped by user")
	}

	if userDetail.InversedRelationships[0].InversedTable != userDetail {
		t.Error("user_detail relationship to user should be mapped by user_detail table")
	}

	if userDetail.InversedRelationships[0].Type != pqt.RelationshipTypeOneToOne {
		t.Errorf("user_detail relationship to user should be %d, but is %d", pqt.RelationshipTypeOneToOne, userDetail.InversedRelationships[0].Type)
	}
}

func TestTable_AddRelationship_oneToOneUnidirectional(t *testing.T) {
	user := pqt.NewTable("user").AddColumn(pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey()))
	userDetail := pqt.NewTable("user_detail").AddColumn(pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey())).
		AddRelationship(pqt.OneToOne(
			user,
			pqt.WithInversedName("user"),
			pqt.WithOwnerName("details"),
		))

	if len(user.InversedRelationships) != 0 {
		t.Fatalf("user should have 0 relationship, but has %d", len(user.InversedRelationships))
	}

	if len(userDetail.OwnedRelationships) != 1 {
		t.Fatalf("user_detail should have 1 relationship, but has %d", len(userDetail.OwnedRelationships))
	}

	if userDetail.OwnedRelationships[0].InversedName != "user" {
		t.Errorf("user_detail relationship to user should be mapped by user")
	}

	if userDetail.OwnedRelationships[0].InversedTable != user {
		t.Errorf("user_detail relationship to user should be mapped by user table")
	}

	if userDetail.OwnedRelationships[0].Type != pqt.RelationshipTypeOneToOne {
		t.Errorf("user_detail relationship to user should be %d, but is %d", pqt.RelationshipTypeOneToOne, userDetail.OwnedRelationships[0].Type)
	}
}

func TestTable_AddRelationship_oneToOneSelfReferencing(t *testing.T) {
	user := pqt.NewTable("user").AddColumn(pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey()))

	user.AddRelationship(pqt.OneToOne(
		pqt.SelfReference(),
		pqt.WithInversedName("child"),
		pqt.WithOwnerName("parent"),
	))

	if len(user.OwnedRelationships) != 1 {
		t.Fatalf("user should have 1 owned relationship, but has %d", len(user.OwnedRelationships))
	}

	if user.OwnedRelationships[0].OwnerName != "parent" {
		t.Errorf("user relationship to user should be mapped by parent")
	}

	if user.OwnedRelationships[0].OwnerTable != user {
		t.Errorf("user relationship to user should be mapped by user table")
	}

	if user.OwnedRelationships[0].Type != pqt.RelationshipTypeOneToOne {
		t.Errorf("user relationship to user should be %d, but is %d", pqt.RelationshipTypeOneToOne, user.OwnedRelationships[0].Type)
	}

	if len(user.InversedRelationships) != 0 {
		t.Fatalf("user should have 0 inversed relationship, but has %d", len(user.InversedRelationships))
	}
}

func TestTable_AddRelationship_oneToMany(t *testing.T) {
	user := pqt.NewTable("user").AddColumn(pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey()))
	comment := pqt.NewTable("comment").AddColumn(pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey()))

	user.AddRelationship(pqt.OneToMany(
		comment,
		pqt.WithBidirectional(),
		pqt.WithInversedName("author"),
		pqt.WithOwnerName("comments"),
	))

	if len(user.InversedRelationships) != 1 {
		t.Fatalf("user should have 1 inversed relationship, but has %d", len(user.InversedRelationships))
	}

	if user.InversedRelationships[0].OwnerName != "comments" {
		t.Errorf("user inversed relationship to comment should be mapped by comments")
	}

	if user.InversedRelationships[0].OwnerTable != comment {
		t.Errorf("user inversed relationship to comment should be mapped by comment table")
	}

	if user.InversedRelationships[0].Type != pqt.RelationshipTypeOneToMany {
		t.Errorf("user inversed relationship to comment should be one to many")
	}

	if len(comment.OwnedRelationships) != 1 {
		t.Fatalf("comment should have 1 owned relationship, but has %d", len(comment.OwnedRelationships))
	}

	if comment.OwnedRelationships[0].InversedName != "author" {
		t.Errorf("comment relationship to user should be mapped by author")
	}

	if comment.OwnedRelationships[0].InversedTable != user {
		t.Errorf("comment relationship to user should be mapped by user table")
	}

	if comment.OwnedRelationships[0].Type != pqt.RelationshipTypeOneToMany {
		t.Errorf("comment relationship to user should be %d, but is %d", pqt.RelationshipTypeOneToMany, comment.OwnedRelationships[0].Type)
	}
}

func TestTable_AddRelationship_manyToMany(t *testing.T) {
	user := pqt.NewTable("user").AddColumn(pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey()))
	group := pqt.NewTable("group").AddColumn(pqt.NewColumn("id", pqt.TypeSerial(), pqt.WithPrimaryKey()))
	userGroups := pqt.NewTable("user_groups")
	user.AddRelationship(pqt.ManyToMany(
		group,
		userGroups,
		pqt.WithInversedName("users"),
		pqt.WithOwnerName("groups"),
	))

	if len(user.OwnedRelationships) != 1 {
		t.Fatalf("user should have 1 relationship, but has %d", len(user.OwnedRelationships))
	}

	if user.OwnedRelationships[0].OwnerName != "groups" {
		t.Errorf("user relationship to group should be mapped by groups")
	}

	if user.OwnedRelationships[0].OwnerTable != group {
		t.Errorf("user relationship to group should be mapped by group table")
	}

	if user.OwnedRelationships[0].Type != pqt.RelationshipTypeManyToMany {
		t.Errorf("user relationship to group should be many to many")
	}

	if len(group.InversedRelationships) != 0 {
		t.Fatalf("group should have 0 relationship, but has %d", len(group.InversedRelationships))
	}
}

func TestTable_FullName(t *testing.T) {
	tbl := pqt.NewTable("table")
	if tbl.FullName() != "table" {
		t.Errorf("wrong full name: %s", tbl.FullName())
	}

	pqt.NewSchema("schema").AddTable(tbl)
	if tbl.FullName() != "schema.table" {
		t.Errorf("wrong full name: %s", tbl.FullName())
	}
}

func TestTable_AddCheck(t *testing.T) {
	a := pqt.NewColumn("a", pqt.TypeInteger())
	b := pqt.NewColumn("b", pqt.TypeInteger())

	tbl := pqt.NewTable("table").
		AddColumn(a).
		AddColumn(b).
		AddCheck("a > b", a, b)

	got := tbl.Constraints.CountOf(pqt.ConstraintTypeCheck)
	if got != 1 {
		t.Errorf("wrong number of check constraints: %d", got)
	}
}

func TestTable_AddUnique(t *testing.T) {
	a := pqt.NewColumn("a", pqt.TypeInteger())
	b := pqt.NewColumn("b", pqt.TypeInteger())

	tbl := pqt.NewTable("table").
		AddColumn(a).
		AddColumn(b).
		AddUnique(a, b)

	got := tbl.Constraints.CountOf(pqt.ConstraintTypeUnique)
	if got != 1 {
		t.Errorf("wrong number of unique constraints: %d", got)
	}
}

func TestTable_AddIndex(t *testing.T) {
	a := pqt.NewColumn("a", pqt.TypeInteger())
	b := pqt.NewColumn("b", pqt.TypeInteger())

	tbl := pqt.NewTable("table").
		AddColumn(a).
		AddColumn(b).
		AddIndex(a, b)

	got := tbl.Constraints.CountOf(pqt.ConstraintTypeIndex)
	if got != 1 {
		t.Errorf("wrong number of index constraints: %d", got)
	}
}
