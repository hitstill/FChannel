package routes

import (
	"encoding/json"

	"github.com/FChannel0/FChannel-Server/config"
	"github.com/FChannel0/FChannel-Server/util"
	"github.com/gofiber/fiber/v2"
)

func NodeInfoDiscover(ctx *fiber.Ctx) error {
	jsonData := map[string]interface{}{
		"links": map[string]interface{}{
			"href": config.Domain + "/nodeinfo/2.1",
			"rel":  "http://nodeinfo.diaspora.software/ns/schema/2.1",
		},
	}
	ctx.Set("Content-Type", "application/json")
	links, _ := json.Marshal(jsonData)
	return ctx.Send(links)
}

func NodeInfo(ctx *fiber.Ctx) error {
	var localthreads int

	query := `SELECT COUNT(*) FROM activitystream WHERE "type" = 'Note'`
	if err := config.DB.QueryRow(query).Scan(&localthreads); err != nil {
		return util.MakeError(err, "NodeInfo")
	}
	jsonData := map[string]interface{}{
		"version": "2.1",
		"software": map[string]interface{}{
			"name":       "FChannel",
			"version":    "0.1.1", // Should be dynamic
			"repository": "https://github.com/FChannel0/FChannel-Server",
			"homepage":   "https://fchannel.org",
		},
		"protocols": []string{"activitypub"},
		"usage": map[string]interface{}{
			"localPosts": localthreads, // local posts
		},
		"openRegistrations": false,
		"services": map[string]interface{}{
			"inbound":  []string{},
			"outbound": []string{},
		},
	}

	nodeinfo, _ := json.Marshal(jsonData)
	ctx.Set("Content-Type", "application/json")
	return ctx.Send(nodeinfo)
}
