package httpclient

// ResponseData represents the response from an HTTP request.
type ResponseData struct {
	Status     int                    // HTTP status code
	Body       string                 // Raw response body
	JSONParsed map[string]interface{} // Parsed JSON response (if applicable)
}
