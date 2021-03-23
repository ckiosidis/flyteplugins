package core

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flytestdlib/storage"
)

// An interface to access the TaskInformation
type TaskReader interface {
	// Returns the core TaskTemplate
	Read(ctx context.Context) (*core.TaskTemplate, error)
}

// An interface that is passed to every plugin invocation. It carries all meta and contextual information for the current
// task execution
type TaskExecutionContext interface {
	AllocateResource(ctx context.Context, namespace ResourceNamespace, allocationToken string, constraintsSpec ResourceConstraintsSpec) (AllocationStatus, error)
	// During execution time, after an outstanding request is completed, the plugin need to use ReleaseResource() to release the allocation of the corresponding token
	// from the token pool in order to gain back the quota taken by the token
	ReleaseResource(ctx context.Context, namespace ResourceNamespace, allocationToken string) error

	// Returns a secret manager that can retrieve configured secrets for this plugin
	SecretManager() SecretManager

	// Returns a method that allows a plugin to indicate that the task has a new update and can be invoked again to check for updates
	TaskRefreshIndicator() SignalAsync

	// Returns the max allowed dataset size that the outputwriter will accept
	MaxDatasetSizeBytes() int64

	// Returns a handle to the currently configured storage backend that can be used to communicate with the tasks or write metadata
	DataStore() *storage.DataStore

	// Returns a reader that retrieves previously stored plugin internal state. the state itself is immutable
	PluginStateReader() PluginStateReader

	// Returns a TaskReader, to retrieve task details
	TaskReader() TaskReader

	// Returns an input reader to retrieve input data
	InputReader() io.InputReader

	// Returns a handle to the Task's execution metadata.
	TaskExecutionMetadata() TaskExecutionMetadata

	// Provides an output sync of type io.OutputWriter
	OutputWriter() io.OutputWriter

	// Get a handle to the PluginStateWriter. Any mutations to the plugins internal state can be persisted using this
	// These mutation will be visible in the next round
	PluginStateWriter() PluginStateWriter

	// Get a handle to catalog client
	Catalog() catalog.AsyncClient

	// Returns a handle to the Task events recorder, which get stored in the Admin.
	EventsRecorder() EventsRecorder
}

// A simple fire-and-forget func
type SignalAsync func(ctx context.Context)

// Task events recorder, which get stored in the Admin. If this is invoked multiple times,
// multiple events will be sent to Admin. It is not recommended that one uses this interface, a transition will trigger an auto event to admin
type EventsRecorder interface {
	RecordRaw(ctx context.Context, ev PhaseInfo) error
}
