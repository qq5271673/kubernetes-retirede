// +build !ignore_autogenerated

/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This file was autogenerated by deepcopy-gen. Do not edit it manually!

package rbac

import (
	api "k8s.io/kubernetes/pkg/api"
	conversion "k8s.io/kubernetes/pkg/conversion"
	runtime "k8s.io/kubernetes/pkg/runtime"
	reflect "reflect"
)

func init() {
	if err := api.Scheme.AddGeneratedDeepCopyFuncs(
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_rbac_ClusterRole, InType: reflect.TypeOf(func() *ClusterRole { var x *ClusterRole; return x }())},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_rbac_ClusterRoleBinding, InType: reflect.TypeOf(func() *ClusterRoleBinding { var x *ClusterRoleBinding; return x }())},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_rbac_ClusterRoleBindingList, InType: reflect.TypeOf(func() *ClusterRoleBindingList { var x *ClusterRoleBindingList; return x }())},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_rbac_ClusterRoleList, InType: reflect.TypeOf(func() *ClusterRoleList { var x *ClusterRoleList; return x }())},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_rbac_PolicyRule, InType: reflect.TypeOf(func() *PolicyRule { var x *PolicyRule; return x }())},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_rbac_Role, InType: reflect.TypeOf(func() *Role { var x *Role; return x }())},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_rbac_RoleBinding, InType: reflect.TypeOf(func() *RoleBinding { var x *RoleBinding; return x }())},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_rbac_RoleBindingList, InType: reflect.TypeOf(func() *RoleBindingList { var x *RoleBindingList; return x }())},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_rbac_RoleList, InType: reflect.TypeOf(func() *RoleList { var x *RoleList; return x }())},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_rbac_Subject, InType: reflect.TypeOf(func() *Subject { var x *Subject; return x }())},
	); err != nil {
		// if one of the deep copy functions is malformed, detect it immediately.
		panic(err)
	}
}

func DeepCopy_rbac_ClusterRole(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*ClusterRole)
		out := out.(*ClusterRole)
		out.TypeMeta = in.TypeMeta
		if err := api.DeepCopy_api_ObjectMeta(&in.ObjectMeta, &out.ObjectMeta, c); err != nil {
			return err
		}
		if in.Rules != nil {
			in, out := &in.Rules, &out.Rules
			*out = make([]PolicyRule, len(*in))
			for i := range *in {
				if err := DeepCopy_rbac_PolicyRule(&(*in)[i], &(*out)[i], c); err != nil {
					return err
				}
			}
		} else {
			out.Rules = nil
		}
		return nil
	}
}

func DeepCopy_rbac_ClusterRoleBinding(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*ClusterRoleBinding)
		out := out.(*ClusterRoleBinding)
		out.TypeMeta = in.TypeMeta
		if err := api.DeepCopy_api_ObjectMeta(&in.ObjectMeta, &out.ObjectMeta, c); err != nil {
			return err
		}
		if in.Subjects != nil {
			in, out := &in.Subjects, &out.Subjects
			*out = make([]Subject, len(*in))
			for i := range *in {
				(*out)[i] = (*in)[i]
			}
		} else {
			out.Subjects = nil
		}
		out.RoleRef = in.RoleRef
		return nil
	}
}

func DeepCopy_rbac_ClusterRoleBindingList(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*ClusterRoleBindingList)
		out := out.(*ClusterRoleBindingList)
		out.TypeMeta = in.TypeMeta
		out.ListMeta = in.ListMeta
		if in.Items != nil {
			in, out := &in.Items, &out.Items
			*out = make([]ClusterRoleBinding, len(*in))
			for i := range *in {
				if err := DeepCopy_rbac_ClusterRoleBinding(&(*in)[i], &(*out)[i], c); err != nil {
					return err
				}
			}
		} else {
			out.Items = nil
		}
		return nil
	}
}

func DeepCopy_rbac_ClusterRoleList(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*ClusterRoleList)
		out := out.(*ClusterRoleList)
		out.TypeMeta = in.TypeMeta
		out.ListMeta = in.ListMeta
		if in.Items != nil {
			in, out := &in.Items, &out.Items
			*out = make([]ClusterRole, len(*in))
			for i := range *in {
				if err := DeepCopy_rbac_ClusterRole(&(*in)[i], &(*out)[i], c); err != nil {
					return err
				}
			}
		} else {
			out.Items = nil
		}
		return nil
	}
}

