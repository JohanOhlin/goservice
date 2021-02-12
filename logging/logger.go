package logging

import (
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	"github.com/microsoft/ApplicationInsights-Go/appinsights/contracts"
	"time"
)

type IrisLogger interface {
	// Log a numeric value that is not specified with a specific event.
	// Typically used to send regular reports of performance indicators.
	Metric(name string, value float64, context IrisLogContext)

	// Log a trace message with the specified severity level.
	// TrackTrace(name string, severity contracts.SeverityLevel)

	Info(code string, message string, data map[string]string, context IrisLogContext)

	Warning(code string, message string, data map[string]string, context IrisLogContext)

	// Log an exception with the specified error, which may be a string,
	// error or Stringer. The current callstack is collected
	// automatically.
	Error(code string, err interface{}, data map[string]string, context IrisLogContext)

	// Log an HTTP request with the specified method, URL, duration and
	// response code.
	// TrackRequest(method, url string, duration time.Duration, responseCode string)

	Request(method string, url string, duration time.Duration, responseCode string, clientAddress string, context IrisLogContext)

	// Log a dependency with the specified name, type, target, and
	// success status.
	// TrackRemoteDependency(name, dependencyType, target string, success bool)

	// Log an availability test result with the specified test name,
	// duration, and success status.
	// TrackAvailability(name string, duration time.Duration, success bool)

}

type IrisLogContext struct {
	CorrelationId string
	UserId        string
}

type irisLogClient struct {
	client appinsights.TelemetryClient
}

func (log irisLogClient) Metric(name string, value float64, context IrisLogContext) {
	telemetry := appinsights.NewMetricTelemetry(name, value)
	if context.UserId != "" {
		telemetry.Tags.User().SetAccountId(context.UserId)
		telemetry.Tags[contracts.UserAccountId] = context.UserId
	}
	if context.CorrelationId != "" {
		telemetry.Tags.Session().SetId(context.CorrelationId)
	}
	log.client.Track(telemetry)
}

func (log irisLogClient) Info(code string, message string, data map[string]string, context IrisLogContext) {
	telemetry := appinsights.NewTraceTelemetry(message, appinsights.Information)
	if data != nil {
		telemetry.Properties = data
	}
	if context.UserId != "" {
		telemetry.Tags.User().SetAccountId(context.UserId)
		telemetry.Tags[contracts.UserAccountId] = context.UserId
	}
	if context.CorrelationId != "" {
		telemetry.Tags.Session().SetId(context.CorrelationId)
	}

	telemetry.Properties["event_code"] = code
	log.client.Track(telemetry)
}
func (log irisLogClient) Warning(code string, message string, data map[string]string, context IrisLogContext) {
	telemetry := appinsights.NewTraceTelemetry(message, appinsights.Warning)
	if data != nil {
		telemetry.Properties = data
	}
	if context.UserId != "" {
		telemetry.Tags.User().SetAccountId(context.UserId)
		telemetry.Tags[contracts.UserAccountId] = context.UserId
	}
	if context.CorrelationId != "" {
		telemetry.Tags.Session().SetId(context.CorrelationId)
	}

	telemetry.Properties["event_code"] = code
	log.client.Track(telemetry)
}
func (log irisLogClient) Error(code string, err interface{}, data map[string]string, context IrisLogContext) {
	telemetry := newExceptionTelemetry(err, 1)
	if data != nil {
		telemetry.Properties = data
	}
	telemetry.Properties["event_code"] = code
	if context.UserId != "" {
		telemetry.Tags.User().SetAccountId(context.UserId)
		telemetry.Tags[contracts.UserAccountId] = context.UserId
	}
	if context.CorrelationId != "" {
		telemetry.Tags.Session().SetId(context.CorrelationId)
	}
	telemetry.Tags[contracts.OperationName] = "handleHomeGet"

	log.client.Track(telemetry)
}

func newExceptionTelemetry(err interface{}, skip int) *appinsights.ExceptionTelemetry {
	return &appinsights.ExceptionTelemetry{
		Error:         err,
		Frames:        appinsights.GetCallstack(2 + skip),
		SeverityLevel: appinsights.Error,
		BaseTelemetry: appinsights.BaseTelemetry{
			Timestamp:  currentClock.Now(),
			Tags:       make(contracts.ContextTags),
			Properties: make(map[string]string),
		},
		BaseTelemetryMeasurements: appinsights.BaseTelemetryMeasurements{
			Measurements: make(map[string]float64),
		},
	}
}

func (log irisLogClient) Request(method string, url string, duration time.Duration, responseCode string, clientAddress string, context IrisLogContext) {
	telemetry := appinsights.NewRequestTelemetry(method, url, duration, responseCode)

	// Note that the timestamp will be set to time.Now() minus the
	// specified duration.  This can be overridden by either manually
	// setting the Timestamp and Duration fields, or with MarkTime:
	// request.MarkTime(requestStartTime, requestEndTime)

	// Source of request
	telemetry.Source = clientAddress

	// Success is normally inferred from the responseCode, but can be overridden:
	// request.Success = responsecode == "200"

	// Request ID's are randomly generated GUIDs, but this can also be overridden:
	// telemetry.Id = "<id>"

	// Custom properties and measurements can be set here
	// request.Properties["user-agent"] = request.headers["User-agent"]
	// request.Measurements["POST size"] = float64(len(data))

	if context.UserId != "" {
		telemetry.Tags.User().SetAccountId(context.UserId)
		telemetry.Tags[contracts.UserAccountId] = context.UserId
	}
	if context.CorrelationId != "" {
		telemetry.Tags.Session().SetId(context.CorrelationId)
	}

	// Finally track it
	log.client.Track(telemetry)
}

func NewLogger(instrumentationKey string, serviceName string) IrisLogger {
	initClock()
	telemetryConfig := appinsights.NewTelemetryConfiguration(instrumentationKey)
	// Configure how many items can be sent in one call to the data collector:
	telemetryConfig.MaxBatchSize = 8192
	// Configure the maximum delay before sending queued telemetry:
	telemetryConfig.MaxBatchInterval = 2 * time.Second
	client := appinsights.NewTelemetryClientFromConfig(telemetryConfig)
	client.Context().Tags.Cloud().SetRole(serviceName)

	return &irisLogClient{
		client,
	}
}
