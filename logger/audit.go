package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// LogType defines the type of action being logged
type LogType string

const (
	ActionUpload   LogType = "UPLOAD"
	ActionDownload LogType = "DOWNLOAD"
	ActionDelete   LogType = "DELETE"
	ActionError    LogType = "ERROR"
)

// AuditLog represents the structure of an audit log entry
type AuditLog struct {
	Timestamp   time.Time `json:"timestamp"`
	Action      LogType   `json:"action"`
	ArtifactUUID string    `json:"artifact_uuid,omitempty"`
	ClientIP    string    `json:"client_ip"`
	UserSession string    `json:"user_session,omitempty"` // For future use
	Details     string    `json:"details,omitempty"`
	Status      string    `json:"status"` // SUCCESS / FAILED
}

// LoggerInterface defines the contract for different logging implementations
type LoggerInterface interface {
	Log(entry AuditLog) error
}

// InternalLogger logs to a local file or stdout (JSON format)
type InternalLogger struct {
	logger *log.Logger
}

func NewInternalLogger() *InternalLogger {
	// For production, maybe write to a file like audit.log
	// For now, we log to stdout but with a specific prefix/format
	return &InternalLogger{
		logger: log.New(os.Stdout, "[AUDIT] ", 0),
	}
}

func (l *InternalLogger) Log(entry AuditLog) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	l.logger.Println(string(data))
	return nil
}

// ExternalLogger is a placeholder for external SIEM/Log service integration
type ExternalLogger struct {
	ServiceURL string
}

func NewExternalLogger(url string) *ExternalLogger {
	return &ExternalLogger{ServiceURL: url}
}

func (l *ExternalLogger) Log(entry AuditLog) error {
	// Simulate sending to external service
	// In real implementation, this would make an HTTP POST or similar
	fmt.Printf("[EXTERNAL LOG] Sending to %s: %+v\n", l.ServiceURL, entry)
	return nil
}

// Global Logger instance
var Instance LoggerInterface

// Switch for log mode
const (
	ModeInternal = "INTERNAL"
	ModeExternal = "EXTERNAL"
)

// InitLogger initializes the global logger based on mode
func InitLogger(mode string, externalURL string) {
	switch mode {
	case ModeExternal:
		Instance = NewExternalLogger(externalURL)
		log.Println("Logger initialized in EXTERNAL mode")
	default:
		Instance = NewInternalLogger()
		log.Println("Logger initialized in INTERNAL mode")
	}
}

// Helper function to record a log easily
func Record(action LogType, uuid, ip, session, status, details string) {
	if Instance == nil {
		// Fallback if not initialized
		InitLogger(ModeInternal, "")
	}
	
	entry := AuditLog{
		Timestamp:    time.Now(),
		Action:       action,
		ArtifactUUID: uuid,
		ClientIP:     ip,
		UserSession:  session,
		Status:       status,
		Details:      details,
	}
	
	if err := Instance.Log(entry); err != nil {
		log.Printf("Failed to write audit log: %v", err)
	}
}
