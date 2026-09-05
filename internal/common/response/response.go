// Package response provides ergonomic, status-code-aware wrappers around
// the pure data shapes in internal/common/dto. Handlers call these instead
// of manually pairing an http.Status with a dto envelope - which removes
// a whole class of bugs where the status code and the envelope disagree
// (e.g. returning 201 wrapped in an "error" shaped body).
package response

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"

	"mi3ad/internal/common/dto"
)

// --- success ---------------------------------------------------------

// OK returns a single item with 200.
func OK[T any](c *echo.Context, data T, message ...string) error {
	return c.JSON(http.StatusOK, dto.NewDataResponse(data, message...))
}

// Created returns a single item with 201.
func Created[T any](c *echo.Context, data T, message ...string) error {
	return c.JSON(http.StatusCreated, dto.NewDataResponse(data, message...))
}

// OKArray returns a list with a total count, 200.
func OKArray[T any](c *echo.Context, data []T, message ...string) error {
	return c.JSON(http.StatusOK, dto.NewDataArrayResponse(data, message...))
}

// OKPaged returns a paginated list, 200. Pass the total row count from
// your query and the page/perPage you queried with - the paging math
// (lastPage/prev/next) is computed for you.
func OKPaged[T any](c *echo.Context, data []T, total int64, page, perPage int, message ...string) error {
	meta := dto.NewPagingMeta(total, page, perPage)
	return c.JSON(http.StatusOK, dto.NewPaginatedResponse(data, meta, message...))
}

// Message returns a plain status+message body with no data payload,
// e.g. after a delete. Caller picks the status code (200 for a
// completed delete, 202 for accepted-but-async, etc).
func Message(c *echo.Context, status int, message string) error {
	return c.JSON(status, dto.NewMessageResponse(dto.StatusSuccess, message))
}

// NoContent returns a bare 204 - no body at all.
func NoContent(c *echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

// --- errors ------------------------------------------------------------

func BadRequest(c *echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, dto.NewErrorResponse(message))
}

func Unauthorized(c *echo.Context, message string) error {
	return c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(message))
}

func Forbidden(c *echo.Context, message string) error {
	return c.JSON(http.StatusForbidden, dto.NewErrorResponse(message))
}

func NotFound(c *echo.Context, message string) error {
	return c.JSON(http.StatusNotFound, dto.NewErrorResponse(message))
}

func Conflict(c *echo.Context, message string) error {
	return c.JSON(http.StatusConflict, dto.NewErrorResponse(message))
}

// InternalError never takes a message from the caller on purpose -
// don't leak internal error detail to the client. Log the real err
// where you call this instead.
func InternalError(c *echo.Context) error {
	return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("internal server error"))
}

// ValidationError formats go-playground/validator errors into a
// field -> failed-tag map, e.g. {"Email": "email", "Password": "min"}.
// Falls back to a plain BadRequest if err isn't a validator error.
func ValidationError(c *echo.Context, err error) error {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return BadRequest(c, err.Error())
	}

	fields := make(map[string]string, len(ve))
	for _, fe := range ve {
		fields[fe.Field()] = fe.Tag()
	}

	return c.JSON(http.StatusUnprocessableEntity, dto.ValidationErrorResponse{
		MessageResponse: dto.NewMessageResponse(dto.StatusError, "validation failed"),
		Fields:          fields,
	})
}
