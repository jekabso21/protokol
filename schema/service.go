package schema

// Service represents a group of related methods.
type Service struct {
	Name        string
	Package     string
	Methods     []Method
	Description string
	Backend     string
}

func (s Service) MethodByName(name string) (Method, bool) {
	for _, m := range s.Methods {
		if m.Name == name {
			return m, true
		}
	}
	return Method{}, false
}

// Schema holds the complete API definition.
type Schema struct {
	Services []Service
	Types    map[string]Type
}

func NewSchema() *Schema {
	return &Schema{
		Types: make(map[string]Type),
	}
}

func (s *Schema) AddService(svc Service) {
	s.Services = append(s.Services, svc)
}

func (s *Schema) ServiceByName(name string) (Service, bool) {
	for _, svc := range s.Services {
		if svc.Name == name {
			return svc, true
		}
	}
	return Service{}, false
}

func (s *Schema) RegisterType(name string, t Type) {
	s.Types[name] = t
}

func (s *Schema) LookupType(name string) (Type, bool) {
	t, ok := s.Types[name]
	return t, ok
}
