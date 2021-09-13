package k8s

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/errorcollector"

	arrayCore "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"

	errors2 "github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/logger"

	corev1 "k8s.io/api/core/v1"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
)

const (
	ErrBuildPodTemplate       errors2.ErrorCode = "POD_TEMPLATE_FAILED"
	ErrReplaceCmdTemplate     errors2.ErrorCode = "CMD_TEMPLATE_FAILED"
	ErrSubmitJob              errors2.ErrorCode = "SUBMIT_JOB_FAILED"
	ErrGetTaskTypeVersion     errors2.ErrorCode = "GET_TASK_TYPE_VERSION_FAILED"
	JobIndexVarName           string            = "BATCH_JOB_ARRAY_INDEX_VAR_NAME"
	FlyteK8sArrayIndexVarName string            = "FLYTE_K8S_ARRAY_INDEX"
)

var arrayJobEnvVars = []corev1.EnvVar{
	{
		Name:  JobIndexVarName,
		Value: FlyteK8sArrayIndexVarName,
	},
}

func formatSubTaskName(_ context.Context, parentName, suffix string) (subTaskName string) {
	return fmt.Sprintf("%v-%v", parentName, suffix)
}

func ApplyPodPolicies(_ context.Context, cfg *Config, pod *corev1.Pod) *corev1.Pod {
	if len(cfg.DefaultScheduler) > 0 {
		pod.Spec.SchedulerName = cfg.DefaultScheduler
	}

	return pod
}

func applyNodeSelectorLabels(ctx context.Context, cfg *Config, pod *corev1.Pod) *corev1.Pod {
	if len(cfg.NodeSelector) != 0 {
		logger.Info(ctx, "Applying pod node selecter using config %v", cfg.NodeSelector)
		pod.Spec.NodeSelector = cfg.NodeSelector
	}

	logger.Info(ctx, "applyNodeSelectorLabels Pod:  %v", pod)
	return pod
}

func applyPodTolerations(ctx context.Context, cfg *Config, pod *corev1.Pod) *corev1.Pod {
	if len(cfg.Tolerations) != 0 {
		logger.Info(ctx, "Applying pod tolerations using config %v", cfg.Tolerations)
		pod.Spec.Tolerations = cfg.Tolerations
	}

	logger.Info(ctx, "applyPodTolerations Pod:  %v", pod)
	return pod
}

func TerminateSubTasks(ctx context.Context, tCtx core.TaskExecutionContext, kubeClient core.KubeClient, config *Config,
	currentState *arrayCore.State) error {

	size := currentState.GetExecutionArraySize()
	errs := errorcollector.NewErrorMessageCollector()
	for childIdx := 0; childIdx < size; childIdx++ {
		task := Task{
			ChildIdx: childIdx,
			Config:   config,
			State:    currentState,
		}

		err := task.Abort(ctx, tCtx, kubeClient)
		if err != nil {
			errs.Collect(childIdx, err.Error())
		}
		err = task.Finalize(ctx, tCtx, kubeClient)
		if err != nil {
			errs.Collect(childIdx, err.Error())
		}
	}

	if errs.Length() > 0 {
		return fmt.Errorf(errs.Summary(config.MaxErrorStringLength))
	}

	return nil
}
