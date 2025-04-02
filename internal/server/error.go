package server

// HandleError represents the structure of error messages in API responses
type ErrorResponse struct {
	Status   string `json:"status"`
	Messsage string `json:"message"`
	Code     int    `json:"code"`
}

type Message struct {
	Message string `json:"message" example:"success"`
	TaskID  string `json:"task_id" binding:"omitempty" example:"1233-flf4djf-alsdik"`
	Result  string `json:"result" binding:"omitempty" example:"Task processed successfully\nWeather for Chicago (33.44,-94.04):\n- Current: 25.6°C, overcast clouds\n- Feels like: 26.1°C\n- Humidity: 73%\n- Wind: 7.7 m/s, 200°\n\nAlerts:\n- Flood Watch: * WHAT...Flooding caused by excessive rainfall continues to be possible.\n\n* WHERE...Portions of south central and southwest Arkansas,\nincluding the following counties, in south central Arkansas,\nUnion. In southwest Arkansas, Columbia, Lafayette and Miller.\n\n* WHEN...From this evening through Sunday morning.\n\n* IMPACTS...Excessive runoff may result in flooding of rivers,\ncreeks, streams, and other low-lying and flood-prone locations.\nCreeks and streams may rise out of their banks."`
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
