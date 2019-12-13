package main

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type VmInstance struct {
	res      apiv1.ResourceRequirements
	replicas int32
}

type VmInstanceResourceCount struct {
	cpu int64
	mem int64
}

const (
	MemorySizeB  = 1
	MemorySizeKi = 1024 * MemorySizeB
	MemorySizeMi = 1024 * MemorySizeKi

	CpuSizeFraction = 1
	CpuSizeInteger  = 1 * 1000

	DefaultInstanceReplicasCount = 0
	TestInstanceReplicasCount    = 1

	RuntimeMulity = 50
	NotBest       = float64(99999999999)

	TrainDataFilePath = "data/train"
)

var (
	vmConfigList = [...]VmInstanceResourceCount{
		{cpu: 125, mem: 128},
		{cpu: 250, mem: 256},
		{cpu: 375, mem: 384},
		//{cpu: 500, mem: 512},
		//{cpu: 625, mem: 640},
		//{cpu: 750, mem: 768},
		//{cpu: 875, mem: 896},
		//{cpu: 1000, mem: 1024},
		//{cpu: 1125, mem: 1152},
		//{cpu: 1250, mem: 1280},
		//{cpu: 1375, mem: 1408},
		//{cpu: 1500, mem: 1536},
		//{cpu: 1625, mem: 1664},
		//{cpu: 1750, mem: 1792},
		//{cpu: 1875, mem: 1920},
		//{cpu: 2000, mem: 2048},
	}

	vmInstanceDefault = VmInstance{
		res: apiv1.ResourceRequirements{
			Limits: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(vmConfigList[0].cpu*CpuSizeFraction, resource.BinarySI),
				apiv1.ResourceMemory: *resource.NewQuantity(vmConfigList[0].mem*MemorySizeMi, resource.BinarySI),
			},
			Requests: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(vmConfigList[0].cpu*CpuSizeFraction, resource.BinarySI),
				apiv1.ResourceMemory: *resource.NewQuantity(vmConfigList[0].mem*MemorySizeMi, resource.BinarySI),
			},
		},
		replicas: DefaultInstanceReplicasCount,
	}

	vmList = func() [len(vmConfigList)]VmInstance {
		vmList := [len(vmConfigList)]VmInstance{}
		for index, vmcl := range vmConfigList {
			vmList[index] = VmInstance{
				res: apiv1.ResourceRequirements{
					Limits: apiv1.ResourceList{
						apiv1.ResourceCPU:    *resource.NewMilliQuantity(vmcl.cpu*CpuSizeFraction, resource.BinarySI),
						apiv1.ResourceMemory: *resource.NewQuantity(vmcl.mem*MemorySizeMi, resource.BinarySI),
					},
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    *resource.NewMilliQuantity(vmcl.cpu*CpuSizeFraction, resource.BinarySI),
						apiv1.ResourceMemory: *resource.NewQuantity(vmcl.mem*MemorySizeMi, resource.BinarySI),
					},
				},
				replicas: TestInstanceReplicasCount,
			}
		}
		return vmList
	}

	vmCost = func() [len(vmConfigList)]int {
		vmCost := [len(vmConfigList)]int{}
		for index := range vmConfigList {
			vmCost[index] = index + 1
		}
		return vmCost
	}
)

var (
	concurrency = [...]int{
		//1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024,
		1, 2, 4, 8, 16, 32, 64,
	}
	CostPerformanceTable = [len(vmConfigList)][len(concurrency)]float64{}
)
