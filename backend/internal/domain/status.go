package domain

type Status string

const (
	StatusQueued        Status = "queued"
	StatusAnalyzing     Status = "analyzing"
	StatusPlanning      Status = "planning"
	StatusCoding        Status = "coding"
	StatusAssets        Status = "assets"
	StatusReviewing     Status = "reviewing"
	StatusAssembling    Status = "assembling"
	StatusChecking      Status = "checking"
	StatusReady         Status = "ready"
	StatusFailed        Status = "failed"
	StatusPatching      Status = "patching"
	StatusNeedsDecision Status = "needs_decision"
)

var statusTransitions = map[Status][]Status{
	StatusQueued:        {StatusAnalyzing},
	StatusAnalyzing:     {StatusPlanning, StatusFailed},
	StatusPlanning:      {StatusCoding, StatusFailed},
	StatusCoding:        {StatusAssets, StatusFailed},
	StatusAssets:        {StatusReviewing, StatusFailed},
	StatusReviewing:     {StatusAssembling, StatusCoding, StatusFailed, StatusNeedsDecision},
	StatusNeedsDecision: {},
	StatusAssembling:    {StatusChecking, StatusFailed},
	StatusChecking:      {StatusReady, StatusFailed, StatusPatching},
	StatusReady:         {StatusPatching, StatusFailed},
	StatusPatching:      {StatusChecking, StatusFailed, StatusReady},
	StatusFailed:        {},
}

func (s Status) String() string {
	return string(s)
}

func (s Status) IsTerminal() bool {
	return s == StatusReady || s == StatusFailed
}

var pipelineOrder = []Status{
	StatusAnalyzing,
	StatusPlanning,
	StatusCoding,
	StatusAssets,
	StatusReviewing,
	StatusAssembling,
	StatusChecking,
}

func (s Status) pipelineIndex() int {
	for index, stage := range pipelineOrder {
		if stage == s {
			return index
		}
	}

	return -1
}

func (s Status) Before(other Status) bool {
	left, right := s.pipelineIndex(), other.pipelineIndex()

	if left < 0 || right < 0 {
		return false
	}

	return left < right
}

func (s Status) Next() Status {
	index := s.pipelineIndex()
	if index < 0 || index+1 >= len(pipelineOrder) {
		return StatusAnalyzing
	}

	return pipelineOrder[index+1]
}

func (s Status) IsPipeline() bool {
	switch s {
	case StatusAnalyzing, StatusPlanning, StatusCoding, StatusAssets,
		StatusReviewing, StatusAssembling, StatusChecking:
		return true
	default:
		return false
	}
}

func (s Status) IsActive() bool {
	_, ok := statusTransitions[s]

	return ok && !s.IsTerminal()
}

func canTransition(from, to Status) bool {
	for _, allowed := range statusTransitions[from] {
		if allowed == to {
			return true
		}
	}

	return false
}
