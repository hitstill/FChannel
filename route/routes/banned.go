package routes

import (
	"github.com/FChannel0/FChannel-Server/activitypub"
	"github.com/FChannel0/FChannel-Server/config"
	"github.com/FChannel0/FChannel-Server/db"
	"github.com/FChannel0/FChannel-Server/route"
	"github.com/FChannel0/FChannel-Server/util"
	"github.com/FChannel0/FChannel-Server/webfinger"

	"github.com/gofiber/fiber/v2"
)

func BannedGet(ctx *fiber.Ctx) error {

	actor, err := activitypub.GetActorFromDB(config.Domain)

	if err != nil {
		return util.MakeError(err, "BannedGet")
	}

	var data route.PageData
	data.PreferredUsername = actor.PreferredUsername
	data.Boards = webfinger.Boards
	data.Board.Name = ""
	data.Key = config.Key
	data.Board.Domain = config.Domain
	data.Board.ModCred, _ = util.GetPasswordFromSession(ctx)
	data.Board.Actor = actor
	data.Board.Post.Actor = actor.Id
	data.Board.Restricted = actor.Restricted

	data.Meta.Description = data.PreferredUsername + " is a federated image board based on ActivityPub. The current version of the code running on the server is still a work-in-progress product, expect a bumpy ride for the time being. Get the server code here: https://git.fchannel.org."
	data.Meta.Url = data.Board.Actor.Id
	data.Meta.Title = data.Title

	data.Themes = &config.Themes
	data.ThemeCookie = route.GetThemeCookie(ctx)

	var banned db.Ban
	banned.IP, banned.Reason, banned.Date, banned.Expires, _ = db.IsIPBanned(ctx.IP())

	return ctx.Render("banned", fiber.Map{"page": data, "banned": banned}, "layouts/main")
}
