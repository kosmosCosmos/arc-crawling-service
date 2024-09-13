package common

var (
	RankMap = map[string]int{"钻石丝瓜": 2, "金钻丝瓜": 2, "小丝瓜": 1, "白金丝瓜": 1, "金丝瓜": 1, "注册会员": 1, "银丝瓜": 1}
)

const (
	LiveTypeStreaming  = 1
	LiveTypeRadio      = 2
	LiveTypeRecording  = 3
	FileTypeImage      = 1
	FileTypeAudio      = 2
	FileTypeVideo      = 3
	StateSale          = 1
	StateHidden        = 2
	StateClosed        = 3
	DiscussionUrl      = "https://www.douban.com/group/%s/discussion?start=%d"
	TopicUrl           = "https://www.douban.com/group/topic/%s/?start=%d"
	QuestionDetailAPI  = "https://pocketapi.48.cn/idolanswer/api/idolanswer/v1/question_answer/detail"
	OwnerMessageAPI    = "https://pocketapi.48.cn/im/api/v1/team/message/list/homeowner"
	AlbumListApi       = "https://pocketapi.48.cn/idolanswer/api/idolanswer/v1/user/nft/user_nft_list"
	LiveListAPI        = "https://pocketapi.48.cn/live/api/v1/live/getLiveList"
	LiveDetailAPI      = "https://pocketapi.48.cn/live/api/v1/live/getLiveOne"
	LoginAPI           = "https://pocketapi.48.cn/user/api/v1/login/app/mobile/code"
	SendSmsAPI         = "https://pocketapi.48.cn/user/api/v1/sms/send2"
	FriendshipsURL     = "https://pocketapi.48.cn/user/api/v1/friendships/friends/id"
	IMServerJumpURL    = "https://pocketapi.48.cn/im/api/v1/im/server/jump"
	TeamLastMessageURL = "https://pocketapi.48.cn/im/api/v1/team/last/message/get"
	TeamRoomInfoURL    = "https://pocketapi.48.cn/im/api/v1/im/team/room/info"
	ShopUserUrl        = "https://shop.48.cn/Account"
	ShopLoginUrl       = "https://user.48.cn/QuickLogin/login/"
	ShopOrderUrl       = "https://shop.48.cn/TOrder/ticket_Add"
)
