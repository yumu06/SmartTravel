package vo

type Code2SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type IdentifyCode struct {
	EncryptedData string `json:"encrypted_data" binding:"required"`
	Iv            string `json:"iv" binding:"required"`
	SessionKey    string `json:"session_id" binding:"required"`
}

type UpdateUserRequest struct {
	Telephone string `json:"telephone"`
	NickName  string `json:"nick_name"`
	Motto     string `json:"motto"`
	Gender    int    `json:"gender"` //0表示男、1表示女
}

type ShowUserRequest struct {
	Telephone string `json:"telephone"`
	NickName  string `json:"nick_name"`
	Motto     string `json:"motto"`
	Gender    int    `json:"gender"` //0表示男、1表示女
}

type CreateFootRequest struct {
	Title            string `json:"title"`
	Origin           string `json:"origin"`
	OriginName       string `json:"origin_name"`
	Destinations     string `json:"destinations"`
	DestinationNames string `json:"destination_names"`
	Mode             string `json:"mode"`
	RouteResult      string `json:"routeResult"`
}

type UserSummary struct {
	ID        uint64 `json:"id"`
	NickName  string `json:"nick_name"`
	Motto     string `json:"motto"`
	Gender    int    `json:"gender"`
	AvatarURL string `json:"avatar_url"`
}

type ChatSendRequest struct {
	ToUserID uint64 `json:"toUserID"`
	Content  string `json:"content"`
}
