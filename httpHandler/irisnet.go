package httpHandler

import (
	"encoding/json"
	"github.com/google/uuid"
	"irisStructure/errors"
	"irisStructure/logging"
	"net/http"
	"strconv"
	"time"
)

var (
	mapErrorStatusToHttp = map[string]int{
		errors.ERROR_BAD_REQUEST:            http.StatusBadRequest,          // 400
		errors.ERROR_BAD_RESPONSE:           http.StatusNotAcceptable,       // 406
		// ERROR_OUT_OF_DATE:            http.StatusPreconditionFailed,  // 412
		// ERROR_UNSUPPORTED_MEDIA_TYPE: http.StatusBadRequest,          // 400
		errors.ERROR_FORBIDDEN:              http.StatusForbidden,           // 403
		errors.ERROR_INTERNAL_SERVICE:       http.StatusInternalServerError, // 500
		errors.ERROR_NOT_FOUND:              http.StatusNotFound,            // 404
		errors.ERROR_PRECONDITION_FAILED:    http.StatusPreconditionFailed,  // 412
		errors.ERROR_TIMEOUT:                http.StatusGatewayTimeout,      // 504
		errors.ERROR_UNAUTHORIZED:           http.StatusUnauthorized,        // 401
	}
)

func ErrorCodeToStatusCode(errorCode string) int {
	statusCode, ok := mapErrorStatusToHttp[errorCode]
	if ok {
		return statusCode
	}
	return http.StatusInternalServerError
}

type HttpRequestHandlerFunc func(w http.ResponseWriter, r *http.Request, context logging.IrisLogContext) *errors.IrisError

func HttpRequestHandler(h HttpRequestHandlerFunc, logger logging.IrisLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		context := logging.IrisLogContext{
			CorrelationId: uuid.New().String(),
			UserId: "abc123",
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
