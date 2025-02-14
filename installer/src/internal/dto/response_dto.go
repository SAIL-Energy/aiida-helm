package dto

type ResponseDto struct {
	Success bool        `json:"success,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func NewSuccessResponseDto() ResponseDto {
	return ResponseDto{Success: true}
}

func NewSuccessDataResponseDto(data interface{}) ResponseDto {
	return ResponseDto{Success: true, Data: data}
}

func NewErrorResponseDto(error string) ResponseDto {
	return ResponseDto{Success: false, Error: error}
}
