package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	logger "github.com/dictybase/webhooks"
	zmq "github.com/pebbe/zmq4"

	"gopkg.in/codegangsta/cli.v0"
	"gopkg.in/gin-gonic/gin.v0"
)

type Content struct {
	Repository string
	User       string
	Ref        string
	Path       string
}

type webHookPush struct {
	After   string `json:"after"`
	Before  string `json:"before"`
	Commits []struct {
		Added  []interface{} `json:"added"`
		Author struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Username string `json:"username"`
		} `json:"author"`
		Committer struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Username string `json:"username"`
		} `json:"committer"`
		Distinct  bool          `json:"distinct"`
		Id        string        `json:"id"`
		Message   string        `json:"message"`
		Modified  []string      `json:"modified"`
		Removed   []interface{} `json:"removed"`
		Timestamp string        `json:"timestamp"`
		Url       string        `json:"url"`
	} `json:"commits"`
	Compare    string `json:"compare"`
	Created    bool   `json:"created"`
	Deleted    bool   `json:"deleted"`
	Forced     bool   `json:"forced"`
	HeadCommit struct {
		Added  []string `json:"added"`
		Author struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Username string `json:"username"`
		} `json:"author"`
		Committer struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Username string `json:"username"`
		} `json:"committer"`
		Distinct  bool     `json:"distinct"`
		Id        string   `json:"id"`
		Message   string   `json:"message"`
		Modified  []string `json:"modified"`
		Removed   []string `json:"removed"`
		Timestamp string   `json:"timestamp"`
		Url       string   `json:"url"`
	} `json:"head_commit"`
	Pusher struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"pusher"`
	Ref        string `json:"ref"`
	Repository struct {
		ArchiveUrl       string      `json:"archive_url"`
		AssigneesUrl     string      `json:"assignees_url"`
		BlobsUrl         string      `json:"blobs_url"`
		BranchesUrl      string      `json:"branches_url"`
		CloneUrl         string      `json:"clone_url"`
		CollaboratorsUrl string      `json:"collaborators_url"`
		CommentsUrl      string      `json:"comments_url"`
		CommitsUrl       string      `json:"commits_url"`
		CompareUrl       string      `json:"compare_url"`
		ContentsUrl      string      `json:"contents_url"`
		ContributorsUrl  string      `json:"contributors_url"`
		CreatedAt        int64       `json:"created_at"`
		DefaultBranch    string      `json:"default_branch"`
		Description      string      `json:"description"`
		DownloadsUrl     string      `json:"downloads_url"`
		EventsUrl        string      `json:"events_url"`
		Fork             bool        `json:"fork"`
		Forks            int64       `json:"forks"`
		ForksCount       int64       `json:"forks_count"`
		ForksUrl         string      `json:"forks_url"`
		FullName         string      `json:"full_name"`
		GitCommitsUrl    string      `json:"git_commits_url"`
		GitRefsUrl       string      `json:"git_refs_url"`
		GitTagsUrl       string      `json:"git_tags_url"`
		GitUrl           string      `json:"git_url"`
		HasDownloads     bool        `json:"has_downloads"`
		HasIssues        bool        `json:"has_issues"`
		HasWiki          bool        `json:"has_wiki"`
		Homepage         interface{} `json:"homepage"`
		HooksUrl         string      `json:"hooks_url"`
		HtmlUrl          string      `json:"html_url"`
		Id               int64       `json:"id"`
		IssueCommentUrl  string      `json:"issue_comment_url"`
		IssueEventsUrl   string      `json:"issue_events_url"`
		IssuesUrl        string      `json:"issues_url"`
		KeysUrl          string      `json:"keys_url"`
		LabelsUrl        string      `json:"labels_url"`
		Language         interface{} `json:"language"`
		LanguagesUrl     string      `json:"languages_url"`
		MasterBranch     string      `json:"master_branch"`
		MergesUrl        string      `json:"merges_url"`
		MilestonesUrl    string      `json:"milestones_url"`
		MirrorUrl        interface{} `json:"mirror_url"`
		Name             string      `json:"name"`
		NotificationsUrl string      `json:"notifications_url"`
		OpenIssues       int64       `json:"open_issues"`
		OpenIssuesCount  int64       `json:"open_issues_count"`
		Owner            struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"owner"`
		Private         bool   `json:"private"`
		PullsUrl        string `json:"pulls_url"`
		PushedAt        int64  `json:"pushed_at"`
		ReleasesUrl     string `json:"releases_url"`
		Size            int64  `json:"size"`
		SshUrl          string `json:"ssh_url"`
		Stargazers      int64  `json:"stargazers"`
		StargazersCount int64  `json:"stargazers_count"`
		StargazersUrl   string `json:"stargazers_url"`
		StatusesUrl     string `json:"statuses_url"`
		SubscribersUrl  string `json:"subscribers_url"`
		SubscriptionUrl string `json:"subscription_url"`
		SvnUrl          string `json:"svn_url"`
		TagsUrl         string `json:"tags_url"`
		TeamsUrl        string `json:"teams_url"`
		TreesUrl        string `json:"trees_url"`
		UpdatedAt       string `json:"updated_at"`
		Url             string `json:"url"`
		Watchers        int64  `json:"watchers"`
		WatchersCount   int64  `json:"watchers_count"`
	} `json:"repository"`
}

