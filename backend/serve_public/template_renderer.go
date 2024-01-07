package servestatic

import (
	"html/template"
	"io/fs"
	"log"

	controller "be/controllers"
	conf "be/model/config"
)

type TemplateRenderer struct {
	TemplateName          string
	ParsedTemplate        template.Template
	DataRetrieverWithUser func(controller.UserControllerInterface, string) interface{}
	DataRetriever         func(controller.UserControllerInterface) interface{}
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
		DataRetriever: func(uc controller.UserControllerInterface) interface{} {

			msg := ""
			if conf.IsRequireMailVerification(uc.AppDao()) {
				msg = "Go check your email, than "
			}

			retrivedData := struct {
				GoToMailMessage string
			}{
				GoToMailMessage: msg,
			}

			return retrivedData
		},
		DataRetrieverWithUser: func(uc controller.UserControllerInterface, userId string) interface{} {
			msg := ""
			if conf.IsRequireMailVerification(uc.AppDao()) {
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
		DataRetrieverWithUser: func(uc controller.UserControllerInterface, userId string) interface{} {
			log.Println("retrive data for user", userId)
			if userId == "" {
				return nil
			}

			// get user email
			email, err := uc.GetUserEmailFromId(userId)
			if err != nil {
				log.Printf("error getting user email from id %s", userId)
			}
			details, err := uc.GetUserDetails(userId)
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
