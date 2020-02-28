// +build !ignore_autogenerated

/*
Copyright (c) Puppet, Inc.
*/

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Unstructured) DeepCopyInto(out *Unstructured) {
	clone := in.DeepCopy()
	*out = *clone
	return
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in UnstructuredObject) DeepCopyInto(out *UnstructuredObject) {
	{
		in := &in
		*out = make(UnstructuredObject, len(*in))
		for key, val := range *in {
			var outVal *Unstructured
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = (*in).DeepCopy()
			}
			(*out)[key] = outVal
		}
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UnstructuredObject.
func (in UnstructuredObject) DeepCopy() UnstructuredObject {
	if in == nil {
		return nil
	}
	out := new(UnstructuredObject)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Workflow) DeepCopyInto(out *Workflow) {
	*out = *in
	if in.Parameters != nil {
		in, out := &in.Parameters, &out.Parameters
		*out = make(WorkflowParameters, len(*in))
		for key, val := range *in {
			var outVal *Unstructured
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = (*in).DeepCopy()
			}
			(*out)[key] = outVal
		}
	}
	if in.Steps != nil {
		in, out := &in.Steps, &out.Steps
		*out = make([]*WorkflowStep, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(WorkflowStep)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Workflow.
func (in *Workflow) DeepCopy() *Workflow {
	if in == nil {
		return nil
	}
	out := new(Workflow)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in WorkflowParameters) DeepCopyInto(out *WorkflowParameters) {
	{
		in := &in
		*out = make(WorkflowParameters, len(*in))
		for key, val := range *in {
			var outVal *Unstructured
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = (*in).DeepCopy()
			}
			(*out)[key] = outVal
		}
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkflowParameters.
func (in WorkflowParameters) DeepCopy() WorkflowParameters {
	if in == nil {
		return nil
	}
	out := new(WorkflowParameters)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkflowRun) DeepCopyInto(out *WorkflowRun) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.State.DeepCopyInto(&out.State)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkflowRun.
func (in *WorkflowRun) DeepCopy() *WorkflowRun {
	if in == nil {
		return nil
	}
	out := new(WorkflowRun)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *WorkflowRun) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkflowRunList) DeepCopyInto(out *WorkflowRunList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]WorkflowRun, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkflowRunList.
func (in *WorkflowRunList) DeepCopy() *WorkflowRunList {
	if in == nil {
		return nil
	}
	out := new(WorkflowRunList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *WorkflowRunList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in WorkflowRunParameters) DeepCopyInto(out *WorkflowRunParameters) {
	{
		in := &in
		*out = make(WorkflowRunParameters, len(*in))
		for key, val := range *in {
			var outVal *Unstructured
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = (*in).DeepCopy()
			}
			(*out)[key] = outVal
		}
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkflowRunParameters.
func (in WorkflowRunParameters) DeepCopy() WorkflowRunParameters {
	if in == nil {
		return nil
	}
	out := new(WorkflowRunParameters)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkflowRunSpec) DeepCopyInto(out *WorkflowRunSpec) {
	*out = *in
	if in.Parameters != nil {
		in, out := &in.Parameters, &out.Parameters
		*out = make(WorkflowRunParameters, len(*in))
		for key, val := range *in {
			var outVal *Unstructured
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = (*in).DeepCopy()
			}
			(*out)[key] = outVal
		}
	}
	in.Workflow.DeepCopyInto(&out.Workflow)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkflowRunSpec.
func (in *WorkflowRunSpec) DeepCopy() *WorkflowRunSpec {
	if in == nil {
		return nil
	}
	out := new(WorkflowRunSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkflowRunState) DeepCopyInto(out *WorkflowRunState) {
	*out = *in
	if in.Workflow != nil {
		in, out := &in.Workflow, &out.Workflow
		*out = make(WorkflowState, len(*in))
		for key, val := range *in {
			var outVal *Unstructured
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = (*in).DeepCopy()
			}
			(*out)[key] = outVal
		}
	}
	if in.Steps != nil {
		in, out := &in.Steps, &out.Steps
		*out = make(map[string]WorkflowState, len(*in))
		for key, val := range *in {
			var outVal map[string]*Unstructured
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = make(WorkflowState, len(*in))
				for key, val := range *in {
					var outVal *Unstructured
					if val == nil {
						(*out)[key] = nil
					} else {
						in, out := &val, &outVal
						*out = (*in).DeepCopy()
					}
					(*out)[key] = outVal
				}
			}
			(*out)[key] = outVal
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkflowRunState.
func (in *WorkflowRunState) DeepCopy() *WorkflowRunState {
	if in == nil {
		return nil
	}
	out := new(WorkflowRunState)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkflowRunStatus) DeepCopyInto(out *WorkflowRunStatus) {
	*out = *in
	if in.StartTime != nil {
		in, out := &in.StartTime, &out.StartTime
		*out = (*in).DeepCopy()
	}
	if in.CompletionTime != nil {
		in, out := &in.CompletionTime, &out.CompletionTime
		*out = (*in).DeepCopy()
	}
	if in.Steps != nil {
		in, out := &in.Steps, &out.Steps
		*out = make(map[string]WorkflowRunStatusSummary, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make(map[string]WorkflowRunStatusSummary, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkflowRunStatus.
func (in *WorkflowRunStatus) DeepCopy() *WorkflowRunStatus {
	if in == nil {
		return nil
	}
	out := new(WorkflowRunStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkflowRunStatusSummary) DeepCopyInto(out *WorkflowRunStatusSummary) {
	*out = *in
	if in.StartTime != nil {
		in, out := &in.StartTime, &out.StartTime
		*out = (*in).DeepCopy()
	}
	if in.CompletionTime != nil {
		in, out := &in.CompletionTime, &out.CompletionTime
		*out = (*in).DeepCopy()
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkflowRunStatusSummary.
func (in *WorkflowRunStatusSummary) DeepCopy() *WorkflowRunStatusSummary {
	if in == nil {
		return nil
	}
	out := new(WorkflowRunStatusSummary)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in WorkflowState) DeepCopyInto(out *WorkflowState) {
	{
		in := &in
		*out = make(WorkflowState, len(*in))
		for key, val := range *in {
			var outVal *Unstructured
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = (*in).DeepCopy()
			}
			(*out)[key] = outVal
		}
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkflowState.
func (in WorkflowState) DeepCopy() WorkflowState {
	if in == nil {
		return nil
	}
	out := new(WorkflowState)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkflowStep) DeepCopyInto(out *WorkflowStep) {
	*out = *in
	if in.Spec != nil {
		in, out := &in.Spec, &out.Spec
		*out = make(UnstructuredObject, len(*in))
		for key, val := range *in {
			var outVal *Unstructured
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = (*in).DeepCopy()
			}
			(*out)[key] = outVal
		}
	}
	if in.Input != nil {
		in, out := &in.Input, &out.Input
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Args != nil {
		in, out := &in.Args, &out.Args
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.When != nil {
		in, out := &in.When, &out.When
		*out = (*in).DeepCopy()
	}
	if in.DependsOn != nil {
		in, out := &in.DependsOn, &out.DependsOn
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkflowStep.
func (in *WorkflowStep) DeepCopy() *WorkflowStep {
	if in == nil {
		return nil
	}
	out := new(WorkflowStep)
	in.DeepCopyInto(out)
	return out
}
