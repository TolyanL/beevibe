package domain

type Status string

const (
	StatusQueued     Status = "queued"
	StatusAnalyzing  Status = "analyzing"
	StatusPlanning   Status = "planning"
	StatusCoding     Status = "coding"
	StatusAssets     Status = "assets"
	StatusReviewing  Status = "reviewing"
	StatusAssembling Status = "assembling"
	StatusChecking   Status = "checking"
	StatusReady      Status = "ready"
	StatusFailed     Status = "failed"
	StatusPatching   Status = "patching"
)

var statusTransitions = map[Status][]Status{
	StatusQueued:     {StatusAnalyzing},
	StatusAnalyzing:  {StatusPlanning, StatusFailed},
	StatusPlanning:   {StatusCoding, StatusFailed},
	StatusCoding:     {StatusAssets, StatusFailed},
	StatusAssets:     {StatusReviewing, StatusFailed},
	StatusReviewing:  {StatusAssembling, StatusCoding, StatusFailed},
	StatusAssembling: {StatusChecking, StatusFailed},
	StatusChecking:   {StatusReady, StatusFailed, StatusPatching},
	StatusReady:      {StatusPatching, StatusFailed},
	StatusPatching:   {StatusChecking, StatusFailed, StatusReady},
	StatusFailed:     {},
}

func (s Status) String() string {
	return string(s)
}

func (s Status) IsTerminal() bool {
	return s == StatusReady || s == StatusFailed
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
