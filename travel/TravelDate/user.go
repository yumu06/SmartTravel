package TravelDate

import (
	"errors"
	"strings"
	"travel/TravelModel"
	"travel/vo"

	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

// CheckUserExistWithOpenID
// @Description 数据库查找用户是否存在
// @auth Snactop 2024/12/10 0:25
// @param  userName string  传入一个用户名
// @return bool 返回一个布尔值，用户存在返回true，反之则false
func CheckUserExistWithOpenID(OpenID string) bool {
	db := GetDB()
	//TODO 判断用户是否存在
	var user TravelModel.TraUser
	db.Where("  open_id  = ?", OpenID).First(&user)

	if user.ID != 0 {
		return true
	}
	return false
}

func GetUserInformation(OpenID string) (TravelModel.TraUser, error) {
	db := GetDB()
	//TODO 判断用户是否存在
	var user TravelModel.TraUser
	db.Where("  open_id  = ?", OpenID).First(&user)
	if user.ID == 0 {
		return TravelModel.TraUser{}, errors.New("用户不存在")
	}

	return user, nil
}

func UpdateUserInformation(user TravelModel.TraUser, userUpdate vo.UpdateUserRequest) error {
	db := GetDB()

	if err := db.Model(&user).Updates(userUpdate).Error; err != nil {
		return errors.New("系统错误，用户信息更新失败")
	}
	return nil
}

func CheckUserPostStartExist(userID uint64, postID uuid.UUID) bool {
	db := GetDB()
	var existingFavorite TravelModel.TraUserPostStart
	if err := db.Where("user_id = ? AND post_id = ?", userID, postID).First(&existingFavorite).Error; err == nil {
		return true
	}
	return false
}

func RemovePostStart1(userID uint64, postID uuid.UUID) error {
	db := GetDB()
	if err := db.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&TravelModel.TraUserPostStart{}).Error; err != nil {
		return errors.New("取消收藏失败")
	}
	return nil
}

func AddPostStart(userID uint64, postID uuid.UUID) error {
	db := GetDB()
	favorite := TravelModel.TraUserPostStart{
		UserID: userID, // 转换为uint64类型
		PostID: postID,
	}
	if err := db.Create(&favorite).Error; err != nil {
		err = errors.New("收藏失败，请稍后重试")
		return err
	}
	return nil
}

func GetPostStartIDs(userID uint64) ([]uuid.UUID, error) {
	db := GetDB()
	var postIDs []uuid.UUID
	err := db.Table("tra_user_post_starts"). //收藏表的表名
							Select("post_id").
							Where("user_id = ?", userID).
							Pluck("post_id", &postIDs).Error

	if err != nil {
		return nil, err
	}

	return postIDs, nil
}

func GetPostsByIDs(ids []uuid.UUID) ([]TravelModel.Post, error) {
	db := GetDB()

	var posts []TravelModel.Post

	err := db.Table("posts").
		Where("id IN ?", ids).
		Find(&posts).Error

	if err != nil {
		return nil, err
	}

	return posts, nil
}

func GetUserCreatedPosts(userID uint64) (posts []TravelModel.Post, err error) {
	db := GetDB()

	// 查询用户创建的所有商品
	err = db.Table("posts").
		Where("user_id = ?", userID).
		Find(&posts).Error

	if err != nil {
		return []TravelModel.Post{}, errors.New("获取用户创建的文章失败，请稍后重试")
	}

	// 如果用户没有创建任何商品
	if len(posts) == 0 {
		return []TravelModel.Post{}, nil
	}

	return posts, nil
}

func CreateFoot(foot TravelModel.TraFoot) (uuid.UUID, error) {
	db := GetDB()
	if err := db.Create(&foot).Error; err != nil {
		return uuid.Nil, errors.New("创建足迹失败")
	}
	return foot.ID, nil
}

func GetFootByID(footID uuid.UUID) (TravelModel.TraFoot, error) {
	db := GetDB()
	var foot TravelModel.TraFoot
	if err := db.First(&foot, "id = ?", footID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return TravelModel.TraFoot{}, errors.New("足迹不存在")
		}
		return TravelModel.TraFoot{}, errors.New("获取足迹失败")
	}
	return foot, nil
}

