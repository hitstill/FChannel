package routes

import (
	"errors"
	"net/http"
	"regexp"
	"sort"
	"time"

	"github.com/anomalous69/fchannel/activitypub"
	"github.com/anomalous69/fchannel/config"
	"github.com/anomalous69/fchannel/db"
	"github.com/anomalous69/fchannel/util"
	"github.com/gofiber/fiber/v2"
)

func AdminVerify(ctx *fiber.Ctx) error {
	identifier := ctx.FormValue("id")
	code := ctx.FormValue("code")

	v, err := util.GetVerificationByCode(code)

	if err != nil {
		return Send403(ctx, "Invalid code or identifier", err)
	}

	if v.Identifier != identifier {
		return Send500(ctx, "Incorrect identifier for code")
	}

	ctx.Cookie(&fiber.Cookie{
		Name:    "session_token",
		Value:   v.Board + "|" + v.Code,
		Expires: time.Now().UTC().Add(730 * time.Hour),
	})

	return ctx.Redirect("/", http.StatusSeeOther)
}

func AdminIndex(ctx *fiber.Ctx) error {
	id, _ := util.GetPasswordFromSession(ctx)
	actor, _ := activitypub.GetActorFromPath(ctx.Path(), "/"+config.C.ModKey+"/")

	if actor.Id == "" {
		actor, _ = activitypub.GetActorByNameFromDB(config.C.Instance.Domain)
	}

	if id == "" || (id != actor.Id && id != config.C.Instance.Domain) {
		return ctx.Render("verify", fiber.Map{"key": config.C.ModKey})
	}

	actor, err := activitypub.GetActor(config.C.Instance.Domain)

	if err != nil {
		return util.MakeError(err, "AdminIndex")
	}

	reqActivity := activitypub.Activity{Id: actor.Following}
	follow, _ := reqActivity.GetCollection()
	follower, _ := reqActivity.GetCollection()

	var following []string
	var followers []string

	for _, e := range follow.Items {
		following = append(following, e.Id)
	}

	for _, e := range follower.Items {
		followers = append(followers, e.Id)
	}

	var adminData AdminPage
	adminData.Following = following
	adminData.Followers = followers

	var reported = make(map[string][]db.Reports)

	for _, e := range following {
		re := regexp.MustCompile(`.*/(.+)$`)
		boards := re.FindStringSubmatch(e)
		reports, _ := db.GetLocalReport(boards[1])

		for _, k := range reports {
			reported[k.Actor.Name] = append(reported[k.Actor.Name], k)
		}
	}

	for k, e := range reported {
		sort.Sort(db.ReportsSortDesc(e))
		reported[k] = e
	}

	adminData.Actor = actor.Id
	adminData.Key = config.C.ModKey
	adminData.Domain = config.C.Instance.Domain
	adminData.Board.ModCred, _ = util.GetPasswordFromSession(ctx)
	adminData.Title = actor.Name + " Admin page"

	adminData.Boards = activitypub.Boards

	adminData.Board.Post.Actor = actor.Id

	adminData.Instance, _ = activitypub.GetActorFromDB(config.C.Instance.Domain)

	adminData.PostBlacklist, _ = util.GetRegexBlacklist()

	adminData.Meta.Description = adminData.Title
	adminData.Meta.Url = adminData.Board.Actor.Id
	adminData.Meta.Title = adminData.Title

	adminData.Themes = &config.Themes
	adminData.ThemeCookie = GetThemeCookie(ctx)

	adminData.ServerVersion = config.Version

	return ctx.Render("admin", fiber.Map{
		"page":    adminData,
		"reports": reported,
	}, "layouts/main")
}

func AdminFollow(ctx *fiber.Ctx) error {
	follow := ctx.FormValue("follow")
	actorId := ctx.FormValue("actor")

	actor := activitypub.Actor{Id: actorId}
	followActivity, _ := actor.MakeFollowActivity(follow)

	objActor := activitypub.Actor{Id: followActivity.Object.Actor}

	if isLocal, _ := objActor.IsLocal(); !isLocal && followActivity.Actor.Id == config.C.Instance.Domain {
		_, err := ctx.Write([]byte("main board can only follow local boards. Create a new board and then follow outside boards from it."))
		return util.MakeError(err, "AdminIndex")
	}

	if actor, _ := activitypub.FingerActor(follow); actor.Id != "" {
		if err := followActivity.MakeRequestOutbox(); err != nil {
			return util.MakeError(err, "AdminFollow")
		}
	}

	var redirect string
	actor, _ = activitypub.GetActorFromPath(ctx.Path(), "/"+config.C.ModKey+"/")

	if actor.Name != "main" {
		redirect = actor.Name
	}

	time.Sleep(time.Duration(500) * time.Millisecond)

	return ctx.Redirect("/"+config.C.ModKey+"/"+redirect, http.StatusSeeOther)
}

