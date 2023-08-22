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
		jsonData := `[
			{
				"id": "i55lo3cshpxcrjp",
				"created": "2022-12-19 21:24:06.804Z",
				"updated": "2023-05-15 11:05:19.903Z",
				"name": "user_details",
				"type": "base",
				"system": false,
				"schema": [
					{
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
					},
					{
						"system": false,
						"id": "sayjsjcf",
						"name": "nickname",
						"type": "text",
						"required": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					}
				],
				"indexes": [
					"CREATE INDEX ` + "`" + `_i55lo3cshpxcrjp_created_idx` + "`" + ` ON ` + "`" + `user_details` + "`" + ` (` + "`" + `created` + "`" + `)",
					"CREATE UNIQUE INDEX \"idx_unique_zspevcal\" on \"user_details\" (\"related_user\")"
				],
				"listRule": "@request.auth.id = related_user.id",
				"viewRule": "@request.auth.id = related_user.id",
				"createRule": null,
				"updateRule": "@request.auth.id = related_user.id",
				"deleteRule": null,
				"options": {}
			},
			{
				"id": "55cz3o94i0qsma6",
				"created": "2023-01-18 21:02:34.691Z",
				"updated": "2023-05-15 11:05:19.905Z",
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
						"unique": false,
						"options": {}
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
			},
			{
				"id": "5v6iqh41x3rsuym",
				"created": "2023-03-13 10:49:19.956Z",
				"updated": "2023-05-15 11:05:19.907Z",
				"name": "pages",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "jwertuox",
						"name": "owner",
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
						"id": "6ntrcsnn",
						"name": "link",
						"type": "url",
						"required": true,
						"unique": false,
						"options": {
							"exceptDomains": null,
							"onlyDomains": null
						}
					},
					{
						"system": false,
						"id": "uavmdrbl",
						"name": "page_title",
						"type": "text",
						"required": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "z45bhnd8",
						"name": "fts_ref",
						"type": "text",
						"required": true,
						"unique": true,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "gsiv3tnj",
						"name": "votes",
						"type": "number",
						"required": true,
						"unique": false,
						"options": {
							"min": 1,
							"max": null
						}
					},
					{
						"system": false,
						"id": "8ha17oaa",
						"name": "with_proxy",
						"type": "bool",
						"required": false,
						"unique": false,
						"options": {}
					},
					{
						"system": false,
						"id": "iku3ksdf",
						"name": "origin",
						"type": "select",
						"required": true,
						"unique": false,
						"options": {
							"maxSelect": 1,
							"values": [
								"scrape",
								"extention",
								"text_field"
							]
						}
					}
				],
				"indexes": [
					"CREATE INDEX ` + "`" + `_5v6iqh41x3rsuym_created_idx` + "`" + ` ON \"pages\" (` + "`" + `created` + "`" + `)",
					"CREATE UNIQUE INDEX \"idx_unique_z45bhnd8\" on \"pages\" (\"fts_ref\")"
				],
				"listRule": "@request.auth.id = owner.id",
				"viewRule": "@request.auth.id = owner.id",
				"createRule": "@request.auth.id = owner.id",
				"updateRule": "@request.auth.id = owner.id",
				"deleteRule": "@request.auth.id = owner.id",
				"options": {}
			},
			{
				"id": "ihrmv170av8ev9q",
				"created": "2023-03-14 09:42:30.901Z",
				"updated": "2023-05-15 11:05:19.911Z",
				"name": "categories",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "peaejaw7",
						"name": "name",
						"type": "text",
						"required": true,
						"unique": true,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "1n2zhnxq",
						"name": "color",
						"type": "text",
						"required": true,
						"unique": false,
						"options": {
							"min": 3,
							"max": 6,
							"pattern": ""
						}
					}
				],
				"indexes": [
					"CREATE INDEX ` + "`" + `_ihrmv170av8ev9q_created_idx` + "`" + ` ON ` + "`" + `categories` + "`" + ` (` + "`" + `created` + "`" + `)",
					"CREATE UNIQUE INDEX \"idx_unique_peaejaw7\" on \"categories\" (\"name\")"
				],
				"listRule": null,
				"viewRule": null,
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {}
			},
			{
				"id": "4ynwd1qhh8f42n1",
				"created": "2023-03-14 14:48:17.955Z",
				"updated": "2023-05-15 11:05:19.912Z",
				"name": "page_to_categories",
				"type": "base",
				"system": false,
				"schema": [
					{
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
					},
					{
						"system": false,
						"id": "3o5geiyy",
						"name": "category_id",
						"type": "relation",
						"required": true,
						"unique": false,
						"options": {
							"collectionId": "ihrmv170av8ev9q",
							"cascadeDelete": false,
							"minSelect": null,
							"maxSelect": 1,
							"displayFields": null
						}
					}
				],
				"indexes": [
					"CREATE INDEX ` + "`" + `_4ynwd1qhh8f42n1_created_idx` + "`" + ` ON ` + "`" + `page_to_categories` + "`" + ` (` + "`" + `created` + "`" + `)"
				],
				"listRule": null,
				"viewRule": null,
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {}
			},
			{
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
			},
			{
				"id": "dj5nwif6tpvibcb",
				"created": "2023-03-18 13:30:22.996Z",
				"updated": "2023-05-15 11:05:19.915Z",
				"name": "proxy_connections",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "dtps1i9a",
						"name": "enabled",
						"type": "bool",
						"required": false,
						"unique": false,
						"options": {}
					},
					{
						"system": false,
						"id": "o1s4bpon",
						"name": "address",
						"type": "text",
						"required": true,
						"unique": false,
						"options": {
							"min": 7,
							"max": 16,
							"pattern": "((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)"
						}
					},
					{
						"system": false,
						"id": "p6kzkz9q",
						"name": "port",
						"type": "number",
						"required": true,
						"unique": false,
						"options": {
							"min": 20,
							"max": 20000
						}
					}
				],
				"indexes": [
					"CREATE INDEX ` + "`" + `_dj5nwif6tpvibcb_created_idx` + "`" + ` ON ` + "`" + `proxy_connections` + "`" + ` (` + "`" + `created` + "`" + `)"
				],
				"listRule": null,
				"viewRule": null,
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {}
			},
			{
				"id": "rxiblprp8cb8e9q",
				"created": "2023-03-18 14:03:30.379Z",
				"updated": "2023-05-15 11:05:19.917Z",
				"name": "config",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "hdaii9oc",
						"name": "key",
						"type": "text",
						"required": true,
						"unique": true,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "prhbxww3",
						"name": "text_value",
						"type": "text",
						"required": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "uqpkjmjo",
						"name": "float_value",
						"type": "number",
						"required": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null
						}
					},
					{
						"system": false,
						"id": "rkap5ncd",
						"name": "note",
						"type": "text",
						"required": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "3yydiron",
						"name": "boolean_value",
						"type": "bool",
						"required": false,
						"unique": false,
						"options": {}
					}
				],
				"indexes": [
					"CREATE INDEX ` + "`" + `_rxiblprp8cb8e9q_created_idx` + "`" + ` ON ` + "`" + `config` + "`" + ` (` + "`" + `created` + "`" + `)",
					"CREATE UNIQUE INDEX \"idx_unique_hdaii9oc\" on \"config\" (\"key\")"
				],
				"listRule": null,
				"viewRule": null,
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {}
			},
			{
				"id": "_pb_users_auth_",
				"created": "2023-04-23 17:35:47.008Z",
				"updated": "2023-05-15 11:05:19.898Z",
				"name": "users",
				"type": "auth",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "users_name",
						"name": "name",
						"type": "text",
						"required": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "users_avatar",
						"name": "avatar",
						"type": "file",
						"required": false,
						"unique": false,
						"options": {
							"maxSelect": 1,
							"maxSize": 5242880,
							"mimeTypes": [
								"image/jpg",
								"image/jpeg",
								"image/png",
								"image/svg+xml",
								"image/gif"
							],
							"thumbs": null,
							"protected": false
						}
					}
				],
				"indexes": [
					"CREATE INDEX ` + "`" + `__pb_users_auth__created_idx` + "`" + ` ON ` + "`" + `users` + "`" + ` (` + "`" + `created` + "`" + `)"
				],
				"listRule": "id = @request.auth.id",
				"viewRule": "id = @request.auth.id",
				"createRule": "",
				"updateRule": "id = @request.auth.id",
				"deleteRule": "id = @request.auth.id",
				"options": {
					"allowEmailAuth": true,
					"allowOAuth2Auth": false,
					"allowUsernameAuth": false,
					"exceptEmailDomains": null,
					"manageRule": null,
					"minPasswordLength": 8,
					"onlyEmailDomains": null,
					"requireEmail": true
				}
			}
		]`

		collections := []*models.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collections); err != nil {
			return err
		}

		return daos.New(db).ImportCollections(collections, true, nil)
	}, func(db dbx.Builder) error {
		return nil
	})
}
