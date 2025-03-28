package postrepo

import (
	"errors"
	"log"
	"time"

	"forum/repositories/shared"
)

// PostService manages post operations
type PostService struct {
	post   PostRepo
	shared *shared.SharedConfig
}

// NewPostService creates a new instance of PostService
func NewPostService(post PostRepo) *PostService {
	shared := shared.NewSharedConfig()
	return &PostService{post, shared}
}

func (p *PostService) CreatePost(recievedPost *Post) (*Post, error) {
	sanitized, err := p.shared.SanitizeInput(recievedPost)
	if err != nil {
		return nil, err
	}

	post, ok := sanitized.(*Post)
	if !ok {
		return nil, errors.New("bad request")
	}

	if post.UserID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	if post.PostTitle == "" {
		return nil, errors.New("post title cannot be empty")
	}

	if post.PostContent == "" {
		return nil, errors.New("post content cannot be empty")
	}

	if post.PostID == "" {
		post.PostID, _ = p.shared.GenerateUUID()
	}

	if !post.CreatedAt.IsZero() {
		return p.post.UpdatePost(post)
	}

	post.AuthorImg = "/static/profiles/" + post.UserID
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()
	post.HasComments = true

	go p.RecordActivity(post.UserID, "created_post", "created a post: "+post.PostTitle)

	return p.post.CreatePost(post)
}

