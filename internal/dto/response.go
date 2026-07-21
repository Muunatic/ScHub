package dto

type APIResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type APIResponseWithRedirect struct {
	Status      int    `json:"status"`
	Message     string `json:"message"`
	RedirectURL string `json:"redirectUrl,omitempty"`
}

type PostResponse struct {
	Status   int        `json:"status"`
	Message  string     `json:"message"`
	Platform string     `json:"platform"`
	Data     []PostData `json:"data"`
}

type PostData struct {
	ID         string  `json:"id"`
	ExternalID string  `json:"externalId"`
	Content    string  `json:"content"`
	Attachment string  `json:"attachment"`
	Comments   int     `json:"comments"`
	Likes      int     `json:"likes"`
	CreatedAt  string  `json:"createdAt"`
	AccountID  string  `json:"accountId"`
	Platform   string  `json:"platform,omitempty"`
	Thumbnail  *string `json:"thumbnail,omitempty"`
}

type CommentResponse struct {
	Status   int           `json:"status"`
	Message  string        `json:"message"`
	Platform string        `json:"platform"`
	Data     []CommentData `json:"data"`
}

type CommentData struct {
	ID             string `json:"id"`
	ExternalID     string `json:"externalId"`
	PostID         string `json:"postId"`
	Content        string `json:"content"`
	Likes          int    `json:"likes"`
	CreatedAt      string `json:"createdAt"`
	AccountID      string `json:"accountId"`
	Platform       string `json:"platform,omitempty"`
	PostContent    string `json:"postContent"`
	PostAttachment string `json:"postAttachment"`
}

type TikTokVideo struct {
	LikeCount     int    `json:"like_count"`
	ShareCount    int    `json:"share_count"`
	ShareURL      string `json:"share_url"`
	Title         string `json:"title"`
	CommentCount  int    `json:"comment_count"`
	CoverImageURL string `json:"cover_image_url"`
	CreateTime    int64  `json:"create_time"`
	ID            string `json:"id"`
}

type TwitterPostResponse struct {
	Data     []TwitterPostData `json:"data"`
	Includes TwitterIncludes   `json:"includes"`
}

type TwitterPostData struct {
	ID               string               `json:"id"`
	Text             string               `json:"text"`
	AuthorID         string               `json:"author_id"`
	CreatedAt        string               `json:"created_at"`
	ConversationID   string               `json:"conversation_id"`
	EditHistoryIDs   []string             `json:"edit_history_tweet_ids"`
	PublicMetrics    TwitterPublicMetrics `json:"public_metrics"`
	ReferencedTweets *[]TwitterReference  `json:"referenced_tweets,omitempty"`
	InReplyToUserID  *string              `json:"in_reply_to_user_id,omitempty"`
}

type TwitterPublicMetrics struct {
	RetweetCount    int `json:"retweet_count"`
	ReplyCount      int `json:"reply_count"`
	LikeCount       int `json:"like_count"`
	QuoteCount      int `json:"quote_count"`
	BookmarkCount   int `json:"bookmark_count"`
	ImpressionCount int `json:"impression_count"`
}

type TwitterReference struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type TwitterIncludes struct {
	Users []TwitterUser `json:"users"`
}

type TwitterUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type TwitterCommentResponse struct {
	Data []TwitterCommentData `json:"data"`
	Meta TwitterCommentMeta   `json:"meta"`
}

type TwitterCommentData struct {
	PublicMetrics   TwitterPublicMetrics `json:"public_metrics"`
	AuthorID        string               `json:"author_id"`
	ID              string               `json:"id"`
	EditHistoryIDs  []string             `json:"edit_history_tweet_ids"`
	Text            string               `json:"text"`
	InReplyToUserID string               `json:"in_reply_to_user_id"`
	CreatedAt       string               `json:"created_at"`
	ConversationID  string               `json:"conversation_id"`
}

type TwitterCommentMeta struct {
	NewestID    string `json:"newest_id"`
	OldestID    string `json:"oldest_id"`
	ResultCount int    `json:"result_count"`
}

type YouTubeVideoItems struct {
	Kind       string             `json:"kind"`
	Etag       string             `json:"etag"`
	RegionCode string             `json:"regionCode"`
	PageInfo   YouTubePageInfo    `json:"pageInfo"`
	Items      []YouTubeVideoItem `json:"items"`
}

type YouTubeVideoItem struct {
	Kind    string         `json:"kind"`
	Etag    string         `json:"etag"`
	ID      YouTubeVideoID `json:"id"`
	Snippet YouTubeSnippet `json:"snippet"`
}

type YouTubeVideoID struct {
	Kind    string `json:"kind"`
	VideoID string `json:"videoId"`
}

