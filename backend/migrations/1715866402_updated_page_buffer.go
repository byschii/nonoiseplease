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

		// update
		edit_origin := &schema.SchemaField{}
		if err := json.Unmarshal([]byte(`{
			"system": false,
			"id": "lgnuwy2x",
			"name": "origin",
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
		}`), edit_origin); err != nil {
			return err
		}
		collection.Schema.AddField(edit_origin)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("i0nm1zkcnlq7gw2")
		if err != nil {
			return err
		}

		// update
		edit_origin := &schema.SchemaField{}
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
		}`), edit_origin); err != nil {
			return err
		}
		collection.Schema.AddField(edit_origin)

		return dao.SaveCollection(collection)
	})
}