func (p *PostService) ListPosts() ([]*Post, error) {
	posts, err := p.post.ListPosts()
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (p *PostService) GetLikedPosts(userId string) ([]*Post, error) {
	return p.post.GetLikedPostsByUserID(userId)
}

func (p *PostService) DeletePost(post *Post) error {
	if post.PostID == "" {
		return errors.New("bad request")
	}

	post, err := p.post.GetPostByID(post.PostID)
	if err != nil {
		return errors.New("bad request")
	}

	go p.RecordActivity(post.UserID, "deleted_post", "deleted a post: "+post.PostTitle)

	return p.post.DeletePost(post.PostID)
}

func (p *PostService) PostAddLike(like *Like) (*Like, error) {
	if like.UserID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	post, err := p.post.GetPostByID(like.PostID)
	if err != nil {
		return nil, errors.New("bad request")
	}

	msg := "removed a post like"

	hasDisliked, _ := p.post.HasUserDisliked(post.PostID, like.UserID, "Post")
	if hasDisliked != "" {
		like.LikeID = hasDisliked

		p.DeleteLike(like, "dislikes")
	}

	haslike, _ := p.post.HasUserLiked(post.PostID, like.UserID, "Post")
	if haslike != "" {
		like.LikeID = haslike

		go p.RecordActivity(like.UserID, msg, msg+" on: "+post.PostTitle)

		return nil, p.DeleteLike(like, "likes")
	}

	like.LikeID, _ = p.shared.GenerateUUID()
	like.CreatedAt = time.Now()

	msg = "liked a post: "
	go p.RecordActivity(like.UserID, msg, msg+post.PostTitle)

	nId, _ := p.shared.GenerateUUID()
	_, err = p.CreateNotification(&Notification{
		NotificationID:   nId,
		UserId:           post.UserID,
		SenderID:         like.UserID,
		PostID:           p.shared.ToNullString(post.PostID),
		NotificationType: "success",
		Message:          "liked: " + post.PostTitle,
		IsRead:           false,
		CreatedAt:        time.Now(),
	})
	if err != nil {
		log.Println(err)
	}

	return p.post.AddLike(like)
}

func (p *PostService) PostDisLike(dislike *Like) (*Like, error) {
	if dislike.UserID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	post, err := p.post.GetPostByID(dislike.PostID)
	if err != nil {
		return nil, errors.New("bad request")
	}

	msg := "removed a post dislike"

	haslike, _ := p.post.HasUserLiked(post.PostID, dislike.UserID, "Post")
	if haslike != "" {
		dislike.LikeID = haslike
		p.DeleteLike(dislike, "likes")
	}

	hasDisliked, _ := p.post.HasUserDisliked(post.PostID, dislike.UserID, "Post")
	if hasDisliked != "" {
		dislike.LikeID = hasDisliked

		go p.RecordActivity(dislike.UserID, msg, msg+" on: "+post.PostTitle)

		return nil, p.DeleteLike(dislike, "dislikes")
	}

	dislike.LikeID, _ = p.shared.GenerateUUID()
	dislike.CreatedAt = time.Now()

	msg = "disliked a post"
	go p.RecordActivity(dislike.UserID, msg, msg+" : "+post.PostTitle)

	nId, _ := p.shared.GenerateUUID()
	_, err = p.CreateNotification(&Notification{
		NotificationID:   nId,
		UserId:           post.UserID,
		SenderID:         dislike.UserID,
		PostID:           p.shared.ToNullString(post.PostID),
		NotificationType: "warning",
		Message:          post.PostTitle + " performing badly",
		IsRead:           false,
		CreatedAt:        time.Now(),
	})
	if err != nil {
		log.Println(err)
	}

	return p.post.PostDislike(dislike)
}

func (p *PostService) CommentAddLike(like *Like) (*Like, error) {
	if like.UserID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	comment, err := p.post.GetCommentByID(like.CommentID)
	if err != nil {
		return nil, errors.New("bad request")
	}

	hasDisLike, _ := p.post.HasUserDisliked(like.CommentID, like.UserID, "Comment")
	if hasDisLike != "" {
		like.LikeID = hasDisLike

		p.DeleteLike(like, "dislikes")
	}

	msg := "removed a comment like"
	haslike, _ := p.post.HasUserLiked(like.CommentID, like.UserID, "Comment")
	if haslike != "" {
		like.LikeID = haslike

		go p.RecordActivity(like.UserID, msg, msg+" on: "+comment.Content)

		return nil, p.DeleteLike(like, "likes")
	}

	like.LikeID, _ = p.shared.GenerateUUID()
	like.CreatedAt = time.Now()

	msg = "liked a comment: "
	go p.RecordActivity(like.UserID, msg, msg+comment.Content)

	nId, _ := p.shared.GenerateUUID()
	_, err = p.CreateNotification(&Notification{
		NotificationID:   nId,
		UserId:           comment.UserID,
		SenderID:         like.UserID,
		CommentID:        p.shared.ToNullString(comment.CommentID),
		NotificationType: "success",
		Message:          "liked: " + comment.Content,
		IsRead:           false,
		CreatedAt:        time.Now(),
	})
	if err != nil {
		log.Println(err)
	}

	return p.post.AddLike(like)
}

func (p *PostService) CommentAddDisLike(dislike *Like) (*Like, error) {
	if dislike.UserID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	comment, err := p.post.GetCommentByID(dislike.CommentID)
	if err != nil {
		return nil, errors.New("bad request")
	}

	msg := "removed a comment dislike"

	haslike, _ := p.post.HasUserLiked(dislike.CommentID, dislike.UserID, "Comment")
	if haslike != "" {
		dislike.LikeID = haslike
		p.DeleteLike(dislike, "likes")
	}

	hasDisLike, _ := p.post.HasUserDisliked(dislike.CommentID, dislike.UserID, "Comment")
	if hasDisLike != "" {
		dislike.LikeID = hasDisLike

		go p.RecordActivity(dislike.UserID, msg, msg+" on: "+comment.Content)

		return nil, p.DeleteLike(dislike, "dislikes")
	}

	dislike.LikeID, _ = p.shared.GenerateUUID()
	dislike.CreatedAt = time.Now()

	msg = "disliked a comment: "
	go p.RecordActivity(dislike.UserID, msg, msg+comment.Content)

	nId, _ := p.shared.GenerateUUID()
	_, err = p.CreateNotification(&Notification{
		NotificationID:   nId,
		UserId:           comment.UserID,
		SenderID:         dislike.UserID,
		CommentID:        p.shared.ToNullString(comment.CommentID),
		NotificationType: "success",
		Message:          "disliked: " + comment.Content,
		IsRead:           false,
		CreatedAt:        time.Now(),
	})
	if err != nil {
		log.Println(err)
	}

	return p.post.PostDislike(dislike)
}

func (p *PostService) DeleteLike(dislike *Like, entityType string) error {
	if dislike.LikeID == "" {
		return errors.New("like ID cannot be empty")
	}

	return p.post.DisLike(dislike, entityType)
}

// CreateComment creates a new comment
func (p *PostService) CreatePostComment(receivedComment *Comment) (*Comment, error) {
	sanitized, err := p.shared.SanitizeInput(receivedComment)
	if err != nil {
		return nil, err
	}

	comment, ok := sanitized.(*Comment)
	if !ok {
		return nil, errors.New("bad request")
	}

	if comment.UserID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	if comment.PostID == "" {
		return nil, errors.New("post ID cannot be empty")
	}

	if comment.Content == "" {
		return nil, errors.New("comment content cannot be empty")
	}

	post, err := p.post.GetPostByID(comment.PostID)
	if err != nil {
		return nil, errors.New("bad request")
	}

	comment.PostTitle = post.PostTitle
	comment.PostAuthor = post.PostAuthor
	comment.PostAuthorImg = post.AuthorImg
	comment.AuthorImg = "/static/profiles/" + comment.UserID
	comment.CommentID, _ = p.shared.GenerateUUID()
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	go p.RecordActivity(comment.UserID, "created_comment", "added a comment: "+comment.Content)

	nId, _ := p.shared.GenerateUUID()
	go p.CreateNotification(&Notification{
		NotificationID:   nId,
		UserId:           post.UserID,
		SenderID:         comment.UserID,
		PostID:           p.shared.ToNullString(post.PostID),
		NotificationType: "success",
		Message:          "commented: " + post.PostTitle,
		IsRead:           false,
		CreatedAt:        time.Now(),
	})

	return p.post.CreateComment(comment)
}

// UpdateComment updates a comment
func (p *PostService) UpdateComment(comment *Comment) (*Comment, error) {
	if comment.CommentID == "" {
		return nil, errors.New("comment ID cannot be empty")
	}

	if comment.Content == "" {
		return nil, errors.New("comment content cannot be empty")
	}

	go p.RecordActivity(comment.UserID, "updated_comment", "updated a comment: "+comment.Content)

	return p.post.UpdateComment(comment)
}

// DeleteComment deletes a comment
func (p *PostService) DeleteComment(comment *Comment) error {
	if comment.CommentID == "" {
		return errors.New("comment ID cannot be empty")
	}

	go p.RecordActivity(comment.UserID, "deleted_comment", "deleted a comment: "+comment.Content)

	return p.post.DeleteComment(comment.CommentID)
}

// CreateReply creates a new reply
func (p *PostService) CreateCommentReply(reply *Reply) (*Reply, error) {
	if reply.UserID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	if reply.CommentID == "" {
		return nil, errors.New("comment ID cannot be empty")
	}

	if reply.Content == "" {
		return nil, errors.New("reply content cannot be empty")
	}

	com, err := p.post.GetCommentByID(reply.CommentID)
	if err != nil {
		return nil, errors.New("bad request")
	}

	reply.AuthorImg = "/static/profiles/" + reply.UserID
	reply.ReplyID, _ = p.shared.GenerateUUID()
	reply.CreatedAt = time.Now()
	reply.UpdatedAt = time.Now()

	go p.RecordActivity(reply.UserID, "replied to a comment", "replied: "+reply.Content)

	nId, _ := p.shared.GenerateUUID()
	_, err = p.CreateNotification(&Notification{
		NotificationID:   nId,
		UserId:           com.UserID,
		SenderID:         reply.UserID,
		ReplyID:          p.shared.ToNullString(reply.ReplyID),
		NotificationType: "info",
		Message:          "replied: " + reply.Content,
		IsRead:           false,
		CreatedAt:        time.Now(),
	})
	if err != nil {
		log.Println(err)
	}

	return p.post.CreateReply(reply)
}

func (p *PostService) ReplyAddLike(like *Like) (*Like, error) {
	if like.UserID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	reply, err := p.post.GetReplyByID(like.ReplyID)
	if err != nil {
		return nil, errors.New("bad request")
	}

	msg := "removed a replt like"

	haslike, _ := p.post.HasUserLiked(like.ReplyID, like.UserID, "Reply")
	if haslike != "" {
		like.LikeID = haslike

		go p.RecordActivity(like.UserID, msg, msg+" on: "+reply.Content)
		return nil, p.DeleteLike(like, "likes")
	}

	like.LikeID, _ = p.shared.GenerateUUID()
	like.CreatedAt = time.Now()

	msg = "liked a reply: "
	go p.RecordActivity(like.UserID, msg, msg+reply.Content)

	nId, _ := p.shared.GenerateUUID()
	_, err = p.CreateNotification(&Notification{
		NotificationID:   nId,
		UserId:           reply.UserID,
		SenderID:         like.UserID,
		CommentID:        p.shared.ToNullString(reply.CommentID),
		NotificationType: "success",
		Message:          "liked: " + reply.Content,
		IsRead:           false,
		CreatedAt:        time.Now(),
	})
	if err != nil {
		log.Println(err)
	}

	return p.post.AddLike(like)
}

// GetPostsByUserID retrieves all posts created by a specific user
func (p *PostService) GetPostsByUserID(userID string) ([]*Post, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	return p.post.GetPostsByUserID(userID)
}

// GetCommentsByUserID retrieves all comments created by a specific user
func (p *PostService) GetCommentsByUserID(userID string) ([]*Comment, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	return p.post.GetCommentsByUserID(userID)
}

// GetRepliesByUserID retrieves all replies created by a specific user
func (p *PostService) GetRepliesByUserID(userID string) ([]*Reply, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	return p.post.GetRepliesByUserID(userID)
}

// GetLikesByUserID retrieves all likes created by a specific user
func (p *PostService) GetLikesByUserID(userID string) ([]*Like, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	return p.post.GetLikesByUserID(userID)
}

// GetDislikesByUserID retrieves all dislikes created by a specific user
func (p *PostService) GetDislikesByUserID(userID string) ([]*Like, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	return p.post.GetDislikesByUserID(userID)
}

// AddActivity adds a new activity to the database
func (p *PostService) AddActivity(activity *Activity) (*Activity, error) {
	if activity.UserId == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	return p.post.AddActivity(activity)
}

// RecordActivity records an activity
func (p *PostService) RecordActivity(userID, activityType, activityData string) (*Activity, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	var activity Activity

	activity.ActivityID, _ = p.shared.GenerateUUID()
	activity.UserId = userID
	activity.ActivityType = activityType
	activity.ActivityData = activityData
	activity.CreatedAt = time.Now()

	return p.post.AddActivity(&activity)
}

// GetActivitiesByUserID retrieves all activities created by a specific user
func (p *PostService) GetActivitiesByUserID(userID string) ([]*Activity, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	return p.post.GetActivitiesByUserID(userID)
}

// CreateNotification creates a new notification
func (p *PostService) CreateNotification(n *Notification) (*Notification, error) {
	if n.UserId == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	return p.post.CreateNotification(n)
}

// GetNotificationsByUserID retrieves all notifications created by a specific user
func (p *PostService) GetNotificationsByUserID(userID string) ([]*Notification, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	return p.post.GetNotificationsByUserID(userID)
}
