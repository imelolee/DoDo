package model

import "time"

type Video struct {
	Id          int64 `json:"id"`
	AuthorId    int64
	PlayUrl     string `json:"play_url"`
	CoverUrl    string `json:"cover_url"`
	PublishTime time.Time
	Title       string `json:"title"` //视频名，5.23添加
}

func GetVideoIdList(userId int64) ([]int64, error) {
	var id []int64
	//通过pluck来获得单独的切片
	result := Db.Model(&Video{}).Where("author_id = ?", userId).Pluck("id", &id)
	//如果出现问题，返回对应到空，并且返回error
	if result.Error != nil {
		return nil, result.Error
	}

	return id, nil
}
