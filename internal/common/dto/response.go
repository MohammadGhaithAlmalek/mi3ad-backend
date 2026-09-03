package dto

// ResponseStatus mirrors the NestJS ResponseStatus enum.
type ResponseStatus string

const (
	StatusSuccess ResponseStatus = "success"
	StatusError   ResponseStatus = "error"
)

// MessageResponse is the base shape every other response embeds.
// Equivalent of MessageResponseDto.
type MessageResponse struct {
	Status  ResponseStatus `json:"status"`
	Message string         `json:"message"`
}

func NewMessageResponse(status ResponseStatus, message string) MessageResponse {
	return MessageResponse{Status: status, Message: message}
}

// NewErrorResponse is a shortcut for an error-status message response.
func NewErrorResponse(message string) MessageResponse {
	return NewMessageResponse(StatusError, message)
}

// DataResponse wraps a single item. Equivalent of DataResponse<T>.
type DataResponse[T any] struct {
	MessageResponse
	Data T `json:"data"`
}

func NewDataResponse[T any](data T, message ...string) DataResponse[T] {
	msg := "Data fetched successfully"
	if len(message) > 0 {
		msg = message[0]
	}
	return DataResponse[T]{
		MessageResponse: NewMessageResponse(StatusSuccess, msg),
		Data:            data,
	}
}

// DataArrayResponse wraps a list with a total count. Equivalent of DataArrayResponse<T>.
type DataArrayResponse[T any] struct {
	MessageResponse
	Data         []T `json:"data"`
	TotalResults int `json:"totalResults"`
}

func NewDataArrayResponse[T any](data []T, message ...string) DataArrayResponse[T] {
	msg := "Data array fetched successfully"
	if len(message) > 0 {
		msg = message[0]
	}
	return DataArrayResponse[T]{
		MessageResponse: NewMessageResponse(StatusSuccess, msg),
		Data:            data,
		TotalResults:    len(data),
	}
}

// PagingMeta mirrors IPagingMeta / PagingMeta.
type PagingMeta struct {
	Total       int  `json:"total"`
	LastPage    int  `json:"lastPage"`
	CurrentPage int  `json:"currentPage"`
	PerPage     int  `json:"perPage"`
	Prev        *int `json:"prev"`
	Next        *int `json:"next"`
}

// PaginatedResponse wraps a paginated list. Equivalent of PaginatedResponseDto<T>.
type PaginatedResponse[T any] struct {
	MessageResponse
	Data []T        `json:"data"`
	Meta PagingMeta `json:"meta"`
}

func NewPaginatedResponse[T any](data []T, meta PagingMeta, message ...string) PaginatedResponse[T] {
	msg := "Paginated data fetched successfully"
	if len(message) > 0 {
		msg = message[0]
	}
	return PaginatedResponse[T]{
		MessageResponse: NewMessageResponse(StatusSuccess, msg),
		Data:            data,
		Meta:            meta,
	}
}