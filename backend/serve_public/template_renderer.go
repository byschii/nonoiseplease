package servestatic

import (
	"html/template"
	"io/fs"
	"log"

	conf "be/model/config"
	u "be/rest"

	"github.com/pocketbase/pocketbase/daos"
)

type TemplateRenderer struct {
	TemplateName          string
	ParsedTemplate        template.Template
	DataRetrieverWithUser func(*daos.Dao, string) interface{}
	DataRetriever         func(*daos.Dao) interface{}
}

// returns a list of all templates that are used in the app
func getTemplatedPages(fileSystem fs.FS) []*TemplateRenderer {
	var templatedNames = []*TemplateRenderer{}

	meAccount := MeAccountTemplate("me/account.html", fileSystem)
	register := RegisterTemplate("register.html", fileSystem)
	templatedNames = append(
		templatedNames,
		meAccount,
		register,
	)
	return templatedNames
}

// manage template for 'register.html'
func RegisterTemplate(name string, subFs fs.FS) *TemplateRenderer {
	templateToLoad := template.Must(template.ParseFS(subFs, name))
	return &TemplateRenderer{
		TemplateName:   name,
		ParsedTemplate: *templateToLoad,
		DataRetriever: func(dao *daos.Dao) interface{} {

			msg := ""
			if conf.IsRequireMailVerification(dao) {
				msg = "Go check your email, than "
			}

			retrivedData := struct {
				GoToMailMessage string
			}{
				GoToMailMessage: msg,
			}

			return retrivedData
		},
		DataRetrieverWithUser: func(dao *daos.Dao, userId string) interface{} {
			msg := ""
			if conf.IsRequireMailVerification(dao) {
				msg = "Go check your email, than "
			}

			retrivedData := struct {
				GoToMailMessage string
			}{
				GoToMailMessage: msg,
			}

			return retrivedData
		},
	}
}

// manage template for 'me/account.html'
func MeAccountTemplate(name string, subFs fs.FS) *TemplateRenderer {

	templateToLoad := template.Must(template.ParseFS(subFs, name))

	return &TemplateRenderer{
		TemplateName:   name,
		ParsedTemplate: *templateToLoad,
		DataRetrieverWithUser: func(dao *daos.Dao, userId string) interface{} {

			if userId == "" {
				return nil
			}

			// get user email
			email, err := u.GetUserEmailFromId(dao, userId)
			if err != nil {
				log.Printf("error getting user email from id %s", userId)
			}
			details, err := u.GetUserPartFromId(dao, userId)
			if err != nil {
				log.Println("error getting user part from id, ", err.Error())
			}

			retrivedData := struct {
				UserId         string
				Email          string
				Nickname       string
				ExtensionToken string
			}{
				UserId:   userId,
				Email:    email,
				Nickname: details.Nickname,
			}

			return retrivedData
		},
	}
}
