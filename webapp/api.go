// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package webapp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	GET   string = "GET"
	POST  string = "POST"
	PUT   string = "PUT"
	PATCH string = "PATCH"
	HEAD  string = "HEAD"
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
	Method       string
	AuthRequired bool
	Handler      gin.HandlerFunc
	Description  string
	Args         []*EndpointArg
	RSS          string
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
			Name:         "Signup",
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
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "username",
					Type:        "string",
					Required:    true,
					Description: "Username of the new account",
				},
				&EndpointArg{
					Name:        "email",
					Type:        "string",
					Required:    true,
					Description: "Email address of the new account",
				},
			},
		},
		&Endpoint{
			Name:         "Login",
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
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "username",
					Type:        "string",
					Required:    true,
					Description: "Username or email address of the new account",
				},
			},
		},
		&Endpoint{
			Name:         "Logout",
			Path:         "/logout",
			Method:       GET,
			AuthRequired: false,
			Handler:      logout,
			Description:  "Destroys user session",
		},
		&Endpoint{
			Name:         "Snapshots",
			Path:         "/snapshots",
			Method:       GET,
			AuthRequired: false,
			Handler:      snapshots,
			Description:  "Search in snapshots by URL",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "query",
					Type:        "string",
					Required:    false,
					Description: "Search term to filter snapshots",
				},
			},
			RSS: "Snapshots",
		},
		&Endpoint{
			Name:         "Snapshot diff",
			Path:         "/snapshot_diff",
			Method:       GET,
			AuthRequired: false,
			Handler:      snapshotDiff,
			Description:  "Compare two snapshots and display the differences",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "s1",
					Type:        "string",
					Required:    true,
					Description: "ID of the first snapshot",
				},
				&EndpointArg{
					Name:        "s2",
					Type:        "string",
					Required:    true,
					Description: "ID of the second snapshot",
				},
			},
		},
		&Endpoint{
			Name:         "Snapshot diff side by side",
			Path:         "/snapshot_diff_side_by_side",
			Method:       GET,
			AuthRequired: false,
			Handler:      snapshotDiffSideBySide,
			Description:  "Display two snapshots and their differences side by side",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "s1",
					Type:        "string",
					Required:    true,
					Description: "ID of the first snapshot",
				},
				&EndpointArg{
					Name:        "s2",
					Type:        "string",
					Required:    true,
					Description: "ID of the second snapshot",
				},
			},
		},
		&Endpoint{
			Name:         "User",
			Path:         "/users/:username",
			Method:       GET,
			AuthRequired: false,
			Handler:      userProfile,
			Description:  "User profile page",
		},
		&Endpoint{
			Name:         "Public bookmarks",
			Path:         "/bookmarks",
			Method:       GET,
			AuthRequired: false,
			Handler:      bookmarks,
			Description:  "List public bookmarks with optional filters",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "query",
					Type:        "string",
					Required:    false,
					Description: "Search term to filter bookmarks by title",
				},
				&EndpointArg{
					Name:        "owner",
					Type:        "string",
					Required:    false,
					Description: "Search term to filter bookmarks by username",
				},
				&EndpointArg{
					Name:        "tag",
					Type:        "string",
					Required:    false,
					Description: "Search term to filter bookmarks by tag",
				},
				&EndpointArg{
					Name:        "domain",
					Type:        "string",
					Required:    false,
					Description: "Search term to filter bookmarks by domain",
				},
				&EndpointArg{
					Name:        "from",
					Type:        "date",
					Required:    false,
					Description: "Display only newer bookmarks. (Format: YYYY.MM.DD)",
				},
				&EndpointArg{
					Name:        "to",
					Type:        "date",
					Required:    false,
					Description: "Display only older bookmarks. (Format: YYYY.MM.DD)",
				},
				&EndpointArg{
					Name:        "search_in_snapshots",
					Type:        "boolean",
					Required:    false,
					Description: "Query parameter also applied to snapshot content. (Values: 0, 1)",
				},
				&EndpointArg{
					Name:        "search_in_notes",
					Type:        "boolean",
					Required:    false,
					Description: "Query parameter also applied to bookmark notes. (Values: 0, 1)",
				},
			},
			RSS: "Bookmarks",
		},
		&Endpoint{
			Name:         "Snapshot",
			Path:         "/snapshot",
			Method:       GET,
			AuthRequired: false,
			Handler:      snapshotWrapper,
			Description:  "Displays snapshots details with additional bookmark properties",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "sid",
					Type:        "string",
					Required:    true,
					Description: "Snapshot key",
				},
				&EndpointArg{
					Name:        "bid",
					Type:        "int",
					Required:    true,
					Description: "Bookmark ID",
				},
			},
		},
		&Endpoint{
			Name:         "Download snapshot",
			Path:         "/download_snapshot",
			Method:       GET,
			AuthRequired: false,
			Handler:      downloadSnapshot,
			Description:  "Download a self contained single file version of a snapshot",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "sid",
					Type:        "string",
					Required:    true,
					Description: "Snapshot key",
				},
			},
		},
		&Endpoint{
			Name:         "Snapshot details",
			Path:         "/snapshot_details",
			Method:       GET,
			AuthRequired: false,
			Handler:      snapshotDetails,
			Description:  "View snapshot details and resources",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "sid",
					Type:        "string",
					Required:    true,
					Description: "Snapshot key",
				},
			},
		},
		&Endpoint{
			Name:         "Collections",
			Path:         "/collections",
			Method:       POST,
			AuthRequired: false,
			Handler:      showCollections,
			Description:  "List all collections of a user in JSON format",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "token",
					Type:        "string",
					Required:    true,
					Description: "Extension token. It can be found on the profile page",
				},
			},
		},
		&Endpoint{
			Name:         "Add bookmark",
			Path:         "/add_bookmark",
			Method:       POST,
			AuthRequired: false,
			Handler:      addBookmark,
			Description:  "Add new bookmark",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "token",
					Type:        "string",
					Required:    true,
					Description: "Extension token. It can be found on the profile page",
				},
				&EndpointArg{
					Name:        "url",
					Type:        "URL",
					Required:    true,
					Description: "URL of the bookmark",
				},
				&EndpointArg{
					Name:        "title",
					Type:        "string",
					Required:    true,
					Description: "Title of the bookmark",
				},
				&EndpointArg{
					Name:        "notes",
					Type:        "string",
					Required:    false,
					Description: "Bookmark notes",
				},
				&EndpointArg{
					Name:        "favicon",
					Type:        "string",
					Required:    false,
					Description: "Data URL encoded favicon (https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/Data_URIs)",
				},
				&EndpointArg{
					Name:        "public",
					Type:        "boolean",
					Required:    false,
					Description: "Marks bookmark as public",
				},
				&EndpointArg{
					Name:        "tags",
					Type:        "string",
					Required:    false,
					Description: "Comma separated list of tags",
				},
				&EndpointArg{
					Name:        "snapshot_title",
					Type:        "string",
					Required:    false,
					Description: "Title of the snapshot",
				},
				&EndpointArg{
					Name:        "snapshot_text",
					Type:        "string",
					Required:    false,
					Description: "Text content of the snapshot",
				},
				&EndpointArg{
					Name:        "snapshot",
					Type:        "multipart file",
					Required:    false,
					Description: "Snapshot file",
				},
			},
		},
		&Endpoint{
			Name:         "Add resource",
			Path:         "/add_resource",
			Method:       POST,
			AuthRequired: false,
			Handler:      addResource,
			Description:  "Add new resource to a snapshot",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "token",
					Type:        "string",
					Required:    true,
					Description: "Extension token. It can be found on the profile page",
				},
				&EndpointArg{
					Name:        "sid",
					Type:        "string",
					Required:    true,
					Description: "Snapshot ID",
				},
				&EndpointArg{
					Name:        "meta",
					Type:        "JSON string",
					Required:    true,
					Description: "List of resource metadata containing objects with mimetype, extension and filename information",
				},
				&EndpointArg{
					Name:        "resource[0-9]+",
					Type:        "multipart files",
					Required:    true,
					Description: "Resource files",
				},
			},
		},
		&Endpoint{
			Name:         "Check bookmark",
			Path:         "/check_bookmark",
			Method:       GET,
			AuthRequired: false,
			Handler:      checkBookmark,
			Description:  "Checks if a bookmark is already exists",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "token",
					Type:        "string",
					Required:    true,
					Description: "Extension token. It can be found on the profile page",
				},
				&EndpointArg{
					Name:        "url",
					Type:        "URL",
					Required:    true,
					Description: "URL of the bookmark",
				},
			},
		},
		&Endpoint{
			Name:         "Bookmark",
			Path:         "/bookmark",
			Method:       GET,
			AuthRequired: false,
			Handler:      viewBookmark,
			Description:  "Displays all details of a bookmark",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "id",
					Type:        "int",
					Required:    true,
					Description: "Bookmark ID",
				},
			},
		},
		&Endpoint{
			Name:         "API",
			Path:         "/api",
			Method:       GET,
			AuthRequired: false,
			Handler:      api,
			Description:  "Displays API documentation (this page)",
		},
		&Endpoint{
			Name:         "Oauth",
			Path:         "/oauth",
			Method:       GET,
			AuthRequired: false,
			Handler:      oauthHandler,
			Description:  "Creates OAuth requests",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "provider",
					Type:        "string",
					Required:    true,
					Description: "Oauth provider name",
				},
			},
		},
		&Endpoint{
			Name:         "Oauth verification",
			Path:         "/oauth_redirect_handler",
			Method:       GET,
			AuthRequired: false,
			Handler:      oauthRedirectHandler,
			Description:  "Verifies OAuth requests",
		},
		&Endpoint{
			Name:         "Check token",
			Path:         "/check_token",
			Method:       POST,
			AuthRequired: false,
			Handler:      checkAddonToken,
			Description:  "Verifies addon tokens",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "token",
					Type:        "string",
					Required:    true,
					Description: "Addon token",
				},
			},
		},
		&Endpoint{
			Name:         "ActivityPub inbox",
			Path:         "/inbox/:username",
			Method:       POST,
			AuthRequired: false,
			Handler:      apInboxResponse,
			Description:  "Inbox for ActivityPub messages",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "message",
					Type:        "JSON",
					Required:    true,
					Description: "ActivityPub message",
				},
			},
		},
		&Endpoint{
			Name:         "ActivityPub outbox",
			Path:         "/outbox/:username",
			Method:       GET,
			AuthRequired: false,
			Handler:      apOutboxResponse,
			Description:  "Outbox of ActivityPub messages",
		},
		&Endpoint{
			Name:         "ActivityPub webfinger",
			Path:         "/.well-known/webfinger",
			Method:       GET,
			AuthRequired: false,
			Handler:      apWebfingerResponse,
			Description:  "Webfinger response for ActivityPub",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "resource",
					Type:        "string",
					Required:    true,
					Description: "ActivityPub resource",
				},
			},
		},
		/****************************************\
		| LOGIN REQUIRED FOR THE ENDPOINTS BELOW |
		\****************************************/
		&Endpoint{
			Name:         "Profile",
			Path:         "/profile",
			Method:       GET,
			AuthRequired: true,
			Handler:      profile,
			Description:  "Displays the user profile page",
		},
		&Endpoint{
			Name:         "Profile page",
			Path:         "/profile",
			Method:       POST,
			AuthRequired: true,
			Handler:      profile,
			Description:  "Displays the profile page with addon tokens",
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
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "id",
					Type:        "int",
					Required:    true,
					Description: "Token ID",
				},
			},
		},
		&Endpoint{
			Name:         "Create bookmark form",
			Path:         "/create_bookmark",
			Method:       GET,
			AuthRequired: true,
			Handler:      createBookmarkForm,
			Description:  "Show create bookmark form",
		},
		&Endpoint{
			Name:         "Create bookmark",
			Path:         "/create_bookmark",
			Method:       POST,
			AuthRequired: true,
			Handler:      createBookmark,
			Description:  "Create new bookmark from webapp",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "url",
					Type:        "string",
					Required:    true,
					Description: "URL of the bookmark",
				},
				&EndpointArg{
					Name:        "title",
					Type:        "string",
					Required:    true,
					Description: "Title of the bookmark",
				},
				&EndpointArg{
					Name:        "notes",
					Type:        "string",
					Required:    false,
					Description: "Bookmark notes",
				},
				&EndpointArg{
					Name:        "public",
					Type:        "boolean",
					Required:    false,
					Description: "Bookmark is publicly accessible",
				},
			},
		},
		&Endpoint{
			Name:         "My bookmarks",
			Path:         "/my_bookmarks",
			Method:       GET,
			AuthRequired: true,
			Handler:      myBookmarks,
			Description:  "Displays bookmarks belongs to the current user with optional filters",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "query",
					Type:        "string",
					Required:    false,
					Description: "Search term to filter bookmarks by title",
				},
				&EndpointArg{
					Name:        "tag",
					Type:        "string",
					Required:    false,
					Description: "Search term to filter bookmarks by tag",
				},
				&EndpointArg{
					Name:        "domain",
					Type:        "string",
					Required:    false,
					Description: "Search term to filter bookmarks by domain",
				},
				&EndpointArg{
					Name:        "from",
					Type:        "date",
					Required:    false,
					Description: "Display only newer bookmarks. (Format: YYYY.MM.DD)",
				},
				&EndpointArg{
					Name:        "to",
					Type:        "date",
					Required:    false,
					Description: "Display only older bookmarks. (Format: YYYY.MM.DD)",
				},
				&EndpointArg{
					Name:        "is_public",
					Type:        "boolean",
					Required:    false,
					Description: "Display only public bookmarks. (Values: 0, 1)",
				},
				&EndpointArg{
					Name:        "is_private",
					Type:        "boolean",
					Required:    false,
					Description: "Display only private bookmarks. (Values: 0, 1)",
				},
				&EndpointArg{
					Name:        "search_in_snapshots",
					Type:        "boolean",
					Required:    false,
					Description: "Query parameter also applied to snapshot content. (Values: 0, 1)",
				},
				&EndpointArg{
					Name:        "search_in_notes",
					Type:        "boolean",
					Required:    false,
					Description: "Query parameter also applied to bookmark notes. (Values: 0, 1)",
				},
			},
			RSS: "Bookmarks",
		},
		&Endpoint{
			Name:         "Feeds",
			Path:         "/feeds",
			Method:       GET,
			AuthRequired: true,
			Handler:      feeds,
			Description:  "List feeds",
		},
		&Endpoint{
			Name:         "Add feed",
			Path:         "/add_feed",
			Method:       POST,
			AuthRequired: true,
			Handler:      addFeed,
			Description:  "Add new feed",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "name",
					Type:        "string",
					Required:    true,
					Description: "Feed name",
				},
				&EndpointArg{
					Name:        "url",
					Type:        "string",
					Required:    true,
					Description: "Feed URL",
				},
			},
		},
		&Endpoint{
			Name:         "Edit bookmark",
			Path:         "/edit_bookmark",
			Method:       GET,
			AuthRequired: true,
			Handler:      editBookmark,
			Description:  "Displays a bookmark with all the editable properties",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "id",
					Type:        "int",
					Required:    true,
					Description: "Bookmark ID",
				},
			},
		},
		&Endpoint{
			Name:         "Save bookmark",
			Path:         "/save_bookmark",
			Method:       POST,
			AuthRequired: true,
			Handler:      saveBookmark,
			Description:  "Saves an edited bookmark",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "id",
					Type:        "int",
					Required:    true,
					Description: "Bookmark ID",
				},
				&EndpointArg{
					Name:        "title",
					Type:        "string",
					Required:    true,
					Description: "Title of the bookmark",
				},
				&EndpointArg{
					Name:        "notes",
					Type:        "string",
					Required:    false,
					Description: "Bookmark notes",
				},
				&EndpointArg{
					Name:        "public",
					Type:        "boolean",
					Required:    false,
					Description: "Bookmark is publicly accessible. (Omit this argument in case of private bookmarks)",
				},
			},
		},
		&Endpoint{
			Name:         "Delete snapshot",
			Path:         "/delete_snapshot",
			Method:       POST,
			AuthRequired: true,
			Handler:      deleteSnapshot,
			Description:  "Deletes a snapshot",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "bid",
					Type:        "int",
					Required:    true,
					Description: "Bookmark ID",
				},
				&EndpointArg{
					Name:        "sid",
					Type:        "int",
					Required:    true,
					Description: "Snapshot ID",
				},
			},
		},
		&Endpoint{
			Name:         "Delete bookmark",
			Path:         "/delete_bookmark",
			Method:       POST,
			AuthRequired: true,
			Handler:      deleteBookmark,
			Description:  "Deletes a bookmark",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "id",
					Type:        "int",
					Required:    true,
					Description: "Bookmark ID",
				},
			},
		},
		&Endpoint{
			Name:         "Add tag",
			Path:         "/add_tag",
			Method:       POST,
			AuthRequired: true,
			Handler:      addTag,
			Description:  "Add tag to a bookmark",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "bid",
					Type:        "int",
					Required:    true,
					Description: "Bookmark ID",
				},
				&EndpointArg{
					Name:        "tag",
					Type:        "string",
					Required:    true,
					Description: "Tag string",
				},
			},
		},
		&Endpoint{
			Name:         "Delete tag",
			Path:         "/delete_tag",
			Method:       POST,
			AuthRequired: true,
			Handler:      deleteTag,
			Description:  "Delete tag's bookmark",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "bid",
					Type:        "int",
					Required:    true,
					Description: "Bookmark ID",
				},
				&EndpointArg{
					Name:        "tid",
					Type:        "int",
					Required:    true,
					Description: "Tag ID",
				},
			},
		},
		&Endpoint{
			Name:         "edit collection form",
			Path:         "/edit_collection",
			Method:       GET,
			AuthRequired: true,
			Handler:      editCollection,
			Description:  "View to create or edit collections",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "collection_id",
					Type:        "int",
					Required:    false,
					Description: "collection id",
				},
			},
		},
		&Endpoint{
			Name:         "save collection",
			Path:         "/edit_collection",
			Method:       POST,
			AuthRequired: true,
			Handler:      saveCollection,
			Description:  "Create or edit collections",
			Args: []*EndpointArg{
				&EndpointArg{
					Name:        "collection_id",
					Type:        "int",
					Required:    false,
					Description: "collection id",
				},
			},
		},
	}
}

func api(c *gin.Context) {
	render(c, http.StatusOK, "api", map[string]interface{}{
		"Endpoints": Endpoints,
	})
}
