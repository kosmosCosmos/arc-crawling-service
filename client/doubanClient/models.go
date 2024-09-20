package doubanClient

import (
	_ "github.com/go-sql-driver/mysql"
)

type Topic struct {
	Id            int64  `xorm:"pk autoincr"`
	TopicId       string `xorm:"varchar(255) notnull" json:"topic_id"`
	TopicUrl      string `xorm:"varchar(255) notnull" json:"topic_url"`
	UserName      string `xorm:"varchar(255) notnull" json:"user_name"`
	UserId        string `xorm:"varchar(255) notnull" json:"user_id"`
	UserUrl       string `xorm:"varchar(255) notnull" json:"user_url"`
	Title         string `xorm:"varchar(255) notnull" json:"title"`
	GroupId       string `xorm:"varchar(255) notnull" json:"group_id"`
	TopicStatus   string `xorm:"varchar(255) notnull" json:"topic_status"`
	Content       string `xorm:"longtext notnull" json:"content"`
	ReplyCount    int    `xorm:"int notnull" json:"reply_count"`
	CreateTime    int64  `xorm:"BigInt(20) notnull" json:"create_time"`
	LastReplyTime string `xorm:"varchar(255) notnull" json:"last_reply_time"`
}

type Reply struct {
	Id        int64  `xorm:"pk autoincr"`
	TopicId   string `xorm:"varchar(255) notnull" json:"topic_id"`
	Username  string `xorm:"varchar(255) notnull" json:"username"`
	UserURL   string `xorm:"varchar(255) notnull" json:"user_url"`
	Content   string `xorm:"longtext notnull" json:"content"`
	Time      string `xorm:"varchar(255) notnull" json:"time"`
	IP        string `xorm:"varchar(255) notnull" json:"ip"`
	DataCid   string `xorm:"varchar(255) notnull" json:"data_cid"`
	LikeCount int    `xorm:"int notnull" json:"like_count"`
}
