package pocketClient

type Radio struct {
	Id             int64  `xorm:"pk autoincr"`
	LiveId         int64  `xorm:"BigInt(20) unique" json:"live_id"`
	LiveType       string `xorm:"varchar(255)" json:"live_type"`
	OnlineNum      int    `xorm:"int" json:"online_num"`
	MsgFilePath    string `xorm:"varchar(255)" json:"msg_file_path"`
	OwnerName      string `xorm:"varchar(255)" json:"owner_name"`
	Title          string `xorm:"text" json:"title"`
	PlayStreamPath string `xorm:"varchar(255)" json:"play_stream_path"`
	Ctime          int    `xorm:"int" json:"ctime"`
}

type Album struct {
	Id        int64  `xorm:"pk autoincr"`
	OwnerName string `xorm:"varchar(255)" json:"owner_name"`
	Ctime     int64  `xorm:"BigInt(20)" json:"ctime"`
	FileType  string `xorm:"varchar(255)" json:"file_type"`
	Url       string `xorm:"varchar(255) unique" json:"url"`
	State     string `xorm:"varchar(255)" json:"state"`
	Money     int    `xorm:"int" json:"money"`
	Total     int    `xorm:"int" json:"total"`
	Sold      int    `xorm:"int" json:"sold"`
}
