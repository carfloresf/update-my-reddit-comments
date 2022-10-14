package pushreddit

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
)

type Subreddit struct {
	Author                   string        `json:"author"`
	AuthorFlairCSSClass      interface{}   `json:"author_flair_css_class"`
	AuthorFlairRichtext      []interface{} `json:"author_flair_richtext,omitempty"`
	AuthorFlairText          interface{}   `json:"author_flair_text"`
	AuthorFlairType          string        `json:"author_flair_type,omitempty"`
	BrandSafe                bool          `json:"brand_safe"`
	CanModPost               bool          `json:"can_mod_post"`
	ContestMode              bool          `json:"contest_mode"`
	CreatedUtc               int           `json:"created_utc"`
	Domain                   string        `json:"domain"`
	FullLink                 string        `json:"full_link"`
	Gilded                   int           `json:"gilded"`
	ID                       string        `json:"id"`
	IsCrosspostable          bool          `json:"is_crosspostable"`
	IsOriginalContent        bool          `json:"is_original_content"`
	IsRedditMediaDomain      bool          `json:"is_reddit_media_domain"`
	IsSelf                   bool          `json:"is_self"`
	IsVideo                  bool          `json:"is_video"`
	LinkFlairRichtext        []interface{} `json:"link_flair_richtext"`
	LinkFlairTextColor       string        `json:"link_flair_text_color"`
	LinkFlairType            string        `json:"link_flair_type"`
	Locked                   bool          `json:"locked"`
	NoFollow                 bool          `json:"no_follow"`
	NumComments              int           `json:"num_comments"`
	NumCrossposts            int           `json:"num_crossposts"`
	Over18                   bool          `json:"over_18"`
	ParentWhitelistStatus    string        `json:"parent_whitelist_status"`
	Permalink                string        `json:"permalink"`
	Pinned                   bool          `json:"pinned"`
	RetrievedOn              int           `json:"retrieved_on"`
	RteMode                  string        `json:"rte_mode"`
	Score                    int           `json:"score"`
	Selftext                 string        `json:"selftext"`
	SendReplies              bool          `json:"send_replies"`
	Spoiler                  bool          `json:"spoiler"`
	Stickied                 bool          `json:"stickied"`
	Subreddit                string        `json:"subreddit"`
	SubredditID              string        `json:"subreddit_id"`
	SubredditSubscribers     int           `json:"subreddit_subscribers"`
	SubredditType            string        `json:"subreddit_type"`
	Thumbnail                string        `json:"thumbnail"`
	Title                    string        `json:"title"`
	URL                      string        `json:"url"`
	WhitelistStatus          string        `json:"whitelist_status"`
	LinkFlairBackgroundColor string        `json:"link_flair_background_color,omitempty"`
	LinkFlairCSSClass        string        `json:"link_flair_css_class,omitempty"`
	LinkFlairTemplateID      string        `json:"link_flair_template_id,omitempty"`
	LinkFlairText            string        `json:"link_flair_text,omitempty"`
	PostHint                 string        `json:"post_hint,omitempty"`
	Preview                  struct {
		Enabled bool `json:"enabled"`
		Images  []struct {
			ID          string `json:"id"`
			Resolutions []struct {
				Height int    `json:"height"`
				URL    string `json:"url"`
				Width  int    `json:"width"`
			} `json:"resolutions"`
			Source struct {
				Height int    `json:"height"`
				URL    string `json:"url"`
				Width  int    `json:"width"`
			} `json:"source"`
			Variants struct {
			} `json:"variants"`
		} `json:"images"`
	} `json:"preview,omitempty"`
	AuthorFlairBackgroundColor string        `json:"author_flair_background_color,omitempty"`
	AuthorFlairTextColor       string        `json:"author_flair_text_color,omitempty"`
	ThumbnailHeight            int           `json:"thumbnail_height,omitempty"`
	ThumbnailWidth             int           `json:"thumbnail_width,omitempty"`
	Edited                     int           `json:"edited,omitempty"`
	PreviousVisits             []interface{} `json:"previous_visits,omitempty"`
	AuthorCakeday              bool          `json:"author_cakeday,omitempty"`
	Media                      struct {
		Oembed struct {
			Description     string `json:"description"`
			Height          int    `json:"height"`
			HTML            string `json:"html"`
			ProviderName    string `json:"provider_name"`
			ProviderURL     string `json:"provider_url"`
			ThumbnailHeight int    `json:"thumbnail_height"`
			ThumbnailURL    string `json:"thumbnail_url"`
			ThumbnailWidth  int    `json:"thumbnail_width"`
			Title           string `json:"title"`
			Type            string `json:"type"`
			Version         string `json:"version"`
			Width           int    `json:"width"`
		} `json:"oembed"`
		Type string `json:"type"`
	} `json:"media,omitempty"`
	MediaEmbed struct {
		Content        string `json:"content"`
		Height         int    `json:"height"`
		MediaDomainURL string `json:"media_domain_url"`
		Scrolling      bool   `json:"scrolling"`
		Width          int    `json:"width"`
	} `json:"media_embed,omitempty"`
	SecureMedia struct {
		Oembed struct {
			Description     string `json:"description"`
			Height          int    `json:"height"`
			HTML            string `json:"html"`
			ProviderName    string `json:"provider_name"`
			ProviderURL     string `json:"provider_url"`
			ThumbnailHeight int    `json:"thumbnail_height"`
			ThumbnailURL    string `json:"thumbnail_url"`
			ThumbnailWidth  int    `json:"thumbnail_width"`
			Title           string `json:"title"`
			Type            string `json:"type"`
			Version         string `json:"version"`
			Width           int    `json:"width"`
		} `json:"oembed"`
		Type string `json:"type"`
	} `json:"secure_media,omitempty"`
	SecureMediaEmbed struct {
		Content        string `json:"content"`
		Height         int    `json:"height"`
		MediaDomainURL string `json:"media_domain_url"`
		Scrolling      bool   `json:"scrolling"`
		Width          int    `json:"width"`
	} `json:"secure_media_embed,omitempty"`
	Distinguished string `json:"distinguished,omitempty"`
	SuggestedSort string `json:"suggested_sort,omitempty"`
}

