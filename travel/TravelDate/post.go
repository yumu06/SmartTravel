package TravelDate

import (
	"errors"
	"travel/TravelModel"
	"travel/vo"

	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

func InsertPost(post TravelModel.Post) (uuid.UUID, error) {
	db := GetDB()
	if err := db.Create(&post).Error; err != nil {
		return uuid.Nil, err
	}
	return post.ID, nil
}

func UpdatePost(post *TravelModel.Post, postR vo.PostRequest) error {
	db := GetDB()
	if err := db.Model(post).Updates(postR).Error; err != nil {
		err := errors.New("商品信息更新失败")
		return err
	}
	return nil
}

func DeletePost(post TravelModel.Post) error {
	db := GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", post.ID).Delete(&TravelModel.PostComment{}).Error; err != nil {
			return errors.New("系统错误，评论删除失败")
		}
		if err := tx.Where("post_id = ?", post.ID).Delete(&TravelModel.TraUserPostLike{}).Error; err != nil {
			return errors.New("系统错误，点赞记录删除失败")
		}
		if err := tx.Where("post_id = ?", post.ID).Delete(&TravelModel.TraUserPostStart{}).Error; err != nil {
			return errors.New("系统错误，收藏记录删除失败")
		}
		if err := tx.Delete(&post).Error; err != nil {
			return errors.New("系统错误，文章删除失败")
		}
		return nil
	})
}

func GetPostInfo(postID string) (TravelModel.Post, error) {
	db := GetDB()
	// 判断商品是否存在
	var post TravelModel.Post
	db.Where("id = ?", postID).First(&post)
	if post.ID == uuid.Nil {
		err := errors.New("商品不存在")
		return TravelModel.Post{}, err
	}
	return post, nil
}

func PageList(pageNum int, pageList int) ([]TravelModel.Post, int64) {
	db := GetDB()
	var commodities []TravelModel.Post
	db.Order("created_at desc").Offset((pageNum - 1) * pageList).Limit(pageList).Find(&commodities)
	//前端渲染分页需要知道总数
	var total int64
	db.Model([]TravelModel.Post{}).Count(&total)
	return commodities, total
}

func IncrementPostView(postID uuid.UUID) error {
	db := GetDB()
	return db.Model(&TravelModel.Post{}).Where("id = ?", postID).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

func SearchPosts(keyword string, pageNum int, pageSize int) ([]TravelModel.Post, int64, error) {
	db := GetDB()
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (pageNum - 1) * pageSize

	like := "%" + keyword + "%"
	var total int64
	if err := db.Model(&TravelModel.Post{}).Where("title LIKE ? OR content LIKE ?", like, like).Count(&total).Error; err != nil {
		return nil, 0, errors.New("搜索失败")
	}

	var posts []TravelModel.Post
	if err := db.Where("title LIKE ? OR content LIKE ?", like, like).Order("created_at desc").Limit(pageSize).Offset(offset).Find(&posts).Error; err != nil {
		return nil, 0, errors.New("搜索失败")
	}
	return posts, total, nil
}

func ListPostsByUserID(userID uint64, pageNum int, pageSize int) ([]TravelModel.Post, int64, error) {
	db := GetDB()
	if userID == 0 {
		return []TravelModel.Post{}, 0, errors.New("userID 不能为空")
	}
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (pageNum - 1) * pageSize

	var total int64
	if err := db.Model(&TravelModel.Post{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, errors.New("获取文章失败")
	}

	var posts []TravelModel.Post
	if err := db.Where("user_id = ?", userID).Order("created_at desc").Limit(pageSize).Offset(offset).Find(&posts).Error; err != nil {
		return nil, 0, errors.New("获取文章失败")
	}
	return posts, total, nil
}

func CheckUserPostLikeExist(userID uint64, postID uuid.UUID) bool {
	db := GetDB()
	var existing TravelModel.TraUserPostLike
	if err := db.Where("user_id = ? AND post_id = ?", userID, postID).First(&existing).Error; err == nil {
		return true
	}
	return false
}

func AddPostLike(userID uint64, postID uuid.UUID) error {
	db := GetDB()
	like := TravelModel.TraUserPostLike{
		UserID: userID,
		PostID: postID,
	}
	if err := db.Create(&like).Error; err != nil {
		return errors.New("点赞失败")
	}
	if err := db.Model(&TravelModel.Post{}).Where("id = ?", postID).UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
		return errors.New("点赞失败")
	}
	return nil
}

func RemovePostLike(userID uint64, postID uuid.UUID) error {
	db := GetDB()
	if err := db.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&TravelModel.TraUserPostLike{}).Error; err != nil {
		return errors.New("取消点赞失败")
	}
	if err := db.Model(&TravelModel.Post{}).Where("id = ? AND like_count > 0", postID).UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).Error; err != nil {
		return errors.New("取消点赞失败")
	}
	return nil
}

