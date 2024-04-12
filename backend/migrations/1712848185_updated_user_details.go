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

		collection, err := dao.FindCollectionByNameOrId("i55lo3cshpxcrjp")
		if err != nil {
			return err
		}

		// remove
		collection.Schema.RemoveField("vmelrxxa")

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("i55lo3cshpxcrjp")
		if err != nil {
			return err
		}

		// add
		del_extension_token := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "vmelrxxa",
			"name": "extension_token",
			"type": "text",
			"required": true,
			"unique": false,
			"options": {
				"min": 6,
				"max": null,
				"pattern": ""
			}
		}`), del_extension_token)
		collection.Schema.AddField(del_extension_token)

		return dao.SaveCollection(collection)
	})
}
