package routes

import (
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/anomalous69/fchannel/activitypub"
	"github.com/anomalous69/fchannel/config"
	"github.com/anomalous69/fchannel/db"
	"github.com/anomalous69/fchannel/route"
	"github.com/anomalous69/fchannel/util"
	"github.com/gofiber/fiber/v2"
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
	//		Allow moderators to use this for batch deletions
	var err error
	var ban db.Ban

	ban.IP, ban.Reason, ban.Date, ban.Expires, _ = db.IsIPBanned(ctx.IP())
	if len(ban.IP) > 1 {
		return ctx.Redirect(ctx.BaseURL()+"/banned", 301)
	}

	minduration, _ := strconv.Atoi(config.MinPostDelete)
	maxduration, _ := strconv.Atoi(config.MaxPostDelete)

	pwd := ctx.FormValue("pwd")

	if len(pwd) < 1 {
		return route.Send400(ctx, "No deletion password was provided")
	}
	data, err := ParseFormData(ctx)
	if err != nil {
		return route.Send400(ctx, "")
	}

	var failed []string
	var succeeded []string
	var posts []string
	var rows *sql.Rows
	var noun string // kind of overkill
	if ctx.FormValue("onlyimg") == "true" {
		noun = "post attachments"
	} else {
		noun = "posts"
	}

	for post, value := range data {
		if value == "delete" {
			posts = append(posts, post)
		}
	}
	query := `select id, posted from identify WHERE id = ANY($1) AND password = crypt($2, password)`
	if rows, err = config.DB.Query(query, posts, pwd); err != nil {
		return route.Send500(ctx, "Failed to delete "+noun, util.MakeError(err, "MultiDelete"))
	}
	valid_posts := map[string]time.Time{}

	defer rows.Close()
	for rows.Next() {
		var id string
		var posted time.Time

		if err := rows.Scan(&id, &posted); err != nil {
			failed = append(failed, fmt.Sprintf("\n%s failed due to server error", id))
		}
		valid_posts[id] = posted
	}

	if len(valid_posts) == 0 {
		return route.Send400(ctx, "Incorrect password or not from this instance, no "+noun+" were deleted")
	}

	for id, posted := range valid_posts {
		switch duration := time.Now().UTC().Sub(posted.UTC()); {
		case duration < time.Duration(minduration)*time.Second:
			failed = append(failed, fmt.Sprintf("%s too new to delete", id))
		case duration > time.Duration(maxduration)*time.Second:
			failed = append(failed, fmt.Sprintf("%s too old to delete", id))
		default:
			var actor activitypub.Actor
			var isOP bool
			var local bool

			obj := activitypub.ObjectBase{Id: id}

			objtype, err := obj.GetType()
			if err != nil {
				failed = append(failed, fmt.Sprintf("%s failed due to server error", id))
				continue
			}
			if objtype == "Tombstone" {
				failed = append(failed, fmt.Sprintf("%s was already deleted", id))
				continue
			}


			local, err = obj.IsLocal()
			if err != nil || !local {
				config.Log.Println(util.MakeError(err, "MultiDelete"))
				failed = append(failed, fmt.Sprintf("%s not from this instance", id))
				continue
			}
			isOP, err = obj.CheckIfOP()
			if err != nil {
				failed = append(failed, fmt.Sprintf("%s failed due to server error", id))
				continue
			}

			//TODO: Message for already deleted attachment
			if ctx.FormValue("onlyimg") == "true" && !isOP && local {
				if err := obj.DeleteAttachmentFromFile(); err != nil {
					config.Log.Println(util.MakeError(err, "MultiDelete"))
					failed = append(failed, fmt.Sprintf("%s failed due to server error", id))
					continue
				}

				if err := obj.TombstoneAttachment(); err != nil {
					config.Log.Println(util.MakeError(err, "MultiDelete"))
					failed = append(failed, fmt.Sprintf("%s failed due to server error", id))
					continue
				}

				if err := obj.DeletePreviewFromFile(); err != nil {
					config.Log.Println(util.MakeError(err, "MultiDelete"))
					failed = append(failed, fmt.Sprintf("%s failed due to server error", id))
					continue
				}
				
				if err := obj.TombstonePreview(); err != nil {
					config.Log.Println(util.MakeError(err, "MultiDelete"))
					failed = append(failed, fmt.Sprintf("%s failed due to server error", id))
					continue
				}
			} else if ctx.FormValue("onlyimg") != "true" {

				if isOP, _ = obj.CheckIfOP(); !isOP {
					if err := obj.Tombstone(); err != nil {
						config.Log.Println(util.MakeError(err, "MultiDelete"))
						failed = append(failed, fmt.Sprintf("%s failed due to server error", id))
						continue
					}
				} else {
					if err := obj.TombstoneReplies(); err != nil {
						config.Log.Println(util.MakeError(err, "MultiDelete"))
						failed = append(failed, fmt.Sprintf("%s failed due to server error", id))
						continue
					}
				}

				if local, _ = obj.IsLocal(); local {
					if err := obj.DeleteRequest(); err != nil {
						config.Log.Println(util.MakeError(err, "MultiDelete"))
						failed = append(failed, fmt.Sprintf("%s failed due to server error", id))
						continue
					}
				}

				if err := actor.UnArchiveLast(); err != nil {
					config.Log.Println(util.MakeError(err, "MultiDelete"))
					failed = append(failed, fmt.Sprintf("%s failed due to server error", id))
					continue
				}
			}
			//TODO: Maybe check if post/attachment is actually tombstone?
			succeeded = append(succeeded, id)
		}
	}
	for _, post := range posts {
		if _, ok := valid_posts[post]; !ok {
			failed = append(failed, fmt.Sprintf("%s password is incorrect", post))
		}
	}
	//Only show status page if something failed to delete, don't display anything if deletion was sucessfull
	if len(failed) > 0 {
		var msg strings.Builder
		fmt.Fprintf(&msg, "Failed to delete (%d) %s:\n", len(failed), noun)
		for _, fail := range failed {
			fmt.Fprintf(&msg, "\n%s", fail)
		}
		if len(succeeded) > 0 {
			fmt.Fprintf(&msg, "\n\nSucessfully deleted (%d) %s:\n", len(succeeded), noun)
			for _, success := range succeeded {
				fmt.Fprintf(&msg, "\n%s", success)
			}
		}
		return route.Send403(ctx, util.StripTransferProtocol(msg.String()))
	}
	//TODO: Maybe add a banner to page to indicate to user if action was successful
	return ctx.RedirectBack("/")
}