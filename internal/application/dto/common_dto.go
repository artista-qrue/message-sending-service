package dto

type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid request"`
	Message string `json:"message,omitempty" example:"Detailed error message"`
	Code    int    `json:"code,omitempty" example:"400"`
}
type SuccessResponse struct {
	Message string      `json:"message" example:"Operation completed successfully"`
	Data    interface{} `json:"data,omitempty"`
}

type PaginationQuery struct {
	Page  int `form:"page,default=1" binding:"min=1" example:"1"`
	Limit int `form:"limit,default=10" binding:"min=1,max=100" example:"10"`
}

func NewErrorResponse(err string, message string, code int) ErrorResponse {
	return ErrorResponse{
		Error:   err,
		Message: message,
		Code:    code,
	}
}

func NewSuccessResponse(message string, data interface{}) SuccessResponse {
	return SuccessResponse{
		Message: message,
		Data:    data,
	}
}
