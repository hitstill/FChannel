package routes

import (
	"encoding/json"
	"strings"

	"github.com/anomalous69/fchannel/config"
	"github.com/anomalous69/fchannel/util"
	"github.com/gofiber/fiber/v2"
)

func NodeInfoDiscover(ctx *fiber.Ctx) error {
	jsonData := map[string]interface{}{
		"links": []map[string]string{
			{
				"href": config.C.Instance.Domain + "/nodeinfo/2.0.json",
				"rel":  "http://nodeinfo.diaspora.software/ns/schema/2.0",
			},
			{
				"href": config.C.Instance.Domain + "/nodeinfo/2.1.json",
				"rel":  "http://nodeinfo.diaspora.software/ns/schema/2.1",
			},
		},
	}

	ctx.Set("Content-Type", "application/json")
	links, _ := json.Marshal(jsonData)
	return ctx.Send(links)
}

func NodeInfo(ctx *fiber.Ctx) error {
	var localPosts int
	var remotePosts int
	var archivedPosts int
	var avgPPD float64

	schemaVersion := strings.TrimSuffix(ctx.Params("version"), ".json")

	if schemaVersion != "2.0" && schemaVersion != "2.1" {
		errorData := map[string]string{
			"error": "Nodeinfo schema version not handled",
		}
		errorJSON, _ := json.Marshal(errorData)
		ctx.Set("Content-Type", "application/json")
		return ctx.Send(errorJSON)
	}

	query := `SELECT  (SELECT COUNT(*) FROM activitystream WHERE type = 'Note' ) AS local, (SELECT COUNT(*) FROM cacheactivitystream WHERE type = 'Note') AS remote, (SELECT count(*) FROM activitystream where type = 'Archive') + (SELECT count(*) FROM cacheactivitystream where type = 'Archive') AS archived, (SELECT AVG(last_4_week) AS avg_ppd FROM
	(SELECT COUNT(*) AS last_4_week
	FROM activitystream
	WHERE published::date  > CURRENT_DATE -28 AND type in ('Note', 'Tombstone') AND mediatype = ''
	GROUP BY published :: date) A)`
	if err := config.DB.QueryRow(query).Scan(&localPosts, &remotePosts, &archivedPosts, &avgPPD); err != nil {
		return util.MakeError(err, "NodeInfo")
	}

	jsonData := map[string]interface{}{
		"version": schemaVersion,
		"software": map[string]interface{}{
			"name":       "FChannel",
			"version":    config.Version,
			"repository": "https://github.com/anomalous69/FChannel",
			"homepage":   "https://github.com/anomalous69/FChannel",
		},
		"protocols": []string{"activitypub"},
		"usage": map[string]interface{}{
			"localPosts": localPosts,
		},
		"openRegistrations": false,
		"services": map[string]interface{}{
			"inbound":  []string{},
			"outbound": []string{},
		},
		"nodeName":        config.C.Instance.Name,
		"nodeDescription": config.C.Instance.Summary,
		"metadata": map[string]interface{}{
			"remotePosts":    remotePosts,
			"archivedPosts":  archivedPosts,
			"avgPostsPerDay": avgPPD,
		},
	}

	nodeinfo, _ := json.Marshal(jsonData)
	ctx.Set("Content-Type", "application/json")
	return ctx.Send(nodeinfo)
}