func CreateComment(comment TravelModel.PostComment) (uuid.UUID, error) {
	db := GetDB()
	if err := db.Create(&comment).Error; err != nil {
		return uuid.Nil, errors.New("创建评论失败")
	}
	return comment.ID, nil
}

func ListCommentsByPost(postID uuid.UUID, pageNum int, pageSize int) ([]TravelModel.PostComment, int64, error) {
	db := GetDB()
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (pageNum - 1) * pageSize

	var total int64
	if err := db.Model(&TravelModel.PostComment{}).Where("post_id = ?", postID).Count(&total).Error; err != nil {
		return nil, 0, errors.New("获取评论失败")
	}

	var comments []TravelModel.PostComment
	if err := db.Where("post_id = ?", postID).Order("created_at desc").Limit(pageSize).Offset(offset).Find(&comments).Error; err != nil {
		return nil, 0, errors.New("获取评论失败")
	}
	return comments, total, nil
}

func DeleteComment(userID uint64, commentID uuid.UUID) error {
	db := GetDB()
	var comment TravelModel.PostComment
	if err := db.First(&comment, "id = ?", commentID).Error; err != nil {
		return errors.New("评论不存在")
	}
	if comment.UserID != userID {
		return errors.New("无权限删除该评论")
	}
	if err := db.Delete(&comment).Error; err != nil {
		return errors.New("删除评论失败")
	}
	return nil
}

func CreateNotice(notice TravelModel.Notice) (uuid.UUID, error) {
	db := GetDB()
	if err := db.Create(&notice).Error; err != nil {
		return uuid.Nil, errors.New("创建公告失败")
	}
	return notice.ID, nil
}

func UpdateNotice(noticeID uuid.UUID, updates map[string]interface{}) error {
	db := GetDB()
	if err := db.Model(&TravelModel.Notice{}).Where("id = ?", noticeID).Updates(updates).Error; err != nil {
		return errors.New("更新公告失败")
	}
	return nil
}

func DeleteNotice(noticeID uuid.UUID) error {
	db := GetDB()
	if err := db.Delete(&TravelModel.Notice{}, "id = ?", noticeID).Error; err != nil {
		return errors.New("删除公告失败")
	}
	return nil
}

func GetNoticeByID(noticeID uuid.UUID) (TravelModel.Notice, error) {
	db := GetDB()
	var notice TravelModel.Notice
	if err := db.First(&notice, "id = ?", noticeID).Error; err != nil {
		return TravelModel.Notice{}, errors.New("公告不存在")
	}
	return notice, nil
}

func ListNotices(pageNum int, pageSize int) ([]TravelModel.Notice, int64, error) {
	db := GetDB()
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (pageNum - 1) * pageSize

	var total int64
	if err := db.Model(&TravelModel.Notice{}).Count(&total).Error; err != nil {
		return nil, 0, errors.New("获取公告失败")
	}

	var notices []TravelModel.Notice
	if err := db.Order("created_at desc").Limit(pageSize).Offset(offset).Find(&notices).Error; err != nil {
		return nil, 0, errors.New("获取公告失败")
	}
	return notices, total, nil
}

func RecommendPosts(userID uint64, limit int) ([]TravelModel.Post, error) {
	db := GetDB()
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	var posts []TravelModel.Post
	if err := db.Model(&TravelModel.Post{}).
		Where("user_id <> ?", userID).
		Order("like_count desc, view_count desc, created_at desc").
		Limit(limit).
		Find(&posts).Error; err != nil {
		return []TravelModel.Post{}, errors.New("获取推荐失败")
	}
	return posts, nil
}

func ListLatestPosts(limit int) ([]TravelModel.Post, error) {
	db := GetDB()
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	var posts []TravelModel.Post
	if err := db.Model(&TravelModel.Post{}).Order("created_at desc").Limit(limit).Find(&posts).Error; err != nil {
		return []TravelModel.Post{}, errors.New("获取文章失败")
	}
	return posts, nil
}
