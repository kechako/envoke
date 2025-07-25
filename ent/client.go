// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/kechako/envoke/ent/migrate"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/kechako/envoke/ent/environment"
	"github.com/kechako/envoke/ent/variable"
)

// Client is the client that holds all ent builders.
type Client struct {
	config
	// Schema is the client for creating, migrating and dropping schema.
	Schema *migrate.Schema
	// Environment is the client for interacting with the Environment builders.
	Environment *EnvironmentClient
	// Variable is the client for interacting with the Variable builders.
	Variable *VariableClient
}

// NewClient creates a new client configured with the given options.
func NewClient(opts ...Option) *Client {
	client := &Client{config: newConfig(opts...)}
	client.init()
	return client
}

func (c *Client) init() {
	c.Schema = migrate.NewSchema(c.driver)
	c.Environment = NewEnvironmentClient(c.config)
	c.Variable = NewVariableClient(c.config)
}

type (
	// config is the configuration for the client and its builder.
	config struct {
		// driver used for executing database requests.
		driver dialect.Driver
		// debug enable a debug logging.
		debug bool
		// log used for logging on debug mode.
		log func(...any)
		// hooks to execute on mutations.
		hooks *hooks
		// interceptors to execute on queries.
		inters *inters
	}
	// Option function to configure the client.
	Option func(*config)
)

// newConfig creates a new config for the client.
func newConfig(opts ...Option) config {
	cfg := config{log: log.Println, hooks: &hooks{}, inters: &inters{}}
	cfg.options(opts...)
	return cfg
}

// options applies the options on the config object.
func (c *config) options(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
	if c.debug {
		c.driver = dialect.Debug(c.driver, c.log)
	}
}

// Debug enables debug logging on the ent.Driver.
func Debug() Option {
	return func(c *config) {
		c.debug = true
	}
}

// Log sets the logging function for debug mode.
func Log(fn func(...any)) Option {
	return func(c *config) {
		c.log = fn
	}
}

// Driver configures the client driver.
func Driver(driver dialect.Driver) Option {
	return func(c *config) {
		c.driver = driver
	}
}

// Open opens a database/sql.DB specified by the driver name and
// the data source name, and returns a new client attached to it.
// Optional parameters can be added for configuring the client.
func Open(driverName, dataSourceName string, options ...Option) (*Client, error) {
	switch driverName {
	case dialect.MySQL, dialect.Postgres, dialect.SQLite:
		drv, err := sql.Open(driverName, dataSourceName)
		if err != nil {
			return nil, err
		}
		return NewClient(append(options, Driver(drv))...), nil
	default:
		return nil, fmt.Errorf("unsupported driver: %q", driverName)
	}
}

// ErrTxStarted is returned when trying to start a new transaction from a transactional client.
var ErrTxStarted = errors.New("ent: cannot start a transaction within a transaction")

// Tx returns a new transactional client. The provided context
// is used until the transaction is committed or rolled back.
func (c *Client) Tx(ctx context.Context) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, ErrTxStarted
	}
	tx, err := newTx(ctx, c.driver)
	if err != nil {
		return nil, fmt.Errorf("ent: starting a transaction: %w", err)
	}
	cfg := c.config
	cfg.driver = tx
	return &Tx{
		ctx:         ctx,
		config:      cfg,
		Environment: NewEnvironmentClient(cfg),
		Variable:    NewVariableClient(cfg),
	}, nil
}

// BeginTx returns a transactional client with specified options.
func (c *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, errors.New("ent: cannot start a transaction within a transaction")
	}
	tx, err := c.driver.(interface {
		BeginTx(context.Context, *sql.TxOptions) (dialect.Tx, error)
	}).BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("ent: starting a transaction: %w", err)
	}
	cfg := c.config
	cfg.driver = &txDriver{tx: tx, drv: c.driver}
	return &Tx{
		ctx:         ctx,
		config:      cfg,
		Environment: NewEnvironmentClient(cfg),
		Variable:    NewVariableClient(cfg),
	}, nil
}

// Debug returns a new debug-client. It's used to get verbose logging on specific operations.
//
//	client.Debug().
//		Environment.
//		Query().
//		Count(ctx)
func (c *Client) Debug() *Client {
	if c.debug {
		return c
	}
	cfg := c.config
	cfg.driver = dialect.Debug(c.driver, c.log)
	client := &Client{config: cfg}
	client.init()
	return client
}

