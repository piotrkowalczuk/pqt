package main

import "github.com/piotrkowalczuk/pqt"

func schema(sn string) *pqt.Schema {
	news := tableNews()
	comment := tableComment()
	category := tableCategory()

	comment.AddRelationship(pqt.ManyToOne(news, pqt.WithBidirectional()), pqt.WithNotNull())

	pqt.ManyToMany(category, news, pqt.WithBidirectional())

	return pqt.NewSchema(sn).
		AddTable(news).
		AddTable(comment).
		AddTable(category)
}

func tableNews() *pqt.Table {
	t := pqt.NewTable("news", pqt.WithIfNotExists()).
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig(), pqt.WithPrimaryKey())).
		AddColumn(pqt.NewColumn("title", pqt.TypeText(), pqt.WithNotNull())).
		AddColumn(pqt.NewColumn("lead", pqt.TypeText())).
		AddColumn(pqt.NewColumn("content", pqt.TypeText(), pqt.WithNotNull()))

	timestampable(t)

	return t
}

func tableComment() *pqt.Table {
	t := pqt.NewTable("comment", pqt.WithIfNotExists()).
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig())).
		AddColumn(pqt.NewColumn("content", pqt.TypeText(), pqt.WithNotNull()))

	timestampable(t)

	return t
}

func tableCategory() *pqt.Table {
	t := pqt.NewTable("category", pqt.WithIfNotExists()).
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig(), pqt.WithPrimaryKey())).
		AddColumn(pqt.NewColumn("name", pqt.TypeText(), pqt.WithNotNull())).
		AddColumn(pqt.NewColumn("content", pqt.TypeText(), pqt.WithNotNull())).
		AddRelationship(
			pqt.OneToMany(
				pqt.SelfReference(),
				pqt.WithBidirectional(),
				pqt.WithInversedName("child_category"),
				pqt.WithOwnerName("parent_category"),
				pqt.WithColumnName("parent_id"),
			),
		)

	timestampable(t)

	return t
}

func timestampable(t *pqt.Table) {
	t.AddColumn(pqt.NewColumn("created_at", pqt.TypeTimestampTZ(), pqt.WithNotNull(), pqt.WithDefault("NOW()"))).
		AddColumn(pqt.NewColumn("updated_at", pqt.TypeTimestampTZ(), pqt.WithDefault("NOW()", pqt.EventUpdate)))
}
