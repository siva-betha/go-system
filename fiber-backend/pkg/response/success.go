package response

func OK(data any) map[string]any {
	return map[string]any{
		"success": true,
		"data":    data,
	}
}
