package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models/schema"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("i0nm1zkcnlq7gw2")
		if err != nil {
			return err
		}

		// remove
		collection.Schema.RemoveField("jil7qcbs")

		// add
		new_field := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "lgnuwy2x",
			"name": "field",
			"type": "select",
			"required": true,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSelect": 1,
				"values": [
					"scrape",
					"extention",
					"text_field"
				]
			}
		}`), new_field); err != nil {
			return err
		}
		collection.Schema.AddField(new_field)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("i0nm1zkcnlq7gw2")
		if err != nil {
			return err
		}

		// add
		del_origin := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "jil7qcbs",
			"name": "origin",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), del_origin); err != nil {
			return err
		}
		collection.Schema.AddField(del_origin)

		// remove
		collection.Schema.RemoveField("lgnuwy2x")

		return dao.SaveCollection(collection)
	})
}
