package main

import "github.com/piotrkowalczuk/pqt"

func schema(sn string) *pqt.Schema {
	title := pqt.NewColumn("title", pqt.TypeText(), pqt.WithNotNull(), pqt.WithUnique())
	lead := pqt.NewColumn("lead", pqt.TypeText())

	news := pqt.NewTable("news", pqt.WithTableIfNotExists()).
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig(), pqt.WithPrimaryKey())).
		AddColumn(title).
		AddColumn(lead).
		AddColumn(pqt.NewColumn("continue", pqt.TypeBool(), pqt.WithNotNull(), pqt.WithDefault("false"))).
		AddColumn(pqt.NewColumn("content", pqt.TypeText(), pqt.WithNotNull())).
		AddColumn(pqt.NewColumn("score", pqt.TypeNumeric(20, 8), pqt.WithNotNull(), pqt.WithDefault("0"))).
		AddColumn(pqt.NewColumn("views_distribution", pqt.TypeDoubleArray(168))).
		AddUnique(title, lead)

	comment := pqt.NewTable("comment", pqt.WithTableIfNotExists()).
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig())).
		AddColumn(pqt.NewColumn("content", pqt.TypeText(), pqt.WithNotNull())).
		AddColumn(pqt.NewColumn(
			"news_title",
			pqt.TypeText(),
			pqt.WithNotNull(),
			pqt.WithReference(title, pqt.WithBidirectional(), pqt.WithOwnerName("comments_by_news_title"), pqt.WithInversedName("news_by_title")),
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

	pkg := pqt.NewTable("package", pqt.WithTableIfNotExists()).
		AddColumn(pqt.NewColumn("id", pqt.TypeSerialBig(), pqt.WithPrimaryKey())).
		AddColumn(pqt.NewColumn("break", pqt.TypeText())).
		AddRelationship(pqt.ManyToOne(
			category,
			pqt.WithBidirectional(),
		))

	timestampable(news)
	timestampable(comment)
	timestampable(category)
	timestampable(pkg)
	comment.AddRelationship(pqt.ManyToOne(news, pqt.WithBidirectional(), pqt.WithInversedName("news_by_id")), pqt.WithNotNull())

	pqt.ManyToMany(category, news, pqt.WithBidirectional())

	return pqt.NewSchema(sn, pqt.WithSchemaIfNotExists()).
		AddTable(category).
		AddTable(pkg).
		AddTable(news).
		AddTable(comment)
}

func timestampable(t *pqt.Table) {
	t.AddColumn(pqt.NewColumn("created_at", pqt.TypeTimestampTZ(), pqt.WithNotNull(), pqt.WithDefault("NOW()"))).
		AddColumn(pqt.NewColumn("updated_at", pqt.TypeTimestampTZ(), pqt.WithDefault("NOW()", pqt.EventUpdate)))
}
