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

		// add
		new_origin := &schema.SchemaField{}
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
		}`), new_origin); err != nil {
			return err
		}
		collection.Schema.AddField(new_origin)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("i0nm1zkcnlq7gw2")
		if err != nil {
			return err
		}

		// remove
		collection.Schema.RemoveField("jil7qcbs")

		return dao.SaveCollection(collection)
	})
}
