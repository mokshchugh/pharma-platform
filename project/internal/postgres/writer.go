package postgres

type Writer struct {
	client *Client
}

func NewWriter(client *Client) *Writer {
	return &Writer{
		client: client,
	}
}
