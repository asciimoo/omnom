package webapp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type method string

const (
	GET   method = "GET"
	POST  method = "POST"
	PUT   method = "PUT"
	PATCH method = "PATCH"
	HEAD  method = "HEAD"
)

type EndpointArg struct {
	Name        string
	Type        string
	Required    bool
	Description string
}

type Endpoint struct {
	Name         string
	Path         string
	Method       method
	AuthRequired bool
	Handler      func(*gin.Context)
	Description  string
	Args         []*EndpointArg
}

var Endpoints []*Endpoint

func init() {
	Endpoints = []*Endpoint{
		&Endpoint{
			Name:         "Index",
			Path:         "/",
			Method:       GET,
			AuthRequired: false,
			Handler:      index,
			Description:  "Landing page",
		},
		&Endpoint{
			Name:         "Signup form",
			Path:         "/signup",
			Method:       GET,
			AuthRequired: false,
			Handler:      signup,
			Description:  "Signup page",
		},
		&Endpoint{
			Name:         "Signup processor",
			Path:         "/signup",
			Method:       POST,
			AuthRequired: false,
			Handler:      signup,
			Description:  "Signup form processor",
		},
		&Endpoint{
			Name:         "Login form",
			Path:         "/login",
			Method:       GET,
			AuthRequired: false,
			Handler:      login,
			Description:  "Login page",
		},
		&Endpoint{
			Name:         "Login processor",
			Path:         "/login",
			Method:       POST,
			AuthRequired: false,
			Handler:      login,
			Description:  "Login form processor",
		},
		&Endpoint{
			Name:         "Logout page",
			Path:         "/logout",
			Method:       GET,
			AuthRequired: false,
			Handler:      login,
			Description:  "Destroys user session",
		},
		&Endpoint{
			Name:         "Public bookmark listing",
			Path:         "/bookmarks",
			Method:       GET,
			AuthRequired: false,
			Handler:      bookmarks,
			Description:  "Displays public bookmarks with optional filters",
		},
		&Endpoint{
			Name:         "Snapshot view with details",
			Path:         "/snapshot",
			Method:       GET,
			AuthRequired: false,
			Handler:      snapshotWrapper,
			Description:  "Displays bookmark snapshots with additional bookmark properties",
		},
		&Endpoint{
			Name:         "Fullscreen snapshot view",
			Path:         "/view_snapshot",
			Method:       GET,
			AuthRequired: false,
			Handler:      snapshot,
			Description:  "Displays bookmark snapshots as a fullscreen page",
		},
		&Endpoint{
			Name:         "Add bookmark",
			Path:         "/add_bookmark",
			Method:       POST,
			AuthRequired: false,
			Handler:      addBookmark,
			Description:  "Add new bookmark",
		},
		&Endpoint{
			Name:         "Check bookmark",
			Path:         "/check_bookmark",
			Method:       GET,
			AuthRequired: false,
			Handler:      checkBookmark,
			Description:  "Checks if a bookmark is already exists",
		},
		&Endpoint{
			Name:         "View bookmark",
			Path:         "/bookmark",
			Method:       GET,
			AuthRequired: false,
			Handler:      viewBookmark,
			Description:  "Displays all details of a bookmark",
		},
		&Endpoint{
			Name:         "API documentation",
			Path:         "/api",
			Method:       GET,
			AuthRequired: false,
			Handler:      api,
			Description:  "Displays API documentation (this page)",
		},
		/*\
		|*| LOGIN REQUIRED FOR THE ENDPOINTS BELOW
		\*/
		&Endpoint{
			Name:         "Profile page",
			Path:         "/profile",
			Method:       GET,
			AuthRequired: true,
			Handler:      profile,
			Description:  "Displays the profile page",
		},
		&Endpoint{
			Name:         "Generate addon token",
			Path:         "/generate_addon_token",
			Method:       GET,
			AuthRequired: true,
			Handler:      generateAddonToken,
			Description:  "Creates a new addon token",
		},
		&Endpoint{
			Name:         "Delete addon token",
			Path:         "/delete_addon_token",
			Method:       POST,
			AuthRequired: true,
			Handler:      deleteAddonToken,
			Description:  "Deletes an addon token",
		},
		&Endpoint{
			Name:         "View personal bookmarks",
			Path:         "/my_bookmarks",
			Method:       GET,
			AuthRequired: true,
			Handler:      myBookmarks,
			Description:  "Displays bookmarks belongs to the current user with optional filters",
		},
		&Endpoint{
			Name:         "Edit bookmark",
			Path:         "/edit_bookmark",
			Method:       GET,
			AuthRequired: true,
			Handler:      editBookmark,
			Description:  "Displays a bookmark with all the editable properties",
		},
		&Endpoint{
			Name:         "Save bookmark",
			Path:         "/save_bookmark",
			Method:       POST,
			AuthRequired: true,
			Handler:      saveBookmark,
			Description:  "Saves an edited bookmark",
		},
		&Endpoint{
			Name:         "Delete snapshot",
			Path:         "/delete_snapshot",
			Method:       POST,
			AuthRequired: true,
			Handler:      deleteSnapshot,
			Description:  "Deletes a snapshot",
		},
	}
}

func api(c *gin.Context) {
	renderHTML(c, http.StatusOK, "api", map[string]interface{}{
		"Endpoints": Endpoints,
	})
}
