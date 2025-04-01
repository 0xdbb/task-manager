package server

// HandleError represents the structure of error messages in API responses
type ErrorResponse struct {
	Status   string `json:"status"`
	Messsage string `json:"message"`
	Code     int    `json:"code"`
}

type Message struct {
	Message string `json:"message" example:"success"`
	TaskID string	`json:"task_id" binding:"omitempty" example:"1233-flf4djf-alsdik"`
}

func HandleError(err error, code int, message ...string) ErrorResponse {
	var msg string

	if err != nil {
		msg = err.Error()
	} else {
		msg = "Unknown error"
	}

	if len(message) > 0 {
		msg = message[0] + ": " + msg // Override error message if provided
	}

	return ErrorResponse{
		Status:   "error",
		Messsage: msg,
		Code:     code,
	}
}

func HandleMessage(message string) Message {
	return Message{Message: message}
}