var log *logger.Logger

func init() {
	log = logger.NewLogger(os.Stderr, logger.INFO)
}

func main() {
	app := cli.NewApp()
	app.Name = "push2zmq"
	app.Usage = "Relay github push event through zeromq"
	app.Version = "1.0.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "listen, l",
			Usage: "The host and port the server will run, default is :9090",
			Value: ":9090",
		},
		cli.StringFlag{
			Name:  "broadcast, b",
			Usage: "The host and port where the zeromq event will be send, default is *:8989",
			Value: "*:8989",
		},
		cli.StringFlag{
			Name:  "token, t",
			Usage: "The security token to verify the webhook, mandatory. Alteranatively, set the WEBHOOK_TOKEN which takes precedence",
		},
	}
	app.Action = runServer
	app.Run(os.Args)
}

func runServer(c *cli.Context) {
	// check for secret token
	var token string
	if token = c.String("token"); len(token) == 0 {
		if token = os.Getenv("WEBHOOK_TOKEN"); len(token) == 0 {
			log.Fatal("security token is not set")
		}
	}

	handler := getHttpHandler(c, token)
	log.Infof("starting webserver at %s\n", c.String("listen"))
	http.ListenAndServe(c.String("listen"), handler)
}

func getHttpHandler(c *cli.Context, token string) http.Handler {
	zmq := setupZmq(c)
	// setup routes
	resource := &webHookResource{zmq}
	r := gin.Default()
	auth := r.Group("/webhook")
	auth.POST("/send", SecureWebhook(token), resource.SendToZmq)
	return r
}

func setupZmq(c *cli.Context) *zmq.Socket {
	ctx, err := zmq.NewContext()
	if err != nil {
		log.Fatal(err)
	}
	socket, err := ctx.NewSocket(zmq.PUSH)
	socket.Bind(fmt.Sprintf("tcp://%s", c.String("broadcast")))
	return socket
}

func validateToken(messageMAC, body, token []byte) bool {
	mac := hmac.New(sha1.New, token)
	mac.Write(body)
	expected := "sha1=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal(messageMAC, []byte(expected))
}

func SecureWebhook(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validating the hash signature send with the webhook
		// It should match with the given secret token
		messageMAC := c.Request.Header.Get("X-Hub-Signature")
		if len(messageMAC) == 0 {
			log.Error("no digest given in the webhook")
			c.Fail(http.StatusBadRequest, errors.New("no digest given in the webhook"))
			return
		}
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Error(err)
			c.Fail(http.StatusBadRequest, err)
			return
		}
		//make a copy of request body for subsequent use
		rd := ioutil.NopCloser(bytes.NewBuffer(body))
		if !validateToken([]byte(messageMAC), body, []byte(token)) {
			log.Error("unable to validate the hash signatur/e")
			c.Fail(http.StatusBadRequest, errors.New("unable to validate the hash signature"))
			return
		}
		// assign the copy to the request body since the original will get drained by
		// the previous process
		c.Request.Body = rd
		// Validation ends
		c.Next()
	}
}

type webHookResource struct {
	socket *zmq.Socket
}

func (w *webHookResource) SendToZmq(c *gin.Context) {

	var wh webHookPush
	if !c.Bind(&wh) {
		log.Error("could not parse json request")
		c.Fail(http.StatusBadRequest, errors.New("could not parse json request"))
		return
	}

	// figure out the added/modified file(s) and
	// copy them to remote server
	for _, a := range wh.HeadCommit.Added {
		ac := Content{
			Repository: wh.Repository.Name,
			User:       wh.Repository.Owner.Name,
			Ref:        wh.HeadCommit.Id,
			Path:       a,
		}
		b, err := json.Marshal(ac)
		if err != nil {
			log.Error(err)
			c.Fail(http.StatusBadRequest, err)
			return
		}
		log.Infof("going to send %s\n", string(b))
		_, err = w.socket.Send(string(b), 0)
		if err != nil {
			log.Errorf("unable to send data %s through zeromq error:%s\n", string(b), err)
			c.Fail(http.StatusBadRequest, err)
			return
		}
		log.Infof("send data %s through zeromq", string(b))
	}
	for _, m := range wh.HeadCommit.Modified {
		ac := Content{
			Repository: wh.Repository.Name,
			User:       wh.Repository.Owner.Name,
			Ref:        wh.HeadCommit.Id,
			Path:       m,
		}
		b, err := json.Marshal(ac)
		if err != nil {
			log.Error(err)
			c.Fail(http.StatusBadRequest, err)
			return
		}
		log.Infof("going to send %s\n", string(b))
		_, err = w.socket.Send(string(b), 0)
		if err != nil {
			log.Errorf("unable to send data %s through zeromq error:%s\n", string(b), err)
			c.Fail(http.StatusBadRequest, err)
			return
		}
		log.Infof("send data %s through zeromq", string(b))
	}
	c.String(http.StatusOK, "send all urls to the receiver")
	return
}
