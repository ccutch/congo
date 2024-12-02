package monitoring

import (
	"github.com/ccutch/congo/pkg/congo"
	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemStatus struct {
	congo.Model
	UsageCPU    int
	UsageRAM    int
	StorageUsed int
	VolumeUsed  int
}

func GetSystemStatus(db *congo.Database) (*SystemStatus, error) {
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return nil, err
	}
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	diskStat, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}
	volumeStat, err := disk.Usage("/mnt/data")
	if err != nil {
		return nil, err
	}
	return &SystemStatus{
		Model:       db.NewModel(uuid.NewString()),
		UsageCPU:    int(cpuPercent[0] * 100),
		UsageRAM:    int(vmStat.UsedPercent * 100),
		StorageUsed: int(diskStat.UsedPercent * 100),
		VolumeUsed:  int(volumeStat.UsedPercent * 100),
	}, nil
}

func (s *SystemStatus) Save() error {
	return s.Query(`
		
		INSERT INTO system_status (id, cpu_usage, ram_usage, storage_used, volume_used)
		VALUES (?, ?, ?, ?, ?)
		RETURNING created_at, updated_at
	
	`, s.ID, s.UsageCPU, s.UsageRAM, s.StorageUsed, s.VolumeUsed).Scan(&s.CreatedAt, &s.UpdatedAt)
}
