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

		collection, err := dao.FindCollectionByNameOrId("55cz3o94i0qsma6")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	}, func(db dbx.Builder) error {
		jsonData := `{
			"id": "55cz3o94i0qsma6",
			"created": "2023-01-18 21:02:34.691Z",
			"updated": "2024-05-01 18:45:38.376Z",
			"name": "user_activity",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "vfoeexxd",
					"name": "activity_type",
					"type": "select",
					"required": true,
					"presentable": false,
					"unique": false,
					"options": {
						"maxSelect": 1,
						"values": [
							"optimization",
							"update_etf"
						]
					}
				},
				{
					"system": false,
					"id": "pdilo56f",
					"name": "related_user",
					"type": "relation",
					"required": true,
					"presentable": false,
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
					"id": "pvlk0ekk",
					"name": "details",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"maxSize": 2000000
					}
				}
			],
			"indexes": [
				"CREATE INDEX ` + "`" + `_55cz3o94i0qsma6_created_idx` + "`" + ` ON ` + "`" + `user_activity` + "`" + ` (` + "`" + `created` + "`" + `)"
			],
			"listRule": "@request.auth.id = related_user.id",
			"viewRule": "@request.auth.id = related_user.id",
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
