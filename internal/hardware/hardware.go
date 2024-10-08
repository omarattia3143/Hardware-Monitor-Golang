package hardware

import (
	"fmt"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"runtime"
)

func GetSystemSection() (string, error) {
	runTimeOS := runtime.GOOS
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return "", err
	}
	hostStat, err := host.Info()
	if err != nil {
		return "", err
	}

	output := fmt.Sprintf("Hostname: %s\nTotal Memory: %.2f GB\nUsed Memory: %.2f GB\nOS: %s\n", hostStat.Hostname, bytesToGB(vmStat.Total), bytesToGB(vmStat.Used), runTimeOS)
	return output, nil
}
func GetCpuSection() (string, error) {
	cpuStat, err := cpu.Info()
	if err != nil {
		return "", err
	}

	output := fmt.Sprintf("CPU: %s\nCores: %d\n", cpuStat[0].ModelName, cpuStat[0].Cores)
	return output, nil
}

func GetDiskSection() (string, error) {
	diskStat, err := disk.Usage("/")
	if err != nil {
		return "", err
	}
	output := fmt.Sprintf("Total Disk Space: %.2f GB\nFree Disk Space: %.2f GB",
		bytesToGB(diskStat.Total), bytesToGB(diskStat.Free))
	return output, nil
}

func bytesToGB(bytes uint64) float64 {
	const bytesInGB = 1073741824 // 1 GB in bytes
	return float64(bytes) / bytesInGB
}
