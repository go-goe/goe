package handler

type Response[T any] struct {
	Status      int    `json:"status"`
	Message     string `json:"message"`
	Description string `json:"description"`
	Data        T      `json:"data"`
}

type ResponsePost[T any] struct {
	ID T `json:"id"`
}