func DeepCopy_rbac_PolicyRule(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*PolicyRule)
		out := out.(*PolicyRule)
		if in.Verbs != nil {
			in, out := &in.Verbs, &out.Verbs
			*out = make([]string, len(*in))
			copy(*out, *in)
		} else {
			out.Verbs = nil
		}
		if in.AttributeRestrictions == nil {
			out.AttributeRestrictions = nil
		} else if newVal, err := c.DeepCopy(&in.AttributeRestrictions); err != nil {
			return err
		} else {
			out.AttributeRestrictions = *newVal.(*runtime.Object)
		}
		if in.APIGroups != nil {
			in, out := &in.APIGroups, &out.APIGroups
			*out = make([]string, len(*in))
			copy(*out, *in)
		} else {
			out.APIGroups = nil
		}
		if in.Resources != nil {
			in, out := &in.Resources, &out.Resources
			*out = make([]string, len(*in))
			copy(*out, *in)
		} else {
			out.Resources = nil
		}
		if in.ResourceNames != nil {
			in, out := &in.ResourceNames, &out.ResourceNames
			*out = make([]string, len(*in))
			copy(*out, *in)
		} else {
			out.ResourceNames = nil
		}
		if in.NonResourceURLs != nil {
			in, out := &in.NonResourceURLs, &out.NonResourceURLs
			*out = make([]string, len(*in))
			copy(*out, *in)
		} else {
			out.NonResourceURLs = nil
		}
		return nil
	}
}

func DeepCopy_rbac_Role(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*Role)
		out := out.(*Role)
		out.TypeMeta = in.TypeMeta
		if err := api.DeepCopy_api_ObjectMeta(&in.ObjectMeta, &out.ObjectMeta, c); err != nil {
			return err
		}
		if in.Rules != nil {
			in, out := &in.Rules, &out.Rules
			*out = make([]PolicyRule, len(*in))
			for i := range *in {
				if err := DeepCopy_rbac_PolicyRule(&(*in)[i], &(*out)[i], c); err != nil {
					return err
				}
			}
		} else {
			out.Rules = nil
		}
		return nil
	}
}

func DeepCopy_rbac_RoleBinding(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*RoleBinding)
		out := out.(*RoleBinding)
		out.TypeMeta = in.TypeMeta
		if err := api.DeepCopy_api_ObjectMeta(&in.ObjectMeta, &out.ObjectMeta, c); err != nil {
			return err
		}
		if in.Subjects != nil {
			in, out := &in.Subjects, &out.Subjects
			*out = make([]Subject, len(*in))
			for i := range *in {
				(*out)[i] = (*in)[i]
			}
		} else {
			out.Subjects = nil
		}
		out.RoleRef = in.RoleRef
		return nil
	}
}

func DeepCopy_rbac_RoleBindingList(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*RoleBindingList)
		out := out.(*RoleBindingList)
		out.TypeMeta = in.TypeMeta
		out.ListMeta = in.ListMeta
		if in.Items != nil {
			in, out := &in.Items, &out.Items
			*out = make([]RoleBinding, len(*in))
			for i := range *in {
				if err := DeepCopy_rbac_RoleBinding(&(*in)[i], &(*out)[i], c); err != nil {
					return err
				}
			}
		} else {
			out.Items = nil
		}
		return nil
	}
}

func DeepCopy_rbac_RoleList(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*RoleList)
		out := out.(*RoleList)
		out.TypeMeta = in.TypeMeta
		out.ListMeta = in.ListMeta
		if in.Items != nil {
			in, out := &in.Items, &out.Items
			*out = make([]Role, len(*in))
			for i := range *in {
				if err := DeepCopy_rbac_Role(&(*in)[i], &(*out)[i], c); err != nil {
					return err
				}
			}
		} else {
			out.Items = nil
		}
		return nil
	}
}

func DeepCopy_rbac_Subject(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*Subject)
		out := out.(*Subject)
		out.Kind = in.Kind
		out.APIVersion = in.APIVersion
		out.Name = in.Name
		out.Namespace = in.Namespace
		return nil
	}
}
