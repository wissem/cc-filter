package hooks

type HookProcessor interface {
	CanHandle(input map[string]interface{}) bool
	Process(input map[string]interface{}) (string, error)
}

type Registry struct {
	processors []HookProcessor
}

func NewRegistry() *Registry {
	return &Registry{
		processors: make([]HookProcessor, 0),
	}
}

func (r *Registry) Register(processor HookProcessor) {
	r.processors = append(r.processors, processor)
}

func (r *Registry) Process(input map[string]interface{}) (string, bool, error) {
	for _, processor := range r.processors {
		if processor.CanHandle(input) {
			result, err := processor.Process(input)
			if err != nil {
				return "", true, err // Return the error, mark as handled
			}
			return result, true, nil
		}
	}
	return "", false, nil
}