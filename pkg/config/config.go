package config

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type VmInstance struct {
	Res      apiv1.ResourceRequirements
	Replicas int32
}

type VmInstanceResourceCount struct {
	Cpu int64
	Mem int64
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
	VmConfigList = [...]VmInstanceResourceCount{
		{Cpu: 125, Mem: 128},
		{Cpu: 250, Mem: 256},
		//{cpu: 375, mem: 384},
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

	VmInstanceDefault = VmInstance{
		Res: apiv1.ResourceRequirements{
			Limits: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(VmConfigList[0].Cpu*CpuSizeFraction, resource.BinarySI),
				apiv1.ResourceMemory: *resource.NewQuantity(VmConfigList[0].Mem*MemorySizeMi, resource.BinarySI),
			},
			Requests: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(VmConfigList[0].Cpu*CpuSizeFraction, resource.BinarySI),
				apiv1.ResourceMemory: *resource.NewQuantity(VmConfigList[0].Mem*MemorySizeMi, resource.BinarySI),
			},
		},
		Replicas: DefaultInstanceReplicasCount,
	}

	VmList = func() [len(VmConfigList)]VmInstance {
		vmList := [len(VmConfigList)]VmInstance{}
		for index, vmcl := range VmConfigList {
			vmList[index] = VmInstance{
				Res: apiv1.ResourceRequirements{
					Limits: apiv1.ResourceList{
						apiv1.ResourceCPU:    *resource.NewMilliQuantity(vmcl.Cpu*CpuSizeFraction, resource.BinarySI),
						apiv1.ResourceMemory: *resource.NewQuantity(vmcl.Mem*MemorySizeMi, resource.BinarySI),
					},
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    *resource.NewMilliQuantity(vmcl.Cpu*CpuSizeFraction, resource.BinarySI),
						apiv1.ResourceMemory: *resource.NewQuantity(vmcl.Mem*MemorySizeMi, resource.BinarySI),
					},
				},
				Replicas: TestInstanceReplicasCount,
			}
		}
		return vmList
	}

	VmCost = func() [len(VmConfigList)]int {
		vmCost := [len(VmConfigList)]int{}
		for index := range VmConfigList {
			vmCost[index] = index + 1
		}
		return vmCost
	}
)

var (
	Concurrency = [...]int{
		//1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024,
		1, 2, 4, 8, 16, 32, 64, 128, 256,
	}
	CostPerformanceTable = [len(VmConfigList)][len(Concurrency)]float64{}
)