func ListUserFoots(userID uint64, pageNum int, pageSize int) ([]TravelModel.TraFoot, int64, error) {
	db := GetDB()
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (pageNum - 1) * pageSize

	var total int64
	if err := db.Model(&TravelModel.TraFoot{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, errors.New("获取足迹失败")
	}

	var foots []TravelModel.TraFoot
	if err := db.Where("user_id = ?", userID).Order("created_at desc").Limit(pageSize).Offset(offset).Find(&foots).Error; err != nil {
		return nil, 0, errors.New("获取足迹失败")
	}
	return foots, total, nil
}

func CheckUserFootStartExist(userID uint64, footID uuid.UUID) bool {
	db := GetDB()
	var existing TravelModel.TraUserFootStart
	if err := db.Where("user_id = ? AND foot_id = ?", userID, footID).First(&existing).Error; err == nil {
		return true
	}
	return false
}

func AddFootStart(userID uint64, footID uuid.UUID) error {
	db := GetDB()
	favorite := TravelModel.TraUserFootStart{
		UserID: userID,
		FootID: footID,
	}
	if err := db.Create(&favorite).Error; err != nil {
		return errors.New("收藏失败，请稍后重试")
	}
	return nil
}

func RemoveFootStart(userID uint64, footID uuid.UUID) error {
	db := GetDB()
	if err := db.Where("user_id = ? AND foot_id = ?", userID, footID).Delete(&TravelModel.TraUserFootStart{}).Error; err != nil {
		return errors.New("取消收藏失败")
	}
	return nil
}

func GetFootStartIDs(userID uint64) ([]uuid.UUID, error) {
	db := GetDB()
	var footIDs []uuid.UUID
	err := db.Table("tra_user_foot_starts").
		Select("foot_id").
		Where("user_id = ?", userID).
		Pluck("foot_id", &footIDs).Error

	if err != nil {
		return nil, err
	}
	return footIDs, nil
}

func GetFootsByIDs(ids []uuid.UUID) ([]TravelModel.TraFoot, error) {
	db := GetDB()
	var foots []TravelModel.TraFoot
	if err := db.Model(&TravelModel.TraFoot{}).Where("id IN ?", ids).Find(&foots).Error; err != nil {
		return nil, err
	}
	return foots, nil
}

func DeleteFootByID(userID uint64, footID uuid.UUID) error {
	db := GetDB()
	var foot TravelModel.TraFoot
	if err := db.First(&foot, "id = ?", footID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("足迹不存在")
		}
		return errors.New("删除足迹失败")
	}
	if foot.UserID != userID {
		return errors.New("无权限删除该足迹")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("foot_id = ?", footID).Delete(&TravelModel.TraUserFootStart{}).Error; err != nil {
			return errors.New("删除足迹失败")
		}
		if err := tx.Delete(&foot).Error; err != nil {
			return errors.New("删除足迹失败")
		}
		return nil
	})
}

func SearchUsers(keyword string, limit int) ([]TravelModel.TraUser, error) {
	db := GetDB()
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return []TravelModel.TraUser{}, errors.New("keyword 不能为空")
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	like := "%" + keyword + "%"
	var users []TravelModel.TraUser
	if err := db.Model(&TravelModel.TraUser{}).
		Where("nick_name LIKE ? OR telephone LIKE ?", like, like).
		Order("updated_at desc").
		Limit(limit).
		Find(&users).Error; err != nil {
		return []TravelModel.TraUser{}, errors.New("用户搜索失败")
	}
	return users, nil
}

func CreateChatMessage(msg TravelModel.ChatMessage) (uuid.UUID, error) {
	db := GetDB()
	if err := db.Create(&msg).Error; err != nil {
		return uuid.Nil, errors.New("发送失败")
	}
	return msg.ID, nil
}

func ListChatMessages(userID uint64, withUserID uint64, pageNum int, pageSize int) ([]TravelModel.ChatMessage, int64, error) {
	db := GetDB()
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (pageNum - 1) * pageSize

	query := db.Model(&TravelModel.ChatMessage{}).
		Where("(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)", userID, withUserID, withUserID, userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return []TravelModel.ChatMessage{}, 0, errors.New("获取对话失败")
	}

	var msgs []TravelModel.ChatMessage
	if err := query.Order("created_at asc").Limit(pageSize).Offset(offset).Find(&msgs).Error; err != nil {
		return []TravelModel.ChatMessage{}, 0, errors.New("获取对话失败")
	}
	return msgs, total, nil
}
