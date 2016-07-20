package api

import (
	"github.com/grafana/grafana/pkg/bus"
	_ "github.com/grafana/grafana/pkg/log"
	"github.com/grafana/grafana/pkg/middleware"
	m "github.com/grafana/grafana/pkg/models"
)

func ValidateOrgPlaylist(c *middleware.Context) {
	id := c.ParamsInt64(":id")
	query := m.GetPlaylistByIdQuery{Id: id}
	err := bus.Dispatch(&query)

	if err != nil {
		c.JsonApiErr(404, "Playlist not found", err)
		return
	}

	if query.Result.OrgId == 0 {
		c.JsonApiErr(404, "Playlist not found", err)
		return
	}

	if query.Result.OrgId != c.OrgId {
		c.JsonApiErr(403, "You are not allowed to edit/view playlist", nil)
		return
	}
}

func SearchPlaylists(c *middleware.Context) Response {
	query := c.Query("query")
	limit := c.QueryInt("limit")

	if limit == 0 {
		limit = 1000
	}

	searchQuery := m.GetPlaylistsQuery{
		Name:  query,
		Limit: limit,
		OrgId: c.OrgId,
	}

	err := bus.Dispatch(&searchQuery)
	if err != nil {
		return ApiError(500, "Search failed", err)
	}

	return Json(200, searchQuery.Result)
}

func GetPlaylist(c *middleware.Context) Response {
	id := c.ParamsInt64(":id")
	cmd := m.GetPlaylistByIdQuery{Id: id}

	if err := bus.Dispatch(&cmd); err != nil {
		return ApiError(500, "Playlist not found", err)
	}

	playlistDTOs, err := LoadPlaylistItemDTOs(id)

	if err != nil {
		return Json(500, err)
	}

	dto := &m.PlaylistDTO{
		Id:       cmd.Result.Id,
		Name:     cmd.Result.Name,
		Interval: cmd.Result.Interval,
		OrgId:    cmd.Result.OrgId,
		Items:    playlistDTOs,
	}

	return Json(200, dto)
}

func LoadPlaylistItemDTOs(id int64) ([]m.PlaylistItemDTO, error) {
	playlistitems, err := LoadPlaylistItems(id)

	if err != nil {
		return nil, err
	}

	playlistDTOs := make([]m.PlaylistItemDTO, 0)

	for _, item := range playlistitems {
		playlistDTOs = append(playlistDTOs, m.PlaylistItemDTO{
			Id:         item.Id,
			PlaylistId: item.PlaylistId,
			Type:       item.Type,
			Value:      item.Value,
			Order:      item.Order,
			Title:      item.Title,
		})
	}

	return playlistDTOs, nil
}

func LoadPlaylistItems(id int64) ([]m.PlaylistItem, error) {
	itemQuery := m.GetPlaylistItemsByIdQuery{PlaylistId: id}
	if err := bus.Dispatch(&itemQuery); err != nil {
		return nil, err
	}

	return *itemQuery.Result, nil
}

func GetPlaylistItems(c *middleware.Context) Response {
	id := c.ParamsInt64(":id")

	playlistDTOs, err := LoadPlaylistItemDTOs(id)

	if err != nil {
		return ApiError(500, "Could not load playlist items", err)
	}

	return Json(200, playlistDTOs)
}

func GetPlaylistDashboards(c *middleware.Context) Response {
	playlistId := c.ParamsInt64(":id")

	playlists, err := LoadPlaylistDashboards(c.OrgId, c.UserId, playlistId)
	if err != nil {
		return ApiError(500, "Could not load dashboards", err)
	}

	return Json(200, playlists)
}

func DeletePlaylist(c *middleware.Context) Response {
	id := c.ParamsInt64(":id")

	cmd := m.DeletePlaylistCommand{Id: id, OrgId: c.OrgId}
	if err := bus.Dispatch(&cmd); err != nil {
		return ApiError(500, "Failed to delete playlist", err)
	}

	return Json(200, "")
}

func CreatePlaylist(c *middleware.Context, cmd m.CreatePlaylistCommand) Response {
	cmd.OrgId = c.OrgId

	if err := bus.Dispatch(&cmd); err != nil {
		return ApiError(500, "Failed to create playlist", err)
	}

	return Json(200, cmd.Result)
}

func UpdatePlaylist(c *middleware.Context, cmd m.UpdatePlaylistCommand) Response {
	cmd.OrgId = c.OrgId

	if err := bus.Dispatch(&cmd); err != nil {
		return ApiError(500, "Failed to save playlist", err)
	}

	playlistDTOs, err := LoadPlaylistItemDTOs(cmd.Id)
	if err != nil {
		return ApiError(500, "Failed to save playlist", err)
	}

	cmd.Result.Items = playlistDTOs
	return Json(200, cmd.Result)
}
