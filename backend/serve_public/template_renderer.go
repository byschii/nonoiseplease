package servestatic

import (
	"github.com/rs/zerolog/log"

	controller "be/controllers"
	"be/pkg/users"
)

type TemplateRenderer struct {
	DataRetrieverWithUser func(controller.UserControllerInterface, string, controller.AppStateControllerInterface) interface{}
	DataRetriever         func(controller.UserControllerInterface, controller.AppStateControllerInterface) interface{}
}

// returns a list of all templates that are used in the app
func getTemplatedPages() map[string]*TemplateRenderer {
	var templatedNames = make(map[string]*TemplateRenderer)

	templatedNames["register.html"] = RegisterTemplate()
	templatedNames["me/account.html"] = MeAccountTemplate()

	// template.ParseFiles("pb_public/register.html", "views_template/prova_template.html")

	return templatedNames
}

// manage template for 'register.html'
func RegisterTemplate() *TemplateRenderer {

	return &TemplateRenderer{
		DataRetriever: func(uc controller.UserControllerInterface, confController controller.AppStateControllerInterface) interface{} {

			msg := ""
			if confController.IsRequireMailVerification() {
				msg = "Go check your email, than "
			}

			retrivedData := struct {
				GoToMailMessage string
			}{
				GoToMailMessage: msg,
			}

			return retrivedData
		},
		DataRetrieverWithUser: func(uc controller.UserControllerInterface, userId string, confController controller.AppStateControllerInterface) interface{} {
			msg := ""
			if confController.IsRequireMailVerification() {
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
func MeAccountTemplate() *TemplateRenderer {

	return &TemplateRenderer{
		DataRetrieverWithUser: func(uc controller.UserControllerInterface, userId string, confController controller.AppStateControllerInterface) interface{} {
			log.Debug().Msgf("retrive data for user %s", userId)
			if userId == "" {
				return nil
			}

			// get user email
			email, err := users.EmailFromId(uc.AppDao(), userId)
			if err != nil {
				log.Error().Msgf("error getting user email from id %s", userId)
			}
			details, err := uc.GetUserDetails(userId)
			if err != nil {
				log.Error().Msgf("error getting user part from id %v ", err.Error())
			}
			log.Debug().Msgf(details.Nickname)

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
