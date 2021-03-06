package models

import (
	"errors"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/grafana/grafana/pkg/components/simplejson"
)

// Typed errors
var (
	ErrDashboardNotFound           = errors.New("Dashboard not found")
	ErrDashboardSnapshotNotFound   = errors.New("Dashboard snapshot not found")
	ErrDashboardWithSameNameExists = errors.New("A dashboard with the same name already exists")
	ErrDashboardVersionMismatch    = errors.New("The dashboard has been changed by someone else")
	ErrDashboardUserMismatch       = errors.New("The dashboard owned by another user")
)

type UpdatePluginDashboardError struct {
	PluginId string
}

func (d UpdatePluginDashboardError) Error() string {
	return "Dashboard belong to plugin"
}

var (
	DashTypeJson     = "file"
	DashTypeDB       = "db"
	DashTypeScript   = "script"
	DashTypeSnapshot = "snapshot"
)

// Dashboard model
type Dashboard struct {
	Id       int64
	Slug     string
	OrgId    int64
	GnetId   int64
	Version  int
	PluginId string
	UserId   int64
	Private  bool

	Created time.Time
	Updated time.Time

	UpdatedBy int64
	CreatedBy int64

	Title string
	Data  *simplejson.Json
}

// NewDashboard creates a new dashboard
func NewDashboard(title string) *Dashboard {
	dash := &Dashboard{}
	dash.Data = simplejson.New()
	dash.Data.Set("title", title)
	dash.Title = title
	dash.Created = time.Now()
	dash.Updated = time.Now()
	dash.UpdateSlug()
	return dash
}

// GetTags turns the tags in data json into go string array
func (dash *Dashboard) GetTags() []string {
	return dash.Data.Get("tags").MustStringArray()
}

func NewDashboardFromJson(data *simplejson.Json) *Dashboard {
	dash := &Dashboard{}
	dash.Data = data
	dash.Title = dash.Data.Get("title").MustString()
	dash.UpdateSlug()

	if id, err := dash.Data.Get("id").Float64(); err == nil {
		dash.Id = int64(id)

		if version, err := dash.Data.Get("version").Float64(); err == nil {
			dash.Version = int(version)
			dash.Updated = time.Now()
		}
	} else {
		dash.Data.Set("version", 0)
		dash.Created = time.Now()
		dash.Updated = time.Now()
	}

	if gnetId, err := dash.Data.Get("gnetId").Float64(); err == nil {
		dash.GnetId = int64(gnetId)
	}

	if id, err := dash.Data.Get("userId").Float64(); err == nil {
		dash.UserId = int64(id)
		if private, err := dash.Data.Get("private").Bool(); err == nil {
			dash.Private = bool(private)
		}
	}

	return dash
}

// GetDashboardModel turns the command into the savable model
func (cmd *SaveDashboardCommand) GetDashboardModel() *Dashboard {
	dash := NewDashboardFromJson(cmd.Dashboard)

	if dash.Data.Get("version").MustInt(0) == 0 {
		dash.CreatedBy = cmd.UserId
	}

	dash.UpdatedBy = cmd.UserId
	dash.OrgId = cmd.OrgId
	dash.PluginId = cmd.PluginId
	dash.UserId = cmd.UserId
	dash.Private = cmd.Private
	dash.UpdateSlug()
	return dash
}

// GetString a
func (dash *Dashboard) GetString(prop string, defaultValue string) string {
	return dash.Data.Get(prop).MustString(defaultValue)
}

// UpdateSlug updates the slug
func (dash *Dashboard) UpdateSlug() {
	title := strings.ToLower(dash.Data.Get("title").MustString())
	dash.Slug = slug.Make(title)
}

//
// COMMANDS
//

type SaveDashboardCommand struct {
	Dashboard *simplejson.Json `json:"dashboard" binding:"Required"`
	Overwrite bool             `json:"overwrite"`
	OrgId     int64            `json:"-"`
	PluginId  string           `json:"-"`
	UserId    int64            `json:"-"`
	Private   bool             `json:"private"`

	Result *Dashboard
}

type DeleteDashboardCommand struct {
	Slug  string
	OrgId int64
}

//
// QUERIES
//

type GetDashboardQuery struct {
	Slug  string
	OrgId int64

	Result *Dashboard
}

type DashboardTagCloudItem struct {
	Term  string `json:"term"`
	Count int    `json:"count"`
}

type GetDashboardTagsQuery struct {
	OrgId  int64
	Result []*DashboardTagCloudItem
}

type GetDashboardsQuery struct {
	DashboardIds []int64
	Result       []*Dashboard
}

type GetDashboardsByPluginIdQuery struct {
	OrgId    int64
	PluginId string
	Result   []*Dashboard
}

type GetDashboardSlugByIdQuery struct {
	Id     int64
	Result string
}
