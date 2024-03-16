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

		json.Unmarshal([]byte(`[
			"CREATE INDEX ` + "`" + `_i55lo3cshpxcrjp_created_idx` + "`" + ` ON ` + "`" + `user_details` + "`" + ` (` + "`" + `created` + "`" + `)",
			"CREATE UNIQUE INDEX ` + "`" + `idx_unique_zspevcal` + "`" + ` ON ` + "`" + `user_details` + "`" + ` (` + "`" + `related_user` + "`" + `)"
		]`), &collection.Indexes)

		// add
		new_extention_token := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "vmelrxxa",
			"name": "extention_token",
			"type": "text",
			"required": true,
			"unique": false,
			"options": {
				"min": 6,
				"max": null,
				"pattern": ""
			}
		}`), new_extention_token)
		collection.Schema.AddField(new_extention_token)

		// update
		edit_related_user := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "zspevcal",
			"name": "related_user",
			"type": "relation",
			"required": true,
			"unique": false,
			"options": {
				"collectionId": "_pb_users_auth_",
				"cascadeDelete": true,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_related_user)
		collection.Schema.AddField(edit_related_user)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("i55lo3cshpxcrjp")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[
			"CREATE INDEX ` + "`" + `_i55lo3cshpxcrjp_created_idx` + "`" + ` ON ` + "`" + `user_details` + "`" + ` (` + "`" + `created` + "`" + `)",
			"CREATE UNIQUE INDEX \"idx_unique_zspevcal\" on \"user_details\" (\"related_user\")"
		]`), &collection.Indexes)

		// remove
		collection.Schema.RemoveField("vmelrxxa")

		// update
		edit_related_user := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "zspevcal",
			"name": "related_user",
			"type": "relation",
			"required": true,
			"unique": true,
			"options": {
				"collectionId": "_pb_users_auth_",
				"cascadeDelete": true,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_related_user)
		collection.Schema.AddField(edit_related_user)

		return dao.SaveCollection(collection)
	})
}
