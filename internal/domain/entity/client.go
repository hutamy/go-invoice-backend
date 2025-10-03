package entity

type Client struct {
	ID      uint   `json:"id"`
	UserID  uint   `json:"user_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}
