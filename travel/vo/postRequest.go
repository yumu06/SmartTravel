package vo

type PostRequest struct {
	Title   string `json:"title"`
	HeadImg string `json:"head_img"`
	Content string `json:"content"`
}
