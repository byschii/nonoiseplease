package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("ek3rycc8y7z759x")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	}, func(db dbx.Builder) error {
		jsonData := `{
			"id": "ek3rycc8y7z759x",
			"created": "2023-03-17 23:03:48.359Z",
			"updated": "2023-05-15 11:05:19.914Z",
			"name": "comments",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "orzvatcl",
					"name": "text",
					"type": "text",
					"required": true,
					"unique": false,
					"options": {
						"min": null,
						"max": null,
						"pattern": ""
					}
				},
				{
					"system": false,
					"id": "xpwbyxfi",
					"name": "author",
					"type": "relation",
					"required": true,
					"unique": false,
					"options": {
						"collectionId": "_pb_users_auth_",
						"cascadeDelete": false,
						"minSelect": null,
						"maxSelect": 1,
						"displayFields": null
					}
				},
				{
					"system": false,
					"id": "buqtotvw",
					"name": "page",
					"type": "relation",
					"required": true,
					"unique": false,
					"options": {
						"collectionId": "5v6iqh41x3rsuym",
						"cascadeDelete": false,
						"minSelect": null,
						"maxSelect": 1,
						"displayFields": null
					}
				},
				{
					"system": false,
					"id": "hyivtumx",
					"name": "parent_comment",
					"type": "relation",
					"required": false,
					"unique": false,
					"options": {
						"collectionId": "ek3rycc8y7z759x",
						"cascadeDelete": false,
						"minSelect": null,
						"maxSelect": 1,
						"displayFields": null
					}
				}
			],
			"indexes": [
				"CREATE INDEX ` + "`" + `_ek3rycc8y7z759x_created_idx` + "`" + ` ON ` + "`" + `comments` + "`" + ` (` + "`" + `created` + "`" + `)"
			],
			"listRule": null,
			"viewRule": null,
			"createRule": null,
			"updateRule": null,
			"deleteRule": null,
			"options": {}
		}`

		collection := &models.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return daos.New(db).SaveCollection(collection)
	})
}
