package goservice

import (
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"strconv"
	"time"
)

var (
	mapErrorStatusToHttp = map[string]int{
		ERROR_BAD_REQUEST:  http.StatusBadRequest,    // 400
		ERROR_BAD_RESPONSE: http.StatusNotAcceptable, // 406
		// ERROR_OUT_OF_DATE:            http.StatusPreconditionFailed,  // 412
		// ERROR_UNSUPPORTED_MEDIA_TYPE: http.StatusBadRequest,          // 400
		ERROR_FORBIDDEN:           http.StatusForbidden,           // 403
		ERROR_INTERNAL_SERVICE:    http.StatusInternalServerError, // 500
		ERROR_NOT_FOUND:           http.StatusNotFound,            // 404
		ERROR_PRECONDITION_FAILED: http.StatusPreconditionFailed,  // 412
		ERROR_TIMEOUT:             http.StatusGatewayTimeout,      // 504
		ERROR_UNAUTHORIZED:        http.StatusUnauthorized,        // 401
	}
)

func ErrorCodeToStatusCode(errorCode string) int {
	statusCode, ok := mapErrorStatusToHttp[errorCode]
	if ok {
		return statusCode
	}
	return http.StatusInternalServerError
}

type HttpRequestHandlerFunc func(w http.ResponseWriter, r *http.Request, context IrisLogContext) *IrisError

func HttpRequestHandler(h HttpRequestHandlerFunc, logger IrisLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		context := IrisLogContext{
			CorrelationId: uuid.New().String(),
			UserId:        "abc123",
		}
		start := time.Now()
		err := h(w, r, context)
		duration := time.Since(start)
		responseCode := 200
		if err != nil {
			responseCode = ErrorCodeToStatusCode(err.TypeCode)
		}

		clientAddress := r.Header.Get("X-FORWARDED-FOR")
		scheme := r.URL.Scheme
		if scheme == "" {
			scheme = "https"
		}
		url := scheme + "://" + r.Host + r.RequestURI
		responseCodeString := strconv.Itoa(responseCode)
		logger.Request(r.Method, url, duration, responseCodeString, clientAddress, context)

		if err != nil {
			w.WriteHeader(responseCode)
			json.NewEncoder(w).Encode(err)
			return
		}
	}
}
