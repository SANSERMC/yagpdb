package moderation

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jonas747/discordgo"
	"github.com/jonas747/yagpdb/common"
	"github.com/jonas747/yagpdb/web"
	"goji.io/pat"
	"golang.org/x/net/context"
	"html/template"
	"net/http"
)

func (p *Plugin) InitWeb() {
	web.Templates = template.Must(web.Templates.ParseFiles("templates/plugins/moderation.html"))

	web.CPMux.HandleC(pat.Get("/moderation"), web.RequireGuildChannelsMiddleware(web.RenderHandler(HandleModeration, "cp_moderation")))
	web.CPMux.HandleC(pat.Get("/moderation/"), web.RequireGuildChannelsMiddleware(web.RenderHandler(HandleModeration, "cp_moderation")))
	web.CPMux.HandleC(pat.Post("/moderation"), web.RequireGuildChannelsMiddleware(web.FormParserMW(web.RenderHandler(HandlePostModeration, "cp_moderation"), Config{})))
	web.CPMux.HandleC(pat.Post("/moderation/"), web.RequireGuildChannelsMiddleware(web.FormParserMW(web.RenderHandler(HandlePostModeration, "cp_moderation"), Config{})))
}

// The moderation page itself
func HandleModeration(ctx context.Context, w http.ResponseWriter, r *http.Request) interface{} {
	client, activeGuild, templateData := web.GetBaseCPContextData(ctx)
	config, err := GetConfig(client, activeGuild.ID)

	if err != nil {
		templateData.AddAlerts(web.ErrorAlert("Failed retrieving config", err))
		log.WithError(err).WithField("guild", activeGuild.ID).Error("Failed retrieving config")
	}

	templateData["ModConfig"] = config

	return templateData
}

// Update the settings
func HandlePostModeration(ctx context.Context, w http.ResponseWriter, r *http.Request) interface{} {
	client, activeGuild, templateData := web.GetBaseCPContextData(ctx)
	templateData["VisibleURL"] = "/cp/" + activeGuild.ID + "/moderation/"

	newConfig := ctx.Value(web.ContextKeyParsedForm).(*Config)
	templateData["ModConfig"] = newConfig

	ok := ctx.Value(web.ContextKeyFormOk).(bool)
	if !ok {
		return templateData
	}

	err := newConfig.Save(client, activeGuild.ID)
	if web.CheckErr(templateData, err, "Failed saving :(", log.Error) {
		return templateData
	}

	templateData.AddAlerts(web.SucessAlert("Sucessfully saved config! :')"))

	user := ctx.Value(web.ContextKeyUser).(*discordgo.User)
	go common.AddCPLogEntry(user, activeGuild.ID, "Updated moderation settings")

	return templateData
}
