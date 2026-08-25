package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var moodKeywordMap = map[string][]string{
	"探索感": {"老街", "小众景点", "历史街区", "文创馆", "非遗", "冷门打卡地"},
	"放松感": {"咖啡馆", "书店", "静吧", "露营地", "植物园", "海边栈道"},
	"治愈感": {"公园", "海边", "图书馆", "绿地", "湖边", "安静"},
	"冒险感": {"徒步", "山路", "秘境", "自然探险", "灯塔", "岛屿"},
	"社交感": {"夜市", "展览", "步行街", "市集", "美食广场", "主题活动"},
	"浪漫感": {"夜景", "情侣座", "灯光", "海风", "高空", "花园"},
	"空白感": {"随便走走", "街道", "附近好去处", "散步", "休闲"},
	"回忆感": {"老地方", "母校", "中学", "小时候", "熟悉的街", "曾经的打卡点"},
}

func Recommend(ctx *gin.Context) {
	mood := ctx.Query("mood")
	if mood == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "情绪不能为空"})
		return
	}

	region := ctx.Query("region")
	if region == "" {
		region = viper.GetString("baidu.region")
	}
	if region == "" {
		region = "汕头"
	}

	keywords := moodKeywordMap[mood]
	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"mood":     mood,
			"keywords": keywords,
			"region":   region,
		},
		"msg": "获取推荐关键词成功",
	})
}