type Data struct {
	Data []Subreddit `json:"data"`
}

type Client struct {
	h *retryablehttp.Client
}

func NewClient() *Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 10
	retryClient.RetryWaitMin = 150 * time.Second
	retryClient.RetryWaitMax = 300 * time.Second
	retryClient.Backoff = retryablehttp.DefaultBackoff
	retryClient.HTTPClient.Timeout = 30 * time.Second
	retryClient.CheckRetry = retryablehttp.DefaultRetryPolicy

	return &Client{
		retryClient,
	}
}

func (c *Client) GetPostsSubreddit(subreddit string, after time.Time, before time.Time, size int) (Data, error) {
	requestURL := fmt.Sprintf("https://api.pushshift.io/reddit/search/submission/?subreddit=%s&sort=desc&sort_type=created_utc&after=%d&before=%d&size=%d", subreddit, after.Unix(), before.Unix(), size)

	response, err := c.h.Get(requestURL)
	if err != nil {
		log.Errorf("error getting posts: %s", err)

		return Data{}, err
	}

	if response.StatusCode != http.StatusOK {
		log.Errorf("error getting posts: %s", response.Status)

		return Data{}, fmt.Errorf("error getting posts: %s", response.Status) //nolint:goerr113
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Printf("error closing response body: %s\n", err)
		}
	}()

	// read json http response
	jsonDataFromHttp, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	var data Data

	err = json.Unmarshal(jsonDataFromHttp, &data) // here!
	if err != nil {
		panic(err)
	}

	return data, nil
}
