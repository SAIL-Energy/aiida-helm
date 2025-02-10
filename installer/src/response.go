package main

type Response struct {
	Success bool        `json:"success,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func NewSuccessResponse() Response {
	return Response{Success: true}
}

func NewSuccessDataResponse(data interface{}) Response {
	return Response{Success: true, Data: data}
}

func NewErrorResponse(error error) Response {
	return Response{Success: false, Error: error.Error()}
}
