package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/muunatic/schub/internal/config"
	"github.com/muunatic/schub/internal/dto"
	"github.com/muunatic/schub/internal/models"
	"github.com/muunatic/schub/internal/repository"
)

type PlatformService struct {
	accountRepo *repository.AccountRepository
	postRepo    *repository.PostRepository
	commentRepo *repository.CommentRepository
	cfg         *config.Config
}

func NewPlatformService(
	accountRepo *repository.AccountRepository,
	postRepo *repository.PostRepository,
	commentRepo *repository.CommentRepository,
	cfg *config.Config,
) *PlatformService {
	return &PlatformService{
		accountRepo: accountRepo,
		postRepo:    postRepo,
		commentRepo: commentRepo,
		cfg:         cfg,
	}
}

func (s *PlatformService) FetchYouTubePosts(ctx context.Context, userID string) ([]dto.PostData, error) {
	account, err := s.accountRepo.FindByUserIDAndPlatform(ctx, userID, models.PlatformYouTube)
	if err != nil {
		return nil, err
	}

	channelResp, err := s.doYouTubeGet(account.AccessToken, "https://www.googleapis.com/youtube/v3/channels?part=snippet&mine=true")
	if err != nil {
		return nil, err
	}

	var channelData dto.YouTubeChannelResponse
	if err := json.Unmarshal(channelResp, &channelData); err != nil {
		return nil, err
	}

	if len(channelData.Items) == 0 {
		return nil, fmt.Errorf("channel not found")
	}

	channelID := channelData.Items[0].ID

	videoResp, err := s.doYouTubeGet(account.AccessToken, fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?part=snippet&channelId=%s&order=date&maxResults=10", channelID))
	if err != nil {
		return nil, err
	}

	var videoData dto.YouTubeVideoItems
	if err := json.Unmarshal(videoResp, &videoData); err != nil {
		return nil, err
	}

	var items []struct {
		ExternalID string
		Thumbnail  string
		Content    string
		Attachment string
		Comments   int
		Likes      int
		CreatedAt  time.Time
		AccountID  string
	}

	if len(videoData.Items) > 0 {
		videoIDs := make([]string, 0)
		for _, v := range videoData.Items {
			if v.ID.Kind == "youtube#video" {
				videoIDs = append(videoIDs, v.ID.VideoID)
			}
		}

		if len(videoIDs) > 0 {
			statsResp, err := s.doYouTubeGet(account.AccessToken, fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=statistics&id=%s", strings.Join(videoIDs, ",")))
			if err == nil {
				var statsData dto.YouTubeStatistics
				if err := json.Unmarshal(statsResp, &statsData); err == nil {
					for _, item := range videoData.Items {
						if item.ID.Kind != "youtube#video" {
							continue
						}
						if item.Snippet.ChannelID != *account.Username {
							continue
						}

						var videoStats *dto.YouTubeStats
						for _, stat := range statsData.Items {
							if stat.ID == item.ID.VideoID {
								videoStats = &stat.Statistics
								break
							}
						}

						likes := 0
						comments := 0
						if videoStats != nil {
							likes, _ = strconv.Atoi(videoStats.LikeCount)
							comments, _ = strconv.Atoi(videoStats.CommentCount)
						}

						createdAt, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)

						entry := struct {
							ExternalID string
							Thumbnail  string
							Content    string
							Attachment string
							Comments   int
							Likes      int
							CreatedAt  time.Time
							AccountID  string
						}{
							ExternalID: item.ID.VideoID,
							Thumbnail:  item.Snippet.Thumbnails.Medium.URL,
							Content:    item.Snippet.Title,
							Attachment: fmt.Sprintf("https://youtu.be/%s", item.ID.VideoID),
							Comments:   comments,
							Likes:      likes,
							CreatedAt:  createdAt,
							AccountID:  account.ID,
						}
						items = append(items, entry)
					}
				}
			}
		}
	}

	for _, item := range items {
		post := &models.Post{
			ExternalID: item.ExternalID,
			Content:    item.Content,
			Attachment: item.Attachment,
			Likes:      item.Likes,
			CreatedAt:  item.CreatedAt,
			AccountID:  item.AccountID,
		}
		s.postRepo.Upsert(ctx, post)
	}

	posts, err := s.postRepo.FindByAccountID(ctx, account.ID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.PostData, 0, len(posts))
	for _, p := range posts {
		postData := dto.PostData{
			ID:         p.ID,
			ExternalID: p.ExternalID,
			Content:    p.Content,
			Attachment: p.Attachment,
			Likes:      p.Likes,
			CreatedAt:  p.CreatedAt.Format(time.RFC3339),
			AccountID:  p.AccountID,
		}

		for _, item := range items {
			if item.ExternalID == p.ExternalID {
				postData.Comments = item.Comments
				postData.Thumbnail = &item.Thumbnail
				break
			}
		}

		result = append(result, postData)
	}

	return result, nil
}

func (s *PlatformService) UploadYouTubeVideo(ctx context.Context, userID, title string, fileData []byte) (map[string]interface{}, error) {
	account, err := s.accountRepo.FindByUserIDAndPlatform(ctx, userID, models.PlatformYouTube)
	if err != nil {
		return nil, err
	}

	initBody := map[string]interface{}{
		"snippet": map[string]interface{}{
			"title": title,
		},
		"status": map[string]interface{}{
			"privacyStatus": "unlisted",
		},
	}
	initJSON, _ := json.Marshal(initBody)

	req, err := http.NewRequest("POST", "https://www.googleapis.com/upload/youtube/v3/videos?uploadType=resumable&part=snippet,status", bytes.NewReader(initJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+account.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Upload-Content-Type", "video/mp4")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upload init failed")
	}

	uploadURL := resp.Header.Get("Location")

	uploadReq, err := http.NewRequest("PUT", uploadURL, bytes.NewReader(fileData))
	if err != nil {
		return nil, err
	}
	uploadReq.Header.Set("Content-Type", "video/mp4")
	uploadReq.Header.Set("Content-Length", strconv.Itoa(len(fileData)))

	uploadResp, err := http.DefaultClient.Do(uploadReq)
	if err != nil {
		return nil, err
	}
	defer uploadResp.Body.Close()

	var uploadResult map[string]interface{}
	if err := json.NewDecoder(uploadResp.Body).Decode(&uploadResult); err != nil {
		return nil, err
	}

	return uploadResult, nil
}

func (s *PlatformService) FetchTikTokPosts(ctx context.Context, userID string) ([]dto.PostData, error) {
	account, err := s.accountRepo.FindByUserIDAndPlatform(ctx, userID, models.PlatformTikTok)
	if err != nil {
		return nil, err
	}

	body := strings.NewReader(`{"max_count": 20}`)
	req, err := http.NewRequest("POST", "https://open.tiktokapis.com/v2/video/list/?fields=id,title,cover_image_url,like_count,comment_count,share_count,create_time,id,share_url", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+account.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var videoListResp dto.TikTokVideoListResponse
	if err := json.NewDecoder(resp.Body).Decode(&videoListResp); err != nil {
		return nil, err
	}

	var items []struct {
		ExternalID string
		Thumbnail  string
		Content    string
		Attachment string
		Comments   int
		Likes      int
		CreatedAt  time.Time
		AccountID  string
	}

	for _, video := range videoListResp.Data.Videos {
		entry := struct {
			ExternalID string
			Thumbnail  string
			Content    string
			Attachment string
			Comments   int
			Likes      int
			CreatedAt  time.Time
			AccountID  string
		}{
			ExternalID: video.ID,
			Thumbnail:  video.CoverImageURL,
			Content:    video.Title,
			Attachment: video.ShareURL,
			Likes:      video.LikeCount,
			Comments:   video.CommentCount,
			CreatedAt:  time.Unix(video.CreateTime, 0),
			AccountID:  account.ID,
		}
		items = append(items, entry)
	}

	for _, item := range items {
		post := &models.Post{
			ExternalID: item.ExternalID,
			Content:    item.Content,
			Attachment: item.Attachment,
			Likes:      item.Likes,
			CreatedAt:  item.CreatedAt,
			AccountID:  item.AccountID,
		}
		s.postRepo.Upsert(ctx, post)
	}

	posts, err := s.postRepo.FindByAccountID(ctx, account.ID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.PostData, 0, len(posts))
	for _, p := range posts {
		postData := dto.PostData{
			ID:         p.ID,
			ExternalID: p.ExternalID,
			Content:    p.Content,
			Attachment: p.Attachment,
			Likes:      p.Likes,
			CreatedAt:  p.CreatedAt.Format(time.RFC3339),
			AccountID:  p.AccountID,
		}

		for _, item := range items {
			if item.ExternalID == p.ExternalID {
				postData.Comments = item.Comments
				postData.Thumbnail = &item.Thumbnail
				break
			}
		}

		result = append(result, postData)
	}

	return result, nil
}

func (s *PlatformService) UploadTikTokVideo(ctx context.Context, userID, title string, fileData []byte, fileSize int64) (map[string]interface{}, error) {
	account, err := s.accountRepo.FindByUserIDAndPlatform(ctx, userID, models.PlatformTikTok)
	if err != nil {
		return nil, err
	}

	validChunkSizes := []int64{52428800, 10485760}
	var chunkSize int64
	for _, size := range validChunkSizes {
		if fileSize%size == 0 {
			chunkSize = size
			break
		}
	}
	if chunkSize == 0 {
		chunkSize = fileSize
	}

	totalChunkCount := (fileSize + chunkSize - 1) / chunkSize

	initBody := map[string]interface{}{
		"post_info": map[string]interface{}{
			"privacy_level": "SELF_ONLY",
			"title":         title,
		},
		"source_info": map[string]interface{}{
			"source":            "FILE_UPLOAD",
			"video_size":        fileSize,
			"chunk_size":        chunkSize,
			"total_chunk_count": totalChunkCount,
		},
	}
	initJSON, _ := json.Marshal(initBody)

	req, err := http.NewRequest("POST", "https://open.tiktokapis.com/v2/post/publish/video/init/", bytes.NewReader(initJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+account.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var initResp dto.TikTokUploadInitResponse
	if err := json.NewDecoder(resp.Body).Decode(&initResp); err != nil {
		return nil, err
	}

	uploadReq, err := http.NewRequest("PUT", initResp.Data.UploadURL, bytes.NewReader(fileData))
	if err != nil {
		return nil, err
	}
	uploadReq.Header.Set("Content-Type", "video/mp4")
	uploadReq.Header.Set("Content-Length", strconv.FormatInt(fileSize, 10))
	uploadReq.Header.Set("Content-Range", fmt.Sprintf("bytes 0-%d/%d", fileSize-1, fileSize))

	uploadResp, err := http.DefaultClient.Do(uploadReq)
	if err != nil {
		return nil, err
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tiktok upload failed")
	}

	return map[string]interface{}{
		"upload_url": initResp.Data.UploadURL,
	}, nil
}

func (s *PlatformService) FetchTwitterPosts(ctx context.Context, userID string) ([]dto.PostData, error) {
	account, err := s.accountRepo.FindByUserIDAndPlatform(ctx, userID, models.PlatformTwitter)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.twitter.com/2/users/%s/tweets?tweet.fields=public_metrics,conversation_id,in_reply_to_user_id,created_at,referenced_tweets&expansions=author_id,in_reply_to_user_id,referenced_tweets.id&user.fields=name,username", *account.Username)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+account.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result dto.TwitterPostResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var items []struct {
		ExternalID string
		Content    string
		Attachment string
		Likes      int
		CreatedAt  time.Time
		AccountID  string
	}

	for _, tweet := range result.Data {
		if tweet.ReferencedTweets != nil && len(*tweet.ReferencedTweets) > 0 {
			continue
		}
		if tweet.InReplyToUserID != nil && *tweet.InReplyToUserID != "" {
			continue
		}

		var username string
		for _, user := range result.Includes.Users {
			if user.ID == *account.Username {
				username = user.Username
				break
			}
		}

		createdAt, _ := time.Parse(time.RFC3339, tweet.CreatedAt)

		entry := struct {
			ExternalID string
			Content    string
			Attachment string
			Likes      int
			CreatedAt  time.Time
			AccountID  string
		}{
			ExternalID: tweet.ID,
			Content:    tweet.Text,
			Attachment: fmt.Sprintf("https://x.com/%s/status/%s", username, tweet.ID),
			Likes:      tweet.PublicMetrics.LikeCount,
			CreatedAt:  createdAt,
			AccountID:  account.ID,
		}
		items = append(items, entry)
	}

	for _, item := range items {
		post := &models.Post{
			ExternalID: item.ExternalID,
			Content:    item.Content,
			Attachment: item.Attachment,
			Likes:      item.Likes,
			CreatedAt:  item.CreatedAt,
			AccountID:  item.AccountID,
		}
		s.postRepo.Upsert(ctx, post)
	}

	posts, err := s.postRepo.FindByAccountID(ctx, account.ID)
	if err != nil {
		return nil, err
	}

	result2 := make([]dto.PostData, 0, len(posts))
	for _, p := range posts {
		result2 = append(result2, dto.PostData{
			ID:         p.ID,
			ExternalID: p.ExternalID,
			Content:    p.Content,
			Attachment: p.Attachment,
			Likes:      p.Likes,
			CreatedAt:  p.CreatedAt.Format(time.RFC3339),
			AccountID:  p.AccountID,
		})
	}

	return result2, nil
}

func (s *PlatformService) UploadTwitterVideo(ctx context.Context, userID, title string, fileData []byte, fileMime string) (map[string]interface{}, error) {
	account, err := s.accountRepo.FindByUserIDAndPlatform(ctx, userID, models.PlatformTwitter)
	if err != nil {
		return nil, err
	}

	if account.AccessKey == nil || account.AccessSecretKey == nil {
		return nil, fmt.Errorf("twitter OAuth 1.0 credentials not found")
	}

	mediaSize := len(fileData)
	uploadURL := "https://upload.twitter.com/1.1/media/upload.json"

	oauth := NewOAuth1(s.cfg.TwitterAppKey, s.cfg.TwitterAppSecretKey)
	initData := map[string]string{
		"command":        "INIT",
		"total_bytes":    strconv.Itoa(mediaSize),
		"media_type":     fileMime,
		"media_category": "tweet_video",
	}
	initHeader := oauth.AuthorizeWithToken(uploadURL, "POST", initData, *account.AccessKey, *account.AccessSecretKey)

	form := url.Values{}
	for k, v := range initData {
		form.Set(k, v)
	}

	req, _ := http.NewRequest("POST", uploadURL, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for k, v := range initHeader {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var initResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&initResult)

	mediaID, ok := initResult["media_id_string"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to get media ID")
	}

	chunkSize := 5 * 1024 * 1024
	totalChunks := (mediaSize + chunkSize - 1) / chunkSize

	for i := 0; i < totalChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > mediaSize {
			end = mediaSize
		}
		chunk := fileData[start:end]

		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		w.WriteField("command", "APPEND")
		w.WriteField("media_id", mediaID)
		w.WriteField("segment_index", strconv.Itoa(i))

		part, _ := w.CreateFormFile("media", fmt.Sprintf("chunk%d.mp4", i))
		part.Write(chunk)
		w.Close()

		appendHeader := oauth.AuthorizeWithToken(uploadURL, "POST", nil, *account.AccessKey, *account.AccessSecretKey)

		appendReq, _ := http.NewRequest("POST", uploadURL, &buf)
		appendReq.Header.Set("Content-Type", w.FormDataContentType())
		for k, v := range appendHeader {
			appendReq.Header.Set(k, v)
		}

		appendResp, err := http.DefaultClient.Do(appendReq)
		if err != nil {
			return nil, err
		}
		appendResp.Body.Close()

		if appendResp.StatusCode != http.StatusOK && appendResp.StatusCode != http.StatusNoContent {
			bodyBytes, _ := io.ReadAll(appendResp.Body)
			return nil, fmt.Errorf("append failed: %s", string(bodyBytes))
		}
	}

	finalizeData := map[string]string{
		"command":  "FINALIZE",
		"media_id": mediaID,
	}
	finalizeHeader := oauth.AuthorizeWithToken(uploadURL, "POST", finalizeData, *account.AccessKey, *account.AccessSecretKey)

	finalForm := url.Values{}
	finalForm.Set("command", "FINALIZE")
	finalForm.Set("media_id", mediaID)

	finalReq, _ := http.NewRequest("POST", uploadURL, strings.NewReader(finalForm.Encode()))
	finalReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for k, v := range finalizeHeader {
		finalReq.Header.Set(k, v)
	}

	finalResp, err := http.DefaultClient.Do(finalReq)
	if err != nil {
		return nil, err
	}
	defer finalResp.Body.Close()

	var finalizeResult map[string]interface{}
	json.NewDecoder(finalResp.Body).Decode(&finalizeResult)

	checkAfterSecs := 5
	if processingInfo, ok := finalizeResult["processing_info"].(map[string]interface{}); ok {
		if cas, ok := processingInfo["check_after_secs"].(float64); ok {
			checkAfterSecs = int(cas)
		}
	}

	time.Sleep(time.Duration(5+checkAfterSecs) * time.Second)

	tweetBody := map[string]interface{}{
		"text": title,
		"media": map[string]interface{}{
			"media_ids": []string{mediaID},
		},
	}
	tweetJSON, _ := json.Marshal(tweetBody)

	tweetReq, _ := http.NewRequest("POST", "https://api.twitter.com/2/tweets", bytes.NewReader(tweetJSON))
	tweetReq.Header.Set("Authorization", "Bearer "+account.AccessToken)
	tweetReq.Header.Set("Content-Type", "application/json")

	tweetResp, err := http.DefaultClient.Do(tweetReq)
	if err != nil {
		return nil, err
	}
	defer tweetResp.Body.Close()

	var tweetResult map[string]interface{}
	json.NewDecoder(tweetResp.Body).Decode(&tweetResult)

	return tweetResult, nil
}

func (s *PlatformService) FetchYouTubeComments(ctx context.Context, userID string) ([]dto.CommentData, error) {
	account, err := s.accountRepo.FindByUserIDAndPlatform(ctx, userID, models.PlatformYouTube)
	if err != nil {
		return nil, err
	}

	channelResp, err := s.doYouTubeGet(account.AccessToken, "https://www.googleapis.com/youtube/v3/channels?part=snippet&mine=true")
	if err != nil {
		return nil, err
	}

	var channelData dto.YouTubeChannelResponse
	if err := json.Unmarshal(channelResp, &channelData); err != nil {
		return nil, err
	}

	if len(channelData.Items) == 0 {
		return nil, fmt.Errorf("channel not found")
	}

	channelID := channelData.Items[0].ID

	videoResp, err := s.doYouTubeGet(account.AccessToken, fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?part=snippet&channelId=%s&order=date&maxResults=10", channelID))
	if err != nil {
		return nil, err
	}

	var videoData dto.YouTubeVideoItems
	if err := json.Unmarshal(videoResp, &videoData); err != nil {
		return nil, err
	}

	videoIDs := make([]string, 0)
	for _, v := range videoData.Items {
		if v.ID.Kind == "youtube#video" {
			videoIDs = append(videoIDs, v.ID.VideoID)
		}
	}

	if len(videoIDs) == 0 {
		return nil, fmt.Errorf("no videos found")
	}

	statsResp, err := s.doYouTubeGet(account.AccessToken, fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=statistics&id=%s", strings.Join(videoIDs, ",")))
	if err != nil {
		return nil, err
	}

	var statsData dto.YouTubeStatistics
	if err := json.Unmarshal(statsResp, &statsData); err != nil {
		return nil, err
	}

	for _, video := range videoData.Items {
		if video.ID.Kind != "youtube#video" {
			continue
		}

		commentResp, err := s.doYouTubeGet(account.AccessToken, fmt.Sprintf("https://www.googleapis.com/youtube/v3/commentThreads?part=snippet&videoId=%s&maxResults=100", video.ID.VideoID))
		if err != nil {
			continue
		}

		var commentData dto.YouTubeCommentResponse
		if err := json.Unmarshal(commentResp, &commentData); err != nil {
			continue
		}

		for _, comment := range commentData.Items {
			post, err := s.postRepo.FindByExternalID(ctx, comment.Snippet.VideoID)
			if err != nil {
				continue
			}

			createdAt, _ := time.Parse(time.RFC3339, comment.Snippet.TopLevelComment.Snippet.PublishedAt)

			commentModel := &models.Comment{
				ExternalID: comment.ID,
				PostID:     post.ID,
				Content:    comment.Snippet.TopLevelComment.Snippet.TextOriginal,
				Likes:      comment.Snippet.TopLevelComment.Snippet.LikeCount,
				CreatedAt:  createdAt,
				AccountID:  account.ID,
			}
			s.commentRepo.Upsert(ctx, commentModel)
		}
	}

	comments, err := s.commentRepo.FindByAccountIDWithPost(ctx, account.ID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.CommentData, 0, len(comments))
	for _, c := range comments {
		result = append(result, dto.CommentData{
			ID:             c.ID,
			ExternalID:     c.ExternalID,
			PostID:         c.PostID,
			Content:        c.Content,
			Likes:          c.Likes,
			CreatedAt:      c.CreatedAt,
			AccountID:      c.AccountID,
			PostContent:    c.PostContent,
			PostAttachment: c.PostAttachment,
		})
	}

	return result, nil
}

func (s *PlatformService) FetchTikTokComments(ctx context.Context, userID string) ([]dto.CommentData, error) {
	return []dto.CommentData{}, nil
}

func (s *PlatformService) FetchTwitterComments(ctx context.Context, userID string) ([]dto.CommentData, error) {
	account, err := s.accountRepo.FindByUserIDAndPlatform(ctx, userID, models.PlatformTwitter)
	if err != nil {
		return nil, err
	}

	posts, err := s.postRepo.FindLatestByAccountID(ctx, account.ID)
	if err != nil || len(posts) == 0 {
		return nil, fmt.Errorf("no posts found")
	}

	lastPost := posts[0]

	searchURL := fmt.Sprintf("https://api.twitter.com/2/tweets/search/recent?query=conversation_id:%s&tweet.fields=in_reply_to_user_id,author_id,created_at,conversation_id,public_metrics", lastPost.ExternalID)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+account.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var commentData dto.TwitterCommentResponse
	if err := json.NewDecoder(resp.Body).Decode(&commentData); err != nil {
		return nil, err
	}

	for _, comment := range commentData.Data {
		post, err := s.postRepo.FindByExternalID(ctx, comment.ConversationID)
		if err != nil {
			continue
		}

		createdAt, _ := time.Parse(time.RFC3339, comment.CreatedAt)

		commentModel := &models.Comment{
			ExternalID: comment.ID,
			PostID:     post.ID,
			Content:    comment.Text,
			Likes:      comment.PublicMetrics.LikeCount,
			CreatedAt:  createdAt,
			AccountID:  account.ID,
		}
		s.commentRepo.Upsert(ctx, commentModel)
	}

	comments, err := s.commentRepo.FindByAccountIDWithPost(ctx, account.ID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.CommentData, 0, len(comments))
	for _, c := range comments {
		result = append(result, dto.CommentData{
			ID:             c.ID,
			ExternalID:     c.ExternalID,
			PostID:         c.PostID,
			Content:        c.Content,
			Likes:          c.Likes,
			CreatedAt:      c.CreatedAt,
			AccountID:      c.AccountID,
			PostContent:    c.PostContent,
			PostAttachment: c.PostAttachment,
		})
	}

	return result, nil
}

func (s *PlatformService) doYouTubeGet(accessToken, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
