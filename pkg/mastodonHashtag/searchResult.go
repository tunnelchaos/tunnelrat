package mastodonHashtag

import "time"

type searchResult struct {
	Statuses []Statuses `json:"statuses"`
}
type Account struct {
	ID              string    `json:"id"`
	Username        string    `json:"username"`
	Acct            string    `json:"acct"`
	DisplayName     string    `json:"display_name"`
	Locked          bool      `json:"locked"`
	Bot             bool      `json:"bot"`
	Discoverable    bool      `json:"discoverable"`
	Indexable       bool      `json:"indexable"`
	Group           bool      `json:"group"`
	CreatedAt       time.Time `json:"created_at"`
	Note            string    `json:"note"`
	URL             string    `json:"url"`
	URI             string    `json:"uri"`
	Avatar          string    `json:"avatar"`
	AvatarStatic    string    `json:"avatar_static"`
	Header          string    `json:"header"`
	HeaderStatic    string    `json:"header_static"`
	FollowersCount  int       `json:"followers_count"`
	FollowingCount  int       `json:"following_count"`
	StatusesCount   int       `json:"statuses_count"`
	LastStatusAt    string    `json:"last_status_at"`
	HideCollections bool      `json:"hide_collections"`
	Emojis          []any     `json:"emojis"`
	Fields          []any     `json:"fields"`
}
type Mentions struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	URL      string `json:"url"`
	Acct     string `json:"acct"`
}
type Tags struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
type Application struct {
	Name    string `json:"name"`
	Website string `json:"website"`
}
type Fields struct {
	Name       string    `json:"name"`
	Value      string    `json:"value"`
	VerifiedAt time.Time `json:"verified_at"`
}
type Statuses struct {
	ID                 string      `json:"id"`
	CreatedAt          time.Time   `json:"created_at"`
	InReplyToID        string      `json:"in_reply_to_id"`
	InReplyToAccountID string      `json:"in_reply_to_account_id"`
	Sensitive          bool        `json:"sensitive"`
	SpoilerText        string      `json:"spoiler_text"`
	Visibility         string      `json:"visibility"`
	Language           string      `json:"language"`
	URI                string      `json:"uri"`
	URL                string      `json:"url"`
	RepliesCount       int         `json:"replies_count"`
	ReblogsCount       int         `json:"reblogs_count"`
	FavouritesCount    int         `json:"favourites_count"`
	EditedAt           any         `json:"edited_at"`
	Favourited         bool        `json:"favourited"`
	Reblogged          bool        `json:"reblogged"`
	Muted              bool        `json:"muted"`
	Bookmarked         bool        `json:"bookmarked"`
	Content            string      `json:"content"`
	Filtered           []any       `json:"filtered"`
	Reblog             any         `json:"reblog"`
	Account            Account     `json:"account,omitempty"`
	MediaAttachments   []any       `json:"media_attachments"`
	Mentions           []Mentions  `json:"mentions"`
	Tags               []Tags      `json:"tags"`
	Emojis             []any       `json:"emojis"`
	Card               any         `json:"card"`
	Poll               any         `json:"poll"`
	Application        Application `json:"application,omitempty"`
}
