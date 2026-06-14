package core

import "time"

type Remediation struct {
	Step        int
	Action      string
	Command     string
	Priority    int
	Reference   string
	Risk        string
	AutoFixable bool
}

type Evidence struct {
	Source    string
	Raw       string
	Interpret string
}

type CheckResult struct {
	ID          string
	Status      CheckStatus
	Severity    Severity
	Category    Component
	Message     string
	Details     map[string]interface{}
	Metrics     map[string]float64
	Timestamp   time.Time
	Duration    time.Duration
	Error       error
	Remediation []Remediation
	Evidence    []Evidence
}

type AggregatedResult struct {
	Timestamp     time.Time
	Duration      time.Duration
	TotalChecks   int
	PassedChecks  int
	FailedChecks  int
	ErrorChecks   int
	SkippedChecks int
	HealthScore   float64
	Results       map[Component][]*CheckResult
	AllResults    []*CheckResult
}

func NewAggregatedResult() *AggregatedResult {
	return &AggregatedResult{
		Timestamp: time.Now(),
		Results:   make(map[Component][]*CheckResult),
	}
}

func (ar *AggregatedResult) AddResult(r *CheckResult) {
	ar.AllResults = append(ar.AllResults, r)
	ar.Results[r.Category] = append(ar.Results[r.Category], r)
	switch r.Status {
	case StatusPass:
		ar.PassedChecks++
	case StatusFail:
		ar.FailedChecks++
	case StatusError:
		ar.ErrorChecks++
	case StatusSkip:
		ar.SkippedChecks++
	}
}

func (ar *AggregatedResult) CalculateHealthScore() float64 {
	if ar.TotalChecks == 0 {
		return 0
	}
	score := 0.0
	for _, r := range ar.AllResults {
		switch r.Status {
		case StatusPass:
			score += 100
		case StatusFail:
			switch r.Severity {
			case SeverityCritical, SeverityFatal:
				score += 0
			case SeverityWarning:
				score += 50
			default:
				score += 75
			}
		case StatusError:
			score += 25
		case StatusSkip:
			score += 100
		}
	}
	return score / float64(ar.TotalChecks)
}

type Severity int

const (
	SeverityNone Severity = iota
	SeverityInfo
	SeverityWarning
	SeverityCritical
	SeverityFatal
)

func (s Severity) String() string {
	switch s {
	case SeverityNone:
		return "none"
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityCritical:
		return "critical"
	case SeverityFatal:
		return "fatal"
	default:
		return "unknown"
	}
}

func (s Severity) Color() string {
	switch s {
	case SeverityNone:
		return "white"
	case SeverityInfo:
		return "cyan"
	case SeverityWarning:
		return "yellow"
	case SeverityCritical:
		return "red"
	case SeverityFatal:
		return "magenta"
	default:
		return "white"
	}
}

type CheckStatus int

const (
	StatusUnknown CheckStatus = iota
	StatusPass
	StatusFail
	StatusError
	StatusSkip
)

func (s CheckStatus) String() string {
	switch s {
	case StatusUnknown:
		return "unknown"
	case StatusPass:
		return "pass"
	case StatusFail:
		return "fail"
	case StatusError:
		return "error"
	case StatusSkip:
		return "skip"
	default:
		return "unknown"
	}
}

type Component string

const (
	ComponentCPU       Component = "cpu"
	ComponentMemory    Component = "memory"
	ComponentDisk      Component = "disk"
	ComponentNetwork   Component = "network"
	ComponentKernel    Component = "kernel"
	ComponentServices  Component = "services"
	ComponentSecurity  Component = "security"
	ComponentLogs      Component = "logs"
	ComponentHardware  Component = "hardware"
	ComponentUpdates   Component = "updates"
	ComponentContainers Component = "containers"
)

func AllComponents() []Component {
	return []Component{
		ComponentCPU, ComponentMemory, ComponentDisk,
		ComponentNetwork, ComponentKernel, ComponentServices,
		ComponentSecurity, ComponentLogs, ComponentHardware,
		ComponentUpdates, ComponentContainers,
	}
}

type Threshold struct {
	Warning  float64
	Critical float64
}

type Metric struct {
	Name      string
	Value     float64
	Unit      string
	Timestamp time.Time
	Labels    map[string]string
}

type EventType string

const (
	EventCheckStart    EventType = "check.start"
	EventCheckComplete EventType = "check.complete"
	EventCheckFail     EventType = "check.fail"
	EventAlert         EventType = "alert"
	EventThreshold     EventType = "threshold.breach"
	EventSnapshot      EventType = "snapshot"
	EventBaseline      EventType = "baseline"
	EventError         EventType = "error"
)

type Event struct {
	ID        string
	Type      EventType
	Source    string
	Severity  Severity
	Message   string
	Data      map[string]interface{}
	Timestamp time.Time
}

type ResultFilter struct {
	CheckIDs   []string
	Components []Component
	Statuses   []CheckStatus
	Severity   Severity
	Since      time.Time
	Until      time.Time
	Limit      int
	Offset     int
}

type EventFilter struct {
	Types    []EventType
	Severity Severity
	Since    time.Time
	Until    time.Time
	Limit    int
	Offset   int
}