// Close closes the database connection and prevents new queries from starting.
func (c *Client) Close() error {
	return c.driver.Close()
}

// Use adds the mutation hooks to all the entity clients.
// In order to add hooks to a specific client, call: `client.Node.Use(...)`.
func (c *Client) Use(hooks ...Hook) {
	c.Environment.Use(hooks...)
	c.Variable.Use(hooks...)
}

// Intercept adds the query interceptors to all the entity clients.
// In order to add interceptors to a specific client, call: `client.Node.Intercept(...)`.
func (c *Client) Intercept(interceptors ...Interceptor) {
	c.Environment.Intercept(interceptors...)
	c.Variable.Intercept(interceptors...)
}

// Mutate implements the ent.Mutator interface.
func (c *Client) Mutate(ctx context.Context, m Mutation) (Value, error) {
	switch m := m.(type) {
	case *EnvironmentMutation:
		return c.Environment.mutate(ctx, m)
	case *VariableMutation:
		return c.Variable.mutate(ctx, m)
	default:
		return nil, fmt.Errorf("ent: unknown mutation type %T", m)
	}
}

// EnvironmentClient is a client for the Environment schema.
type EnvironmentClient struct {
	config
}

// NewEnvironmentClient returns a client for the Environment from the given config.
func NewEnvironmentClient(c config) *EnvironmentClient {
	return &EnvironmentClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `environment.Hooks(f(g(h())))`.
func (c *EnvironmentClient) Use(hooks ...Hook) {
	c.hooks.Environment = append(c.hooks.Environment, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `environment.Intercept(f(g(h())))`.
func (c *EnvironmentClient) Intercept(interceptors ...Interceptor) {
	c.inters.Environment = append(c.inters.Environment, interceptors...)
}

// Create returns a builder for creating a Environment entity.
func (c *EnvironmentClient) Create() *EnvironmentCreate {
	mutation := newEnvironmentMutation(c.config, OpCreate)
	return &EnvironmentCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Environment entities.
func (c *EnvironmentClient) CreateBulk(builders ...*EnvironmentCreate) *EnvironmentCreateBulk {
	return &EnvironmentCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *EnvironmentClient) MapCreateBulk(slice any, setFunc func(*EnvironmentCreate, int)) *EnvironmentCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &EnvironmentCreateBulk{err: fmt.Errorf("calling to EnvironmentClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*EnvironmentCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &EnvironmentCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Environment.
func (c *EnvironmentClient) Update() *EnvironmentUpdate {
	mutation := newEnvironmentMutation(c.config, OpUpdate)
	return &EnvironmentUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *EnvironmentClient) UpdateOne(e *Environment) *EnvironmentUpdateOne {
	mutation := newEnvironmentMutation(c.config, OpUpdateOne, withEnvironment(e))
	return &EnvironmentUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *EnvironmentClient) UpdateOneID(id int) *EnvironmentUpdateOne {
	mutation := newEnvironmentMutation(c.config, OpUpdateOne, withEnvironmentID(id))
	return &EnvironmentUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Environment.
func (c *EnvironmentClient) Delete() *EnvironmentDelete {
	mutation := newEnvironmentMutation(c.config, OpDelete)
	return &EnvironmentDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *EnvironmentClient) DeleteOne(e *Environment) *EnvironmentDeleteOne {
	return c.DeleteOneID(e.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *EnvironmentClient) DeleteOneID(id int) *EnvironmentDeleteOne {
	builder := c.Delete().Where(environment.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &EnvironmentDeleteOne{builder}
}

// Query returns a query builder for Environment.
func (c *EnvironmentClient) Query() *EnvironmentQuery {
	return &EnvironmentQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeEnvironment},
		inters: c.Interceptors(),
	}
}

// Get returns a Environment entity by its id.
func (c *EnvironmentClient) Get(ctx context.Context, id int) (*Environment, error) {
	return c.Query().Where(environment.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *EnvironmentClient) GetX(ctx context.Context, id int) *Environment {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryVariables queries the variables edge of a Environment.
func (c *EnvironmentClient) QueryVariables(e *Environment) *VariableQuery {
	query := (&VariableClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := e.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(environment.Table, environment.FieldID, id),
			sqlgraph.To(variable.Table, variable.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, environment.VariablesTable, environment.VariablesColumn),
		)
		fromV = sqlgraph.Neighbors(e.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *EnvironmentClient) Hooks() []Hook {
	return c.hooks.Environment
}

// Interceptors returns the client interceptors.
func (c *EnvironmentClient) Interceptors() []Interceptor {
	return c.inters.Environment
}

func (c *EnvironmentClient) mutate(ctx context.Context, m *EnvironmentMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&EnvironmentCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&EnvironmentUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&EnvironmentUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&EnvironmentDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown Environment mutation op: %q", m.Op())
	}
}

// VariableClient is a client for the Variable schema.
type VariableClient struct {
	config
}

// NewVariableClient returns a client for the Variable from the given config.
func NewVariableClient(c config) *VariableClient {
	return &VariableClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `variable.Hooks(f(g(h())))`.
func (c *VariableClient) Use(hooks ...Hook) {
	c.hooks.Variable = append(c.hooks.Variable, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `variable.Intercept(f(g(h())))`.
func (c *VariableClient) Intercept(interceptors ...Interceptor) {
	c.inters.Variable = append(c.inters.Variable, interceptors...)
}

// Create returns a builder for creating a Variable entity.
func (c *VariableClient) Create() *VariableCreate {
	mutation := newVariableMutation(c.config, OpCreate)
	return &VariableCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Variable entities.
func (c *VariableClient) CreateBulk(builders ...*VariableCreate) *VariableCreateBulk {
	return &VariableCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *VariableClient) MapCreateBulk(slice any, setFunc func(*VariableCreate, int)) *VariableCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &VariableCreateBulk{err: fmt.Errorf("calling to VariableClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*VariableCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &VariableCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Variable.
func (c *VariableClient) Update() *VariableUpdate {
	mutation := newVariableMutation(c.config, OpUpdate)
	return &VariableUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *VariableClient) UpdateOne(v *Variable) *VariableUpdateOne {
	mutation := newVariableMutation(c.config, OpUpdateOne, withVariable(v))
	return &VariableUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *VariableClient) UpdateOneID(id int) *VariableUpdateOne {
	mutation := newVariableMutation(c.config, OpUpdateOne, withVariableID(id))
	return &VariableUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Variable.
func (c *VariableClient) Delete() *VariableDelete {
	mutation := newVariableMutation(c.config, OpDelete)
	return &VariableDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *VariableClient) DeleteOne(v *Variable) *VariableDeleteOne {
	return c.DeleteOneID(v.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *VariableClient) DeleteOneID(id int) *VariableDeleteOne {
	builder := c.Delete().Where(variable.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &VariableDeleteOne{builder}
}

// Query returns a query builder for Variable.
func (c *VariableClient) Query() *VariableQuery {
	return &VariableQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeVariable},
		inters: c.Interceptors(),
	}
}

// Get returns a Variable entity by its id.
func (c *VariableClient) Get(ctx context.Context, id int) (*Variable, error) {
	return c.Query().Where(variable.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *VariableClient) GetX(ctx context.Context, id int) *Variable {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryEnvironment queries the environment edge of a Variable.
func (c *VariableClient) QueryEnvironment(v *Variable) *EnvironmentQuery {
	query := (&EnvironmentClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := v.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(variable.Table, variable.FieldID, id),
			sqlgraph.To(environment.Table, environment.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, variable.EnvironmentTable, variable.EnvironmentColumn),
		)
		fromV = sqlgraph.Neighbors(v.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *VariableClient) Hooks() []Hook {
	return c.hooks.Variable
}

// Interceptors returns the client interceptors.
func (c *VariableClient) Interceptors() []Interceptor {
	return c.inters.Variable
}

func (c *VariableClient) mutate(ctx context.Context, m *VariableMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&VariableCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&VariableUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&VariableUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&VariableDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown Variable mutation op: %q", m.Op())
	}
}

// hooks and interceptors per client, for fast access.
type (
	hooks struct {
		Environment, Variable []ent.Hook
	}
	inters struct {
		Environment, Variable []ent.Interceptor
	}
)
