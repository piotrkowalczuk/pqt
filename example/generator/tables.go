package main

import "github.com/piotrkowalczuk/pqt"

func schema(sn string) *pqt.Schema {
	title := pqt.NewColumn("title", pqt.TypeText(), pqt.WithNotNull(), pqt.WithUnique())

	news := pqt.NewTable("news", pqt.WithTableIfNotExists()).
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig(), pqt.WithPrimaryKey())).
		AddColumn(title).
		AddColumn(pqt.NewColumn("lead", pqt.TypeText())).
		AddColumn(pqt.NewColumn("content", pqt.TypeText(), pqt.WithNotNull()))

	comment := pqt.NewTable("comment", pqt.WithTableIfNotExists()).
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig())).
		AddColumn(pqt.NewColumn("content", pqt.TypeText(), pqt.WithNotNull())).
		AddColumn(pqt.NewColumn(
			"news_title",
			pqt.TypeText(),
			pqt.WithNotNull(),
			pqt.WithReference(title),
		))

	category := pqt.NewTable("category", pqt.WithTableIfNotExists()).
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
	timestampable(news)
	timestampable(comment)
	timestampable(category)

	comment.AddRelationship(pqt.ManyToOne(news, pqt.WithBidirectional()), pqt.WithNotNull())

	pqt.ManyToMany(category, news, pqt.WithBidirectional())

	return pqt.NewSchema(sn, pqt.WithSchemaIfNotExists()).
		AddTable(news).
		AddTable(comment).
		AddTable(category)
}

func timestampable(t *pqt.Table) {
	t.AddColumn(pqt.NewColumn("created_at", pqt.TypeTimestampTZ(), pqt.WithNotNull(), pqt.WithDefault("NOW()"))).
		AddColumn(pqt.NewColumn("updated_at", pqt.TypeTimestampTZ(), pqt.WithDefault("NOW()", pqt.EventUpdate)))
}