type YouTubeSnippet struct {
	PublishedAt          string            `json:"publishedAt"`
	ChannelID            string            `json:"channelId"`
	Title                string            `json:"title"`
	Description          string            `json:"description"`
	Thumbnails           YouTubeThumbnails `json:"thumbnails"`
	ChannelTitle         string            `json:"channelTitle"`
	LiveBroadcastContent string            `json:"liveBroadcastContent"`
	PublishTime          string            `json:"publishTime"`
}

type YouTubeThumbnails struct {
	Default YouTubeThumbnail `json:"default"`
	Medium  YouTubeThumbnail `json:"medium"`
	High    YouTubeThumbnail `json:"high"`
}

type YouTubeThumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type YouTubePageInfo struct {
	TotalResults   int `json:"totalResults"`
	ResultsPerPage int `json:"resultsPerPage"`
}

type YouTubeStatistics struct {
	Kind     string            `json:"kind"`
	Etag     string            `json:"etag"`
	Items    []YouTubeStatItem `json:"items"`
	PageInfo YouTubePageInfo   `json:"pageInfo"`
}

type YouTubeStatItem struct {
	Kind       string       `json:"kind"`
	Etag       string       `json:"etag"`
	ID         string       `json:"id"`
	Statistics YouTubeStats `json:"statistics"`
}

type YouTubeStats struct {
	ViewCount     string `json:"viewCount"`
	LikeCount     string `json:"likeCount"`
	DislikeCount  string `json:"dislikeCount"`
	FavoriteCount string `json:"favoriteCount"`
	CommentCount  string `json:"commentCount"`
}

type YouTubeCommentResponse struct {
	Kind    string               `json:"kind"`
	Etag    string               `json:"etag"`
	VideoID string               `json:"videoId"`
	Title   string               `json:"title"`
	Items   []YouTubeCommentItem `json:"items"`
}

type YouTubeCommentItem struct {
	Kind    string                `json:"kind"`
	Etag    string                `json:"etag"`
	ID      string                `json:"id"`
	Snippet YouTubeCommentSnippet `json:"snippet"`
}

type YouTubeCommentSnippet struct {
	ChannelID       string                 `json:"channelId"`
	VideoID         string                 `json:"videoId"`
	TopLevelComment YouTubeTopLevelComment `json:"topLevelComment"`
	CanReply        bool                   `json:"canReply"`
	TotalReplyCount int                    `json:"totalReplyCount"`
	IsPublic        bool                   `json:"isPublic"`
}

type YouTubeTopLevelComment struct {
	Kind    string                        `json:"kind"`
	Etag    string                        `json:"etag"`
	ID      string                        `json:"id"`
	Snippet YouTubeTopLevelCommentSnippet `json:"snippet"`
}

type YouTubeTopLevelCommentSnippet struct {
	ChannelID             string           `json:"channelId"`
	VideoID               string           `json:"videoId"`
	TextDisplay           string           `json:"textDisplay"`
	TextOriginal          string           `json:"textOriginal"`
	AuthorDisplayName     string           `json:"authorDisplayName"`
	AuthorProfileImageURL string           `json:"authorProfileImageUrl"`
	AuthorChannelURL      string           `json:"authorChannelUrl"`
	AuthorChannelID       YouTubeChannelID `json:"authorChannelId"`
	CanRate               bool             `json:"canRate"`
	ViewerRating          string           `json:"viewerRating"`
	LikeCount             int              `json:"likeCount"`
	PublishedAt           string           `json:"publishedAt"`
	UpdatedAt             string           `json:"updatedAt"`
}

type YouTubeChannelID struct {
	Value string `json:"value"`
}

type TikTokUploadInitResponse struct {
	Data TikTokUploadData `json:"data"`
}

type TikTokUploadData struct {
	UploadURL string `json:"upload_url"`
}

type TikTokVideoListResponse struct {
	Data TikTokVideoListData `json:"data"`
}

type TikTokVideoListData struct {
	Videos []TikTokVideo `json:"videos"`
}

type TikTokUserInfoResponse struct {
	Data TikTokUserInfoData `json:"data"`
}

type TikTokUserInfoData struct {
	User TikTokUserInfo `json:"user"`
}

type TikTokUserInfo struct {
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
	AvatarURL   string `json:"avatar_url"`
}

type YouTubeChannelResponse struct {
	Items []YouTubeChannelItem `json:"items"`
}

type YouTubeChannelItem struct {
	ID      string                `json:"id"`
	Snippet YouTubeChannelSnippet `json:"snippet"`
}

type YouTubeChannelSnippet struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Thumbnails  YouTubeThumbnails `json:"thumbnails"`
}

type YouTubeProfileResponse struct {
	Email string `json:"email"`
}
