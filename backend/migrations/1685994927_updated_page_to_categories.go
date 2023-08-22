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

		collection, err := dao.FindCollectionByNameOrId("4ynwd1qhh8f42n1")
		if err != nil {
			return err
		}

		// update
		edit_page_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "absoik9e",
			"name": "page_id",
			"type": "relation",
			"required": true,
			"unique": false,
			"options": {
				"collectionId": "5v6iqh41x3rsuym",
				"cascadeDelete": true,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_page_id)
		collection.Schema.AddField(edit_page_id)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("4ynwd1qhh8f42n1")
		if err != nil {
			return err
		}

		// update
		edit_page_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "absoik9e",
			"name": "page_id",
			"type": "relation",
			"required": true,
			"unique": false,
			"options": {
				"collectionId": "5v6iqh41x3rsuym",
				"cascadeDelete": true,
				"minSelect": null,
				"maxSelect": null,
				"displayFields": null
			}
		}`), edit_page_id)
		collection.Schema.AddField(edit_page_id)

		return dao.SaveCollection(collection)
	})
}
