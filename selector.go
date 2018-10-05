package makroud

import (
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// Selector contains a pool of drivers indexed by their name.
type Selector struct {
	mutex          sync.RWMutex
	configurations map[string]*ClientOptions
	connections    map[string]Driver
}

// NewSelector returns a new selector containing a pool of drivers with given configuration.
func NewSelector(configurations map[string]*ClientOptions) (*Selector, error) {
	connections := map[string]Driver{}

	selector := &Selector{
		configurations: configurations,
		connections:    connections,
	}

	return selector, nil
}

// NewSelectorWithDriver returns a new selector containing the given connection.
func NewSelectorWithDriver(driver Driver) (*Selector, error) {
	selector := &Selector{
		configurations: map[string]*ClientOptions{},
		connections: map[string]Driver{
			"master": driver,
		},
	}

	return selector, nil
}

// Using returns the underlying drivers if it's alias exists.
func (selector *Selector) Using(alias string) (Driver, error) {
	alias = strings.ToLower(alias)

	selector.mutex.RLock()
	connection, found := selector.connections[alias]
	selector.mutex.RUnlock()

	if found {
		return connection, nil
	}

	selector.mutex.Lock()
	defer selector.mutex.Unlock()

	connection, found = selector.connections[alias]
	if found {
		return connection, nil
	}

	for name, configuration := range selector.configurations {
		if alias == strings.ToLower(name) {

			connection, err := NewWithOptions(configuration)
			if err != nil {
				return nil, err
			}

			selector.connections[alias] = connection

			return connection, nil
		}
	}

	return nil, errors.Wrapf(ErrSelectorNotFoundConnection, "connection alias '%s' not found", alias)
}

// RetryAliases is an helper calling Retry with a list of aliases.
func (selector *Selector) RetryAliases(handler func(Driver) error, aliases ...string) error {
	drivers := []Driver{}

	for _, alias := range aliases {
		connection, err := selector.Using(alias)
		if err != nil {
			continue
		}

		drivers = append(drivers, connection)
	}

	return Retry(handler, drivers...)
}

// RetryMaster is an helper calling RetryAliases with a slave then a master connection.
func (selector *Selector) RetryMaster(handler func(Driver) error) error {
	return selector.RetryAliases(handler, "slave", "master")
}

// Close closes all drivers connections.
func (selector *Selector) Close() []error {
	selector.mutex.Lock()
	defer selector.mutex.Unlock()

	errs := []error{}

	for alias, connection := range selector.connections {
		err := connection.Close()
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "cannot close drivers connection for %s", alias))
		}
	}

	selector.connections = map[string]Driver{}

	return errs
}

// Ping checks if a connection is available.
func (selector *Selector) Ping() error {
	return selector.RetryMaster(func(driver Driver) error {
		return driver.Ping()
	})
}

// Retry execute given handler on several drivers until it succeeds on a connection.
func Retry(handler func(Driver) error, drivers ...Driver) (err error) {
	if len(drivers) == 0 {
		return errors.WithStack(ErrSelectorMissingRetryConnection)
	}

	for _, driver := range drivers {
		err = handler(driver)
		if err == nil {
			return nil
		}
	}

	return err
}