func AdminAddBoard(ctx *fiber.Ctx) error {
	actor, _ := activitypub.GetActorFromDB(config.C.Instance.Domain)

	if hasValidation := actor.HasValidation(ctx); !hasValidation {
		return nil
	}

	var newActorActivity activitypub.Activity
	var board activitypub.Actor

	var restrict bool
	if ctx.FormValue("restricted") == "True" {
		restrict = true
	} else {
		restrict = false
	}

	board.Name = ctx.FormValue("name")
	board.PreferredUsername = ctx.FormValue("prefname")
	board.Summary = ctx.FormValue("summary")
	board.Restricted = restrict
	board.BoardType = ctx.FormValue("boardtype")
	if board.BoardType != "image" && board.BoardType != "text" && board.BoardType != "flash" {
		return Send400(ctx, "Board type \""+board.BoardType+"\" is invalid")
	}

	newActorActivity.AtContext.Context = "https://www.w3.org/ns/activitystreams"
	newActorActivity.Type = "New"

	var nobj activitypub.ObjectBase
	newActorActivity.Actor = &actor
	newActorActivity.Object = nobj

	newActorActivity.Object.Alias = board.Name
	newActorActivity.Object.Name = board.PreferredUsername
	newActorActivity.Object.Summary = board.Summary
	newActorActivity.Object.Sensitive = board.Restricted
	newActorActivity.Object.MediaType = board.BoardType // Didn't want to add new struct field, close enough

	newActorActivity.MakeRequestOutbox()

	time.Sleep(time.Duration(500) * time.Millisecond)

	return ctx.Redirect("/"+config.C.ModKey, http.StatusSeeOther)
}

func AdminActorIndex(ctx *fiber.Ctx) error {
	var data AdminPage

	id, pass := util.GetPasswordFromSession(ctx)
	actor, _ := activitypub.GetActorFromPath(ctx.Path(), "/"+config.C.ModKey+"/")

	if actor.Id == "" {
		actor, _ = activitypub.GetActorByNameFromDB(config.C.Instance.Domain)
	}

	var hasAuth bool
	hasAuth, data.Board.ModCred = util.HasAuth(pass, actor.Id)

	if !hasAuth || (id != actor.Id && id != config.C.Instance.Domain) {
		return ctx.Render("verify", fiber.Map{"key": config.C.ModKey})
	}

	reqActivity := activitypub.Activity{Id: actor.Following}
	follow, _ := reqActivity.GetCollection()

	reqActivity.Id = actor.Followers
	follower, _ := reqActivity.GetCollection()

	var following []string
	var followers []string

	for _, e := range follow.Items {
		following = append(following, e.Id)
	}

	for _, e := range follower.Items {
		followers = append(followers, e.Id)
	}

	data.Following = following
	data.Followers = followers

	reports, _ := db.GetLocalReport(actor.Name)

	var reported = make(map[string][]db.Reports)
	for _, k := range reports {
		reported[k.Actor.Name] = append(reported[k.Actor.Name], k)
	}

	for k, e := range reported {
		sort.Sort(db.ReportsSortDesc(e))
		reported[k] = e
	}

	data.Domain = config.C.Instance.Domain
	data.IsLocal, _ = actor.IsLocal()
	data.Title = "Manage /" + actor.Name + "/"
	data.Boards = activitypub.Boards
	data.Board.Name = actor.Name
	data.Board.Actor = actor
	data.Key = config.C.ModKey
	data.Board.TP = config.C.Instance.Tp

	data.Board.Post.Actor = actor.Id

	data.Instance, _ = activitypub.GetActorFromDB(config.C.Instance.Domain)

	data.AutoSubscribe, _ = actor.GetAutoSubscribe()
	data.BoardType = actor.BoardType

	jannies, err := actor.GetJanitors()

	if err != nil {
		return util.MakeError(err, "AdminActorIndex")
	}

	data.Meta.Description = data.Title
	data.Meta.Url = data.Board.Actor.Id
	data.Meta.Title = data.Title

	data.Themes = &config.Themes

	data.RecentPosts, _ = actor.GetRecentPosts()

	if cookie := ctx.Cookies("theme"); cookie != "" {
		data.ThemeCookie = cookie
	}

	data.ServerVersion = config.Version

	return ctx.Render("manage", fiber.Map{
		"page":    data,
		"jannies": jannies,
		"reports": reported,
	}, "layouts/main")
}

