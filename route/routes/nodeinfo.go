package routes

import (
	"encoding/json"
	"strings"

	"github.com/FChannel0/FChannel-Server/config"
	"github.com/FChannel0/FChannel-Server/util"
	"github.com/gofiber/fiber/v2"
)

func NodeInfoDiscover(ctx *fiber.Ctx) error {
	jsonData := map[string]interface{}{
		"links": []map[string]string{
			{
				"href": config.Domain + "/nodeinfo/2.0.json",
				"rel":  "http://nodeinfo.diaspora.software/ns/schema/2.0",
			},
			{
				"href": config.Domain + "/nodeinfo/2.1.json",
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
	schemaVersion := strings.TrimSuffix(ctx.Params("version"), ".json")

	if schemaVersion != "2.0" && schemaVersion != "2.1" {
		errorData := map[string]string{
			"error": "Nodeinfo schema version not handled",
		}
		errorJSON, _ := json.Marshal(errorData)
		ctx.Set("Content-Type", "application/json")
		return ctx.Send(errorJSON)
	}

	query := `SELECT  (SELECT COUNT(*) FROM activitystream WHERE type = 'Note' ) AS local, (SELECT COUNT(*) FROM cacheactivitystream WHERE type = 'Note') AS remote, (SELECT count(*) FROM activitystream where type = 'Archive') + (SELECT count(*) FROM cacheactivitystream where type = 'Archive') AS archived`
	if err := config.DB.QueryRow(query).Scan(&localPosts, &remotePosts, &archivedPosts); err != nil {
		return util.MakeError(err, "NodeInfo")
	}

	jsonData := map[string]interface{}{
		"version": schemaVersion,
		"software": map[string]interface{}{
			"name":       "FChannel",
			"version":    "0.1.1", // Should be dynamic
			"repository": "https://github.com/FChannel0/FChannel-Server",
			"homepage":   "https://fchannel.org",
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
		"nodeName":        config.InstanceName,
		"nodeDescription": config.InstanceSummary,
		"metadata": map[string]interface{}{
			"remotePosts":   remotePosts,
			"archivedPosts": archivedPosts,
		},
	}

	nodeinfo, _ := json.Marshal(jsonData)
	ctx.Set("Content-Type", "application/json")
	return ctx.Send(nodeinfo)
}
