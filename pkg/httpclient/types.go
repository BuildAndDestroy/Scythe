package httpclient

// ResponseData holds the server's response.
type ResponseData struct {
	Status     int                    // HTTP status code
	Body       string                 // Raw response body
	JSONParsed map[string]interface{} // Parsed JSON response (if applicable)
}
