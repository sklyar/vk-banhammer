package entity

// Comment describes comment.
type Comment struct {
	ID      int    `json:"id"`
	FromID  int    `json:"from_id"`
	Date    int    `json:"date"`
	Text    string `json:"text"`
	PostID  int    `json:"post_id"`
	OwnerID int    `json:"owner_id"`
}
