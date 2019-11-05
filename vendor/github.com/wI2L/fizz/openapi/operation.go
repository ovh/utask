package openapi

// OperationInfo represents the informations of an operation
// that will be used when generating the OpenAPI specification.
type OperationInfo struct {
	ID                string
	StatusCode        int
	StatusDescription string
	Headers           []*ResponseHeader
	Summary           string
	Description       string
	Deprecated        bool
	InputModel        interface{}
	Responses         []*OperationReponse
}

// ResponseHeader represents a single header that
// may be returned with an operation response.
type ResponseHeader struct {
	Name        string
	Description string
	Model       interface{}
}

// OperationReponse represents a single response of an
// API operation.
type OperationReponse struct {
	// The response code can be "default"
	// according to OAS3 spec.
	Code        string
	Description string
	Model       interface{}
	Headers     []*ResponseHeader
}
