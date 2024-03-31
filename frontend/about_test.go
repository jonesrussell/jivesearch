package frontend

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestAboutHandler(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	raw := `[
		{
		  "login": "bobsomebody",
		  "id": 0,
		  "avatar_url": "https://avatars1.githubusercontent.com/u/0?v=4",
		  "gravatar_id": "",
		  "url": "https://api.github.com/users/bobsomebody",
		  "html_url": "https://github.com/bobsomebody",
		  "followers_url": "https://api.github.com/users/bobsomebody/followers",
		  "following_url": "https://api.github.com/users/bobsomebody/following{/other_user}",
		  "gists_url": "https://api.github.com/users/bobsomebody/gists{/gist_id}",
		  "starred_url": "https://api.github.com/users/bobsomebody/starred{/owner}{/repo}",
		  "subscriptions_url": "https://api.github.com/users/bobsomebody/subscriptions",
		  "organizations_url": "https://api.github.com/users/bobsomebody/orgs",
		  "repos_url": "https://api.github.com/users/bobsomebody/repos",
		  "events_url": "https://api.github.com/users/bobsomebody/events{/privacy}",
		  "received_events_url": "https://api.github.com/users/bobsomebody/received_events",
		  "type": "User",
		  "site_admin": false,
		  "contributions": 19
		}
	  ]`

	responder := httpmock.NewStringResponder(200, raw)
	httpmock.RegisterResponder("GET", "https://api.github.com/repos/jonesrussell/jivesearch/contributors", responder)

	f := &Frontend{
		GitHub: GitHub{
			HTTPClient: &http.Client{},
		},
		Onion: "my.onion",
	}

	req, err := http.NewRequest("GET", "/about", nil)
	if err != nil {
		t.Fatal(err)
	}

	want := &response{
		status:   http.StatusOK,
		template: "about",
		data: about{
			Context: &Context{},
			Contributors: []*contributor{
				{
					Login:             "bobsomebody",
					ID:                0,
					AvatarURL:         "https://avatars1.githubusercontent.com/u/0?v=4",
					GravatarID:        "",
					URL:               "https://api.github.com/users/bobsomebody",
					HTMLURL:           "https://github.com/bobsomebody",
					FollowersURL:      "https://api.github.com/users/bobsomebody/followers",
					FollowingURL:      "https://api.github.com/users/bobsomebody/following{/other_user}",
					GistsURL:          "https://api.github.com/users/bobsomebody/gists{/gist_id}",
					StarredURL:        "https://api.github.com/users/bobsomebody/starred{/owner}{/repo}",
					SubscriptionsURL:  "https://api.github.com/users/bobsomebody/subscriptions",
					OrganizationsURL:  "https://api.github.com/users/bobsomebody/orgs",
					ReposURL:          "https://api.github.com/users/bobsomebody/repos",
					EventsURL:         "https://api.github.com/users/bobsomebody/events{/privacy}",
					ReceivedEventsURL: "https://api.github.com/users/bobsomebody/received_events",
					Type:              "User",
					SiteAdmin:         false,
					Contributions:     19,
				},
			},
			Onion: f.Onion,
		},
	}

	got := f.aboutHandler(httptest.NewRecorder(), req)

	if gotErr, wantErr := got.err, want.err; gotErr != wantErr {
		t.Fatalf("got error %q; want error %q", gotErr, wantErr)
	}

	httpmock.Reset()
}
