package logic

import (
	"errors"
	"travel/TravelDate"
	"travel/TravelModel"
	"travel/vo"

	uuid "github.com/satori/go.uuid"
)

func PostCreate(userID uint64, p vo.PostRequest) (uuid.UUID, error) {
	//创建文章
	post := TravelModel.Post{
		UserID:  userID,
		Title:   p.Title,
		HeadImg: p.HeadImg,
		Content: p.Content,
	}

	postID, err := TravelDate.InsertPost(post)

	if err != nil {
		return uuid.Nil, errors.New("系统错误，文章创建失败")
	}

	return postID, nil
}

func PostUpdate(userID uint64, postID string, postR vo.PostRequest) (TravelModel.Post, error) {
	//查找commodityID是否存在
	post, err := TravelDate.GetPostInfo(postID)
	if err != nil {
		return TravelModel.Post{}, err
	}

	//判断是否为商品商家
	if userID != post.UserID {
		err := errors.New("用户不是文章作者，用户无修改权限")
		return TravelModel.Post{}, err
	}

	//更新商品信息
	if err := TravelDate.UpdatePost(&post, postR); err != nil {
		return TravelModel.Post{}, err
	}

	return post, nil
}

func PostDelete(userID uint64, postID string) error {
	//查找postID是否存在
	post, err := TravelDate.GetPostInfo(postID)
	if err != nil {
		return err
	}

	//判断是否为商品商家
	if userID != post.UserID {
		err := errors.New("删除失败，用户无修改权限")
		return err
	}

	//删除商品
	if err := TravelDate.DeletePost(post); err != nil {
		return err
	}

	return nil
}

func GetPostInfo(postID string) (TravelModel.Post, error) {
	post, err := TravelDate.GetPostInfo(postID)
	if err != nil {
		return TravelModel.Post{}, err
	}
	return post, nil
}

func PageList(pageNum int, pageList int) ([]TravelModel.Post, int64) {
	commodities, total := TravelDate.PageList(pageNum, pageList)
	return commodities, total
}

func IncrementPostView(postID string) {
	id, err := uuid.FromString(postID)
	if err != nil {
		return
	}
	_ = TravelDate.IncrementPostView(id)
}

func SearchPosts(keyword string, pageNum int, pageSize int) ([]TravelModel.Post, int64, error) {
	if keyword == "" {
		return []TravelModel.Post{}, 0, errors.New("keyword 不能为空")
	}
	return TravelDate.SearchPosts(keyword, pageNum, pageSize)
}

func ListPostsByUserID(userID uint64, pageNum int, pageSize int) ([]TravelModel.Post, int64, error) {
	return TravelDate.ListPostsByUserID(userID, pageNum, pageSize)
}

func AddPostLike(userID uint64, postID string) error {
	post, err := TravelDate.GetPostInfo(postID)
	if err != nil {
		return err
	}
	if TravelDate.CheckUserPostLikeExist(userID, post.ID) {
		return errors.New("已点赞")
	}
	return TravelDate.AddPostLike(userID, post.ID)
}

func RemovePostLike(userID uint64, postID string) error {
	post, err := TravelDate.GetPostInfo(postID)
	if err != nil {
		return err
	}
	if !TravelDate.CheckUserPostLikeExist(userID, post.ID) {
		return errors.New("未点赞")
	}
	return TravelDate.RemovePostLike(userID, post.ID)
}

func CreateComment(userID uint64, postID string, content string) (uuid.UUID, error) {
	if content == "" {
		return uuid.Nil, errors.New("评论内容不能为空")
	}
	post, err := TravelDate.GetPostInfo(postID)
	if err != nil {
		return uuid.Nil, err
	}
	comment := TravelModel.PostComment{
		PostID:  post.ID,
		UserID:  userID,
		Content: content,
	}
	return TravelDate.CreateComment(comment)
}

func ListComments(postID string, pageNum int, pageSize int) ([]TravelModel.PostComment, int64, error) {
	id, err := uuid.FromString(postID)
	if err != nil {
		return []TravelModel.PostComment{}, 0, errors.New("文章ID格式错误")
	}
	return TravelDate.ListCommentsByPost(id, pageNum, pageSize)
}

func DeleteComment(userID uint64, commentID string) error {
	id, err := uuid.FromString(commentID)
	if err != nil {
		return errors.New("评论ID格式错误")
	}
	return TravelDate.DeleteComment(userID, id)
}

func CreateNotice(title string, content string, imageURL string, linkURL string) (uuid.UUID, error) {
	if title == "" || content == "" {
		return uuid.Nil, errors.New("标题和内容不能为空")
	}
	notice := TravelModel.Notice{
		Title:    title,
		Content:  content,
		ImageURL: imageURL,
		LinkURL:  linkURL,
	}
	return TravelDate.CreateNotice(notice)
}

func UpdateNotice(noticeID string, title string, content string, imageURL string, linkURL string) error {
	id, err := uuid.FromString(noticeID)
	if err != nil {
		return errors.New("公告ID格式错误")
	}

	updates := map[string]interface{}{}
	if title != "" {
		updates["title"] = title
	}
	if content != "" {
		updates["content"] = content
	}
	if imageURL != "" {
		updates["image_url"] = imageURL
	}
	if linkURL != "" {
		updates["link_url"] = linkURL
	}
	if len(updates) == 0 {
		return errors.New("至少更新一项")
	}

	return TravelDate.UpdateNotice(id, updates)
}

func DeleteNotice(noticeID string) error {
	id, err := uuid.FromString(noticeID)
	if err != nil {
		return errors.New("公告ID格式错误")
	}
	return TravelDate.DeleteNotice(id)
}

func GetNotice(noticeID string) (TravelModel.Notice, error) {
	id, err := uuid.FromString(noticeID)
	if err != nil {
		return TravelModel.Notice{}, errors.New("公告ID格式错误")
	}
	return TravelDate.GetNoticeByID(id)
}

func ListNotices(pageNum int, pageSize int) ([]TravelModel.Notice, int64, error) {
	return TravelDate.ListNotices(pageNum, pageSize)
}

func RecommendPosts(userID uint64, limit int) ([]TravelModel.Post, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	posts, err := TravelDate.RecommendPosts(userID, limit)
	if err == nil && len(posts) > 0 {
		return posts, nil
	}
	return TravelDate.ListLatestPosts(limit)
}
