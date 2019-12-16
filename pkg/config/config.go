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

	//采用二分法的策略任何等级的并发都不会超过100次
	LatencyMaxHeyCount = 100
)

var (
	VmConfigList = [...]VmInstanceResourceCount{
		{Cpu: 125, Mem: 128},
		{Cpu: 250, Mem: 256},
		{Cpu: 375, Mem: 384},
		{Cpu: 500, Mem: 512},
		{Cpu: 625, Mem: 640},
		{Cpu: 750, Mem: 768},
		{Cpu: 875, Mem: 896},
		{Cpu: 1000, Mem: 1024},
		//{Cpu: 1125, Mem: 1152},
		//{Cpu: 1250, Mem: 1280},
		//{Cpu: 1375, Mem: 1408},
		//{Cpu: 1500, Mem: 1536},
		//{Cpu: 1625, Mem: 1664},
		//{Cpu: 1750, Mem: 1792},
		//{Cpu: 1875, Mem: 1920},
		//{Cpu: 2000, Mem: 2048},
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

var (
	TargetLatency = [...]float64{
		0.010, 0.011, 0.012, 0.013, 0.014, 0.015, 0.016, 0.017, 0.018, 0.019, 0.020, 0.021, 0.022, 0.023, 0.024, 0.025, 0.026, 0.027, 0.028, 0.029, 0.030,
	}
)
