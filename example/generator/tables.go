package main

import "github.com/piotrkowalczuk/pqt"

func schema(sn string) *pqt.Schema {
	multiply := &pqt.Function{
		Name: "multiply",
		Type: pqt.TypeIntegerBig(),
		Body: "SELECT x * y",
		Args: []*pqt.FunctionArg{
			{
				Name: "x",
				Type: pqt.TypeIntegerBig(),
			},
			{
				Name: "y",
				Type: pqt.TypeIntegerBig(),
			},
		},
	}
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
		AddColumn(pqt.NewColumn("meta_data", pqt.TypeJSONB())).
		AddUnique(title, lead)

	commentID := pqt.NewColumn("id", pqt.TypeSerialBig())
	comment := pqt.NewTable("comment", pqt.WithTableIfNotExists()).
		AddColumn(commentID).
		AddColumn(pqt.NewColumn("content", pqt.TypeText(), pqt.WithNotNull())).
		AddColumn(pqt.NewColumn(
			"news_title",
			pqt.TypeText(),
			pqt.WithNotNull(),
			pqt.WithReference(title, pqt.WithBidirectional(), pqt.WithOwnerName("comments_by_news_title"), pqt.WithInversedName("news_by_title")),
		)).
		AddColumn(pqt.NewDynamicColumn("right_now", pqt.FunctionNow())).
		AddColumn(pqt.NewDynamicColumn("id_multiply", multiply, commentID, commentID))

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

	complete := pqt.NewTable("complete", pqt.WithTableIfNotExists()).
		AddColumn(pqt.NewColumn("column_jsonb", pqt.TypeJSONB())).
		AddColumn(pqt.NewColumn("column_jsonb_nn", pqt.TypeJSONB(), pqt.WithNotNull())).
		AddColumn(pqt.NewColumn("column_jsonb_nn_d", pqt.TypeJSONB(), pqt.WithNotNull(), pqt.WithDefault(`'{"field": 1}'`))).
		AddColumn(pqt.NewColumn("column_json", pqt.TypeJSON())).
		AddColumn(pqt.NewColumn("column_json_nn", pqt.TypeJSON(), pqt.WithNotNull())).
		AddColumn(pqt.NewColumn("column_json_nn_d", pqt.TypeJSON(), pqt.WithNotNull(), pqt.WithDefault(`'{"field": 1}'`))).
		AddColumn(pqt.NewColumn("column_bool", pqt.TypeBool())).
		AddColumn(pqt.NewColumn("column_bytea", pqt.TypeBytea())).
		AddColumn(pqt.NewColumn("column_character_0", pqt.TypeCharacter(0))).
		AddColumn(pqt.NewColumn("column_character_100", pqt.TypeCharacter(100))).
		AddColumn(pqt.NewColumn("column_decimal", pqt.TypeDecimal(20, 8))).
		AddColumn(pqt.NewColumn("column_double_array_0", pqt.TypeDoubleArray(0))).
		AddColumn(pqt.NewColumn("column_double_array_100", pqt.TypeDoubleArray(100))).
		AddColumn(pqt.NewColumn("column_integer", pqt.TypeInteger())).
		AddColumn(pqt.NewColumn("column_integer_array_0", pqt.TypeIntegerArray(0))).
		AddColumn(pqt.NewColumn("column_integer_array_100", pqt.TypeIntegerArray(100))).
		AddColumn(pqt.NewColumn("column_integer_big", pqt.TypeIntegerBig())).
		AddColumn(pqt.NewColumn("column_integer_big_array_0", pqt.TypeIntegerBigArray(0))).
		AddColumn(pqt.NewColumn("column_integer_big_array_100", pqt.TypeIntegerBigArray(100))).
		AddColumn(pqt.NewColumn("column_integer_small", pqt.TypeIntegerSmall())).
		AddColumn(pqt.NewColumn("column_integer_small_array_0", pqt.TypeIntegerSmallArray(0))).
		AddColumn(pqt.NewColumn("column_integer_small_array_100", pqt.TypeIntegerSmallArray(100))).
		AddColumn(pqt.NewColumn("column_numeric", pqt.TypeNumeric(20, 8))).
		AddColumn(pqt.NewColumn("column_real", pqt.TypeReal())).
		AddColumn(pqt.NewColumn("column_serial", pqt.TypeSerial())).
		AddColumn(pqt.NewColumn("column_serial_big", pqt.TypeSerialBig())).
		AddColumn(pqt.NewColumn("column_serial_small", pqt.TypeSerialSmall())).
		AddColumn(pqt.NewColumn("column_text", pqt.TypeText())).
		AddColumn(pqt.NewColumn("column_text_array_0", pqt.TypeTextArray(0))).
		AddColumn(pqt.NewColumn("column_text_array_100", pqt.TypeTextArray(100))).
		AddColumn(pqt.NewColumn("column_timestamp", pqt.TypeTimestamp())).
		AddColumn(pqt.NewColumn("column_timestamptz", pqt.TypeTimestampTZ())).
		AddColumn(pqt.NewColumn("column_uuid", pqt.TypeUUID()))

	return pqt.NewSchema(sn, pqt.WithSchemaIfNotExists()).
		AddTable(category).
		AddTable(pkg).
		AddTable(news).
		AddTable(comment).
		AddTable(complete).
		AddFunction(multiply)
}

func timestampable(t *pqt.Table) {
	t.AddColumn(pqt.NewColumn("created_at", pqt.TypeTimestampTZ(), pqt.WithNotNull(), pqt.WithDefault("NOW()"))).
		AddColumn(pqt.NewColumn("updated_at", pqt.TypeTimestampTZ(), pqt.WithDefault("NOW()", pqt.EventUpdate)))
}
