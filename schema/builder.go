package schema

import "errors"

// ServiceBuilder provides a fluent API for defining services.
type ServiceBuilder struct {
	service Service
	err     error
}

// NewService creates a new ServiceBuilder with the given service name.
func NewService(name string) *ServiceBuilder {
	return &ServiceBuilder{
		service: Service{Name: name},
	}
}

// Package sets the package name for the service.
func (b *ServiceBuilder) Package(pkg string) *ServiceBuilder {
	if b.err != nil {
		return b
	}
	b.service.Package = pkg
	return b
}

// Backend sets the backend name that implements this service.
func (b *ServiceBuilder) Backend(backend string) *ServiceBuilder {
	if b.err != nil {
		return b
	}
	b.service.Backend = backend
	return b
}

// Description sets a human-readable description for the service.
func (b *ServiceBuilder) Description(desc string) *ServiceBuilder {
	if b.err != nil {
		return b
	}
	b.service.Description = desc
	return b
}

// Method adds a method to the service.
func (b *ServiceBuilder) Method(m Method) *ServiceBuilder {
	if b.err != nil {
		return b
	}
	if m.Name == "" {
		b.err = errors.New("method name required")
		return b
	}
	b.service.Methods = append(b.service.Methods, m)
	return b
}

// Build returns the constructed Service or an error if validation fails.
func (b *ServiceBuilder) Build() (Service, error) {
	if b.err != nil {
		return Service{}, b.err
	}
	if b.service.Name == "" {
		return Service{}, errors.New("service name required")
	}
	return b.service, nil
}

// MustBuild returns the constructed Service or panics if validation fails.
func (b *ServiceBuilder) MustBuild() Service {
	svc, err := b.Build()
	if err != nil {
		panic(err)
	}
	return svc
}
