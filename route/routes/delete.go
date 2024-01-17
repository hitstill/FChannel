package routes

import (
	"database/sql"
	"net/url"
	"strconv"
	"time"

	"github.com/FChannel0/FChannel-Server/activitypub"
	"github.com/FChannel0/FChannel-Server/config"
	"github.com/FChannel0/FChannel-Server/db"
	"github.com/FChannel0/FChannel-Server/route"
	"github.com/FChannel0/FChannel-Server/util"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

func ParseFormData(ctx *fiber.Ctx) (map[string]string, error) {
	values, err := url.ParseQuery(string(ctx.Body()))
	if err != nil {
		return nil, err
	}

	obj := map[string]string{}
	for k, v := range values {
		if len(v) > 0 {
			obj[k] = v[0]
		}
	}

	return obj, nil
}

func MultiDelete(ctx *fiber.Ctx) error {
	//TODO: Notify user when one or more deletions fails (replace SendString placeholders)
	//		Allow moderators to use this for batch deletions
	var err error
	var ban db.Ban
	var failed int

	ban.IP, ban.Reason, ban.Date, ban.Expires, _ = db.IsIPBanned(ctx.IP())
	if len(ban.IP) > 1 {
		return ctx.Redirect(ctx.BaseURL()+"/banned", 301)
	}

	minduration, _ := strconv.Atoi(config.MinPostDelete)
	maxduration, _ := strconv.Atoi(config.MaxPostDelete)

	pwd := ctx.FormValue("pwd")

	if len(pwd) < 1 {
		return route.Send400(ctx, "No deletion password provided")
	}
	data, err := ParseFormData(ctx)
	if err != nil {
		return util.MakeError(err, "MultiDelete")
	}

	var posts []string
	var rows *sql.Rows

	for name, value := range data {
		if value == "delete" {
			posts = append(posts, name)
		}
	}
	query := `select id, posted from identify WHERE id = ANY($1) AND password = crypt($2, password)`
	if rows, err = config.DB.Query(query, posts, pwd); err != nil {
		return util.MakeError(err, "MultiDelete")
	}
	valid_posts := map[string]time.Time{}

	defer rows.Close()
	for rows.Next() {
		var id string
		var posted time.Time

		if err := rows.Scan(&id, &posted); err != nil {
			if err == pgx.ErrNoRows {
				return route.GenericError(ctx, "Incorrect password, no posts were deleted.")
			} else {
				return route.Send500(ctx, err)
			}
		}
		valid_posts[id] = posted
	}

	for id, posted := range valid_posts {
		switch duration := time.Now().UTC().Sub(posted.UTC()); {
		case duration < time.Duration(minduration)*time.Second:
			failed++
		case duration > time.Duration(maxduration)*time.Second:
			failed++
		default:
			var actor activitypub.Actor
			var isOP bool
			var local bool

			obj := activitypub.ObjectBase{Id: id}

			isOP, err = obj.CheckIfOP()
			if err != nil {
				return util.MakeError(err, "MultiDelete")
			}

			local, err = obj.IsLocal()
			if err != nil || !local {
				return route.GenericError(ctx, "Cannot delete non-local post.")
			}

			if ctx.FormValue("onlyimg") == "true" && !isOP && local {
				if err := obj.DeleteAttachmentFromFile(); err != nil {
					return util.MakeError(err, "MultiDelete")
				}

				if err := obj.TombstoneAttachment(); err != nil {
					return util.MakeError(err, "MultiDelete")
				}

				if err := obj.DeletePreviewFromFile(); err != nil {
					return util.MakeError(err, "MultiDelete")
				}

				if err := obj.TombstonePreview(); err != nil {
					return util.MakeError(err, "MultiDelete")
				}
			} else if ctx.FormValue("onlyimg") != "true" {

				if isOP, _ = obj.CheckIfOP(); !isOP {
					if err := obj.Tombstone(); err != nil {
						return util.MakeError(err, "MultiDelete")
					}
				} else {
					if err := obj.TombstoneReplies(); err != nil {
						return util.MakeError(err, "MultiDelete")
					}
				}

				if local, _ = obj.IsLocal(); local {
					if err := obj.DeleteRequest(); err != nil {
						return util.MakeError(err, "MultiDelete")
					}
				}

				if err := actor.UnArchiveLast(); err != nil {
					return util.MakeError(err, "MultiDelete")
				}
			}
		}
	}
	if failed > 0 {
		return route.GenericError(ctx, strconv.Itoa(failed)+" post(s) were too old or new and have not been deleted.")
	}
	return ctx.RedirectBack("/")
}