func AdminAddJanny(ctx *fiber.Ctx) error {
	id, pass := util.GetPasswordFromSession(ctx)
	actor, _ := activitypub.GetActorFromPath(ctx.Path(), "/"+config.C.ModKey+"/")

	if actor.Id == "" {
		actor, _ = activitypub.GetActorByNameFromDB(config.C.Instance.Domain)
	}

	hasAuth, _type := util.HasAuth(pass, actor.Id)

	if !hasAuth || _type != "admin" || (id != actor.Id && id != config.C.Instance.Domain) {
		return util.MakeError(errors.New("Error"), "AdminJanny")
	}

	var verify util.Verify
	verify.Type = "janitor"
	verify.Identifier = actor.Id
	verify.Label = ctx.FormValue("label")

	if err := actor.CreateVerification(verify); err != nil {
		return util.MakeError(err, "CreateNewBoardDB")
	}

	var redirect string
	actor, _ = activitypub.GetActorFromPath(ctx.Path(), "/"+config.C.ModKey+"/")

	if actor.Name != "main" {
		redirect = actor.Name
	}

	return ctx.Redirect("/"+config.C.ModKey+"/"+redirect, http.StatusSeeOther)
}

func AdminEditSummary(ctx *fiber.Ctx) error {
	id, pass := util.GetPasswordFromSession(ctx)
	actor, _ := activitypub.GetActorFromPath(ctx.Path(), "/"+config.C.ModKey+"/")

	if actor.Id == "" {
		actor, _ = activitypub.GetActorByNameFromDB(config.C.Instance.Domain)
	}

	hasAuth, _type := util.HasAuth(pass, actor.Id)

	if !hasAuth || _type != "admin" || (id != actor.Id && id != config.C.Instance.Domain) {
		return util.MakeError(errors.New("Error"), "AdminEditSummary")
	}

	summary := ctx.FormValue("summary")

	query := `update actor set summary=$1 where id=$2`
	if _, err := config.DB.Exec(query, summary, actor.Id); err != nil {
		return util.MakeError(err, "AdminEditSummary")
	}

	var redirect string
	if actor.Name != "main" {
		redirect = actor.Name
	}

	return ctx.Redirect("/"+config.C.ModKey+"/"+redirect, http.StatusSeeOther)

}

func AdminSetBoardType(ctx *fiber.Ctx) error {
	id, pass := util.GetPasswordFromSession(ctx)
	actor, _ := activitypub.GetActorFromPath(ctx.Path(), "/"+config.C.ModKey+"/")

	if actor.Id == "" {
		actor, _ = activitypub.GetActorByNameFromDB(config.C.Instance.Domain)
	}

	hasAuth, _type := util.HasAuth(pass, actor.Id)

	if !hasAuth || _type != "admin" || (id != actor.Id && id != config.C.Instance.Domain) {
		return util.MakeError(errors.New("Error"), "AdminEditSummary")
	}

	boardtype := ctx.FormValue("boardtype")

	if boardtype == "image" || boardtype == "text" || boardtype == "flash" {

		query := `update actor set boardtype=$1 where id=$2`
		if _, err := config.DB.Exec(query, boardtype, actor.Id); err != nil {
			return util.MakeError(err, "AdminEditSummary")
		}

		var redirect string
		if actor.Name != "main" {
			redirect = actor.Name
		}

		return ctx.Redirect("/"+config.C.ModKey+"/"+redirect, http.StatusSeeOther)
	}
	return Send400(ctx, "Board type \""+boardtype+"\" is invalid")
}

func AdminDeleteJanny(ctx *fiber.Ctx) error {
	id, pass := util.GetPasswordFromSession(ctx)
	actor, _ := activitypub.GetActorFromPath(ctx.Path(), "/"+config.C.ModKey+"/")

	if actor.Id == "" {
		actor, _ = activitypub.GetActorByNameFromDB(config.C.Instance.Domain)
	}

	hasAuth, _type := util.HasAuth(pass, actor.Id)

	if !hasAuth || _type != "admin" || (id != actor.Id && id != config.C.Instance.Domain) {
		return util.MakeError(errors.New("Error"), "AdminJanny")
	}

	var verify util.Verify
	verify.Code = ctx.Query("code")

	if err := actor.DeleteVerification(verify); err != nil {
		return util.MakeError(err, "AdminDeleteJanny")
	}

	var redirect string

	if actor.Name != "main" {
		redirect = actor.Name
	}

	return ctx.Redirect("/"+config.C.ModKey+"/"+redirect, http.StatusSeeOther)
}
